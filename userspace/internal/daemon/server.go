package daemon

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"sync/atomic"

	"github.com/toastsandwich/guardian/internal/guardian"
)

const (
	Path = "/tmp/guardian.sock"
)

type Command byte

const (
	Ping   Command = 0x00
	Attach Command = 0x01
	Detach Command = 0x02
	Show   Command = 0x03
	Stop   Command = 0xFF
)

type Status byte

const (
	StatusOK       Status = 0x00
	StatusAttached Status = 0x01
	StatusClosed   Status = 0x0F
)

type Request struct {
	Command   Command `json:"command"`
	IfaceName string  `json:"iface_name"`
}

type Response struct {
	Status Status       `json:"status"`
	Ips    []ShowResult `json:"ips"`
	Error  error        `json:"error,omitempty"`
}

type Server struct {
	log *slog.Logger

	guardian atomic.Pointer[guardian.Guardian]

	isClosed atomic.Bool
	doneCh   chan struct{}
}

func NewServer() *Server {
	th := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	s := &Server{
		log:    slog.New(th),
		doneCh: make(chan struct{}),
	}

	return s
}

func (s *Server) ListenAndServe() error {
	_ = os.Remove(Path)

	ln, err := net.Listen("unix", Path)
	if err != nil {
		return err
	}
	defer ln.Close()

	ul, ok := ln.(*net.UnixListener)
	if ok {
		ul.SetUnlinkOnClose(true)
	}
	slog.Info("listening for requests")
	go func() {
		for !s.isClosed.Load() {
			conn, err := ln.Accept()
			if err != nil {
				s.log.Error("failed to accept incoming connection", "ERR", err)
			}
			s.handleConn(conn)
		}
	}()
	<-s.doneCh
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
		s.log.Error("failed to read from connection", "ERR", err)
		return
	}
	r := Request{}
	err = json.Unmarshal(b[:n], &r)
	if err != nil {
		s.log.Error("failed to unmarshal incoming request", "ERR", err)
		return
	}
	switch r.Command {
	case Ping:
		res, _ := HandlePing(nil)
		b, err := json.Marshal(&res)
		if err != nil {
			s.log.Error("failed to marshal ping response", "ERR", err)
		}
		_, err = conn.Write(b)
		if err != nil {
			s.log.Error("failed to write ping response to client", "ERR", err)
		}
	case Attach:
		resp, err := s.HandleAttach(&r)
		if err != nil {
			s.log.Error("failed to handle attach request", "ERR", err)
		}
		b, err := json.Marshal(resp)
		if err != nil {
			s.log.Error("failed to marshal response", "ERR", err)
		}

		_, err = conn.Write(b)
		if err != nil {
			s.log.Error("failed to write attach response to client", "ERR", err)
		}
		return

	case Show:
		resp := s.HandleShow()
		b, err := json.Marshal(resp)
		if err != nil {
			s.log.Error("failed to marshal response", "ERR", err)
		}

		_, err = conn.Write(b)
		if err != nil {
			s.log.Error("failed to write attach response to client", "ERR", err)
		}
		return

	case Stop:
		if g := s.guardian.Load(); g != nil {
			if err := g.Close(); err != nil {
				s.log.Error("failed to close guardian", "ERR", err)
			}
			s.guardian.Store(nil)
		}
		s.Close()
		return
	}
}

func (s *Server) Close() {
	close(s.doneCh)

	s.isClosed.Store(true)
}
