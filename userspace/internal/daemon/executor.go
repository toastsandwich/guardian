package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/toastsandwich/guardian/internal/guardian"
)

type Command byte

const (
	StartCmd Command = iota
	AttachCmd
	DettachCmd
	AllowCmd
	DenyCmd
	ListCmd
	StopCmd
)

func (c Command) String() string {
	switch c {
	case StartCmd:
		return "start"
	case AttachCmd:
		return "attach"
	case AllowCmd:
		return "allow"
	case DenyCmd:
		return "deny"
	case DettachCmd:
		return "detach"
	case ListCmd:
		return "list"
	case StopCmd:
		return "stop"
	default:
		return "unknown"
	}
}

type Code byte

const (
	CodeOK = iota
	CodeNotOK
	CodeInternal
	CodeDenied
)

type Request struct {
	Version byte    `json:"version"`
	Command Command `json:"command"`
	Body    []byte  `json:"body"`
}

type Response struct {
	Code byte   `json:"code"`
	Body []byte `json:"body,omitempty"`
}

type ExecFunc func(Request) (Response, error)

type Executor struct {
	execM  map[Command]ExecFunc
	g      *guardian.Guardian
	logger *slog.Logger
}

func NewEmpty() *Executor {
	return &Executor{logger: slog.Default().With("component", "executor")}
}

func codeString(c byte) string {
	switch c {
	case CodeOK:
		return "ok"
	case CodeNotOK:
		return "not_ok"
	case CodeInternal:
		return "internal"
	case CodeDenied:
		return "denied"
	default:
		return fmt.Sprintf("unknown(%d)", c)
	}
}

func (e *Executor) Init() {
	em := map[Command]ExecFunc{}
	em[AttachCmd] = e.ExecAttach
	em[DettachCmd] = e.ExecDetach
	em[ListCmd] = e.ExecList
	em[AllowCmd] = e.ExecAllow
	em[DenyCmd] = e.ExecDeny
	em[StopCmd] = e.ExecStop
	e.execM = em
}

var ErrGuardianIsNil = errors.New("guardian is nil")

func (e *Executor) guardianNilCheck() error {
	if e.g == nil {
		return ErrGuardianIsNil
	}
	return nil
}

var ErrExecutionMissing = errors.New("execution missing")

func (e *Executor) Do(c Command) (ExecFunc, error) {
	f, ok := e.execM[c]
	if !ok {
		return nil, ErrExecutionMissing
	}
	return f, nil
}

type AttachOptions struct {
	IfaceName string `json:"iface_name"`
}

func (e *Executor) ExecAttach(r Request) (Response, error) {
	ao := AttachOptions{}
	err := json.Unmarshal(r.Body, &ao)
	if err != nil {
		e.logger.Error("attach: invalid request body", "err", err)
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Info("attach requested", "iface", ao.IfaceName)
	e.g = guardian.New(ao.IfaceName)
	if err := e.g.Attach(); err != nil {
		e.logger.Error("attach failed", "iface", ao.IfaceName, "err", err)
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Info("attached", "iface", ao.IfaceName)
	return Response{Code: CodeOK}, nil
}

func (e *Executor) ExecDetach(r Request) (Response, error) {
	if e.g == nil {
		e.logger.Error("detach failed: not attached")
		return Response{Code: CodeNotOK}, errors.New("not attached")
	}
	iface := e.g.IfaceName
	e.logger.Info("detach requested", "iface", iface)
	if err := e.g.Close(); err != nil {
		e.logger.Error("detach failed", "iface", iface, "err", err)
		return Response{Code: CodeNotOK}, err
	}
	e.g = nil
	e.logger.Info("detached", "iface", iface)
	return Response{Code: CodeOK}, nil
}

type ListOptions struct {
	Limit int
}

type ListResp struct {
	IPs []IPres `json:"ips"`
}

type IPres struct {
	IP    string `json:"ip"`
	Allow bool   `json:"allow"`
}

func (e *Executor) ExecList(r Request) (Response, error) {
	if e.g == nil {
		e.logger.Error("list failed: not attached")
		return Response{Code: CodeNotOK}, errors.New("not attached")
	}

	e.logger.Info("list requested", "iface", e.g.IfaceName)
	sr := ListResp{IPs: make([]IPres, 0)}

	allowed, denied := 0, 0
	for r := range e.g.List() {
		sr.IPs = append(sr.IPs, IPres{
			IP:    r.IP,
			Allow: r.Allowed,
		})
		if r.Allowed {
			allowed++
		} else {
			denied++
		}
	}

	b, err := json.Marshal(&sr)
	if err != nil {
		e.logger.Error("list: failed to marshal response", "err", err, "count", len(sr.IPs))
		return Response{Code: CodeNotOK}, err
	}
	e.logger.Info("list completed", "iface", e.g.IfaceName, "count", len(sr.IPs), "allowed", allowed, "denied", denied)
	return Response{Body: b, Code: CodeOK}, nil
}

func (e *Executor) ExecStop(r Request) (Response, error) {
	e.logger.Info("stop requested", "attached", e.g != nil)
	if e.g != nil {
		iface := e.g.IfaceName
		e.logger.Info("detaching before stop", "iface", iface)
		if err := e.g.Close(); err != nil {
			e.logger.Error("stop: detach failed", "iface", iface, "err", err)
			return Response{Code: CodeNotOK}, err
		}
		e.g = nil
		e.logger.Info("detached before stop", "iface", iface)
	}
	e.logger.Info("stop completed")
	return Response{Code: CodeOK}, nil
}

type DenyOptions struct {
	IP string `json:"ip"`
}

func (e *Executor) ExecDeny(r Request) (Response, error) {
	do := DenyOptions{}
	err := json.Unmarshal(r.Body, &do)
	if err != nil {
		e.logger.Error("deny: invalid request body", "err", err)
		return Response{}, err
	}
	if e.g == nil {
		e.logger.Error("deny failed: not attached", "ip", do.IP)
		return Response{Code: CodeNotOK}, errors.New("not attached")
	}

	e.logger.Info("deny requested", "ip", do.IP, "iface", e.g.IfaceName)
	err = e.g.Deny(do.IP)
	if err != nil {
		e.logger.Error("deny failed", "ip", do.IP, "iface", e.g.IfaceName, "err", err)
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Info("denied", "ip", do.IP, "iface", e.g.IfaceName)
	return Response{Code: CodeOK}, nil
}

type AllowOptions struct {
	IP string `json:"ip"`
}

func (e *Executor) ExecAllow(r Request) (Response, error) {
	ao := AllowOptions{}
	err := json.Unmarshal(r.Body, &ao)
	if err != nil {
		e.logger.Error("allow: invalid request body", "err", err)
		return Response{Code: CodeNotOK}, err
	}

	if err := e.guardianNilCheck(); err != nil {
		e.logger.Error("allow failed: not attached", "ip", ao.IP)
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Info("allow requested", "ip", ao.IP, "iface", e.g.IfaceName)
	err = e.g.Allow(ao.IP)
	if err != nil {
		e.logger.Error("allow failed", "ip", ao.IP, "iface", e.g.IfaceName, "err", err)
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Info("allowed", "ip", ao.IP, "iface", e.g.IfaceName)
	return Response{Code: CodeOK}, nil
}
