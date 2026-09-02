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
	GetModeCmd
	SetModeCmd
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
	case GetModeCmd:
		return "get_mode"
	case SetModeCmd:
		return "set_mode"
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
	em[GetModeCmd] = e.ExecGetMode
	em[SetModeCmd] = e.ExecSetMode
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
	Mode      string `json:"mode"`
}

func (e *Executor) ExecAttach(r Request) (Response, error) {
	ao := AttachOptions{}
	err := json.Unmarshal(r.Body, &ao)
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Debug("attach requested", "iface", ao.IfaceName, "mode", ao.Mode)
	e.g = guardian.New(ao.IfaceName)
	mode, err := guardian.DetermineMode(ao.Mode)
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}
	if err := e.g.Attach(mode); err != nil {
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Debug("attached", "iface", ao.IfaceName, "mode", ao.Mode)
	return Response{Code: CodeOK}, nil
}

func (e *Executor) ExecDetach(r Request) (Response, error) {
	if e.g == nil {
		return Response{Code: CodeNotOK}, errors.New("not attached")
	}
	iface := e.g.IfaceName
	e.logger.Debug("detach requested", "iface", iface)
	if err := e.g.Close(); err != nil {
		return Response{Code: CodeNotOK}, err
	}
	e.g = nil
	e.logger.Debug("detached", "iface", iface)
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
		return Response{Code: CodeNotOK}, errors.New("not attached")
	}

	e.logger.Debug("list requested", "iface", e.g.IfaceName)
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
		return Response{Code: CodeNotOK}, err
	}
	e.logger.Debug("list completed", "iface", e.g.IfaceName, "count", len(sr.IPs), "allowed", allowed, "denied", denied)
	return Response{Body: b, Code: CodeOK}, nil
}

func (e *Executor) ExecStop(r Request) (Response, error) {
	e.logger.Debug("stop requested", "attached", e.g != nil)
	if e.g != nil {
		iface := e.g.IfaceName
		e.logger.Debug("detaching before stop", "iface", iface)
		if err := e.g.Close(); err != nil {
			return Response{Code: CodeNotOK}, err
		}
		e.g = nil
		e.logger.Debug("detached before stop", "iface", iface)
	}
	e.logger.Debug("stop completed")
	return Response{Code: CodeOK}, nil
}

type DenyOptions struct {
	IP string `json:"ip"`
}

func (e *Executor) ExecDeny(r Request) (Response, error) {
	do := DenyOptions{}
	err := json.Unmarshal(r.Body, &do)
	if err != nil {
		return Response{}, err
	}
	if e.g == nil {
		return Response{Code: CodeNotOK}, errors.New("not attached")
	}

	e.logger.Debug("deny requested", "ip", do.IP, "iface", e.g.IfaceName)
	err = e.g.Deny(do.IP)
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Debug("denied", "ip", do.IP, "iface", e.g.IfaceName)
	return Response{Code: CodeOK}, nil
}

type AllowOptions struct {
	IP string `json:"ip"`
}

func (e *Executor) ExecAllow(r Request) (Response, error) {
	ao := AllowOptions{}
	err := json.Unmarshal(r.Body, &ao)
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}

	if err := e.guardianNilCheck(); err != nil {
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Debug("allow requested", "ip", ao.IP, "iface", e.g.IfaceName)
	err = e.g.Allow(ao.IP)
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Debug("allowed", "ip", ao.IP, "iface", e.g.IfaceName)
	return Response{Code: CodeOK}, nil
}

type ModeResp struct {
	Mode string `json:"mode"`
}

func (e *Executor) ExecGetMode(r Request) (Response, error) {
	if err := e.guardianNilCheck(); err != nil {
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Debug("get mode requested", "iface", e.g.IfaceName)
	mode, err := e.g.GetMode()
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}

	b, err := json.Marshal(&ModeResp{Mode: mode})
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}
	e.logger.Debug("got mode", "iface", e.g.IfaceName, "mode", mode)
	return Response{Body: b, Code: CodeOK}, nil
}

type SetModeOptions struct {
	Mode string `json:"mode"`
}

func (e *Executor) ExecSetMode(r Request) (Response, error) {
	so := SetModeOptions{}
	err := json.Unmarshal(r.Body, &so)
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}

	if err := e.guardianNilCheck(); err != nil {
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Debug("set mode requested", "iface", e.g.IfaceName, "mode", so.Mode)
	mode, err := guardian.DetermineMode(so.Mode)
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}
	if err := e.g.SetMode(mode); err != nil {
		return Response{Code: CodeNotOK}, err
	}

	e.logger.Debug("set mode", "iface", e.g.IfaceName, "mode", so.Mode)
	return Response{Code: CodeOK}, nil
}
