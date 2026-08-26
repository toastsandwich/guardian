package daemon

import (
	"encoding/json"
	"errors"

	"github.com/toastsandwich/guardian/internal/guardian"
)

type Command byte

const (
	StartCmd Command = iota
	AttachCmd
	DettachCmd
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
	execM map[Command]ExecFunc
	g     *guardian.Guardian
}

func NewEmpty() *Executor {
	return &Executor{}
}

func (e *Executor) Init() {
	em := map[Command]ExecFunc{}
	em[AttachCmd] = e.ExecAttach
	em[DettachCmd] = e.ExecDetach
	em[ListCmd] = e.ExecList
	em[StopCmd] = e.ExecStop
	e.execM = em
}

var ErrExecutionMissing = errors.New("exection missing")

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
		return Response{Code: CodeNotOK}, err
	}

	e.g = guardian.New(ao.IfaceName)
	if err := e.g.Attach(); err != nil {
		return Response{Code: CodeNotOK}, err
	}

	return Response{Code: CodeOK}, nil
}

func (e *Executor) ExecDetach(r Request) (Response, error) {
	if e.g == nil {
		return Response{Code: CodeNotOK}, errors.New("not attached")
	}
	if err := e.g.Close(); err != nil {
		return Response{Code: CodeNotOK}, err
	}
	e.g = nil
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
	sr := ListResp{IPs: make([]IPres, 0)}

	for r := range e.g.List() {
		sr.IPs = append(sr.IPs, IPres{
			IP:    r.IP,
			Allow: r.Allowed,
		})
	}

	b, err := json.Marshal(&sr)
	if err != nil {
		return Response{Code: CodeNotOK}, err
	}
	return Response{Body: b, Code: CodeOK}, nil
}

func (e *Executor) ExecStop(r Request) (Response, error) {
	if e.g != nil {
		if err := e.g.Close(); err != nil {
			return Response{Code: CodeNotOK}, err
		}
		e.g = nil
	}
	return Response{Code: CodeOK}, nil
}
