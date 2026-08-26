package daemon

import (
	"fmt"
	"log/slog"
	"net"
	"os"
)

const path = "/tmp/guardian.sock"

type Server struct {
	ln   net.Listener
	exec *Executor

	logger *slog.Logger

	isStarted bool
}

func NewServer() *Server {
	ex := NewEmpty()
	ex.Init()
	return &Server{exec: ex, logger: slog.Default()}
}

func (s *Server) Start() error {
	if s.isStarted {
		return fmt.Errorf("guardian already started")
	}
	_ = os.Remove(path) // precaution

	s.logger.Info("starting listener ...")
	ln, err := net.Listen("unix", path)
	if err != nil {
		return err
	}
	if ul, ok := ln.(*net.UnixListener); ok {
		ul.SetUnlinkOnClose(true)
	}

	s.isStarted = true
	s.ln = ln
	defer s.Close()

	s.logger.Info("listener ready to accept connections.")
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if !s.isStarted {
				return nil
			}
			s.logger.Error("failed to accept connection", "ERR", err)
			continue
		}
		if err := s.handleConn(conn); err != nil {
			s.logger.Error("failed to handle connection", "ERR", err)
		}
		if !s.isStarted {
			s.logger.Info("guardian stopped")
			return nil
		}
	}
}

func (s *Server) handleConn(c net.Conn) error {
	conn := NewConn(c)
	defer conn.Close()

	req := Request{}
	err := conn.Recieve(&req)
	if err != nil {
		s.logger.Error("failed to receive request", "ERR", err)
		return err
	}

	execFn, err := s.exec.Do(req.Command)
	if err != nil {
		s.logger.Error("failed to find command handler", "ERR", err, "command", req.Command.String())
		resp := Response{Code: CodeInternal}
		if sendErr := conn.Send(&resp); sendErr != nil {
			s.logger.Error("failed to send response", "ERR", sendErr)
			return sendErr
		}
		return err
	}

	resp, err := execFn(req)
	if err != nil {
		s.logger.Error("failed to execute command", "ERR", err, "command", req.Command.String())
	}
	if sendErr := conn.Send(&resp); sendErr != nil {
		s.logger.Error("failed to send response", "ERR", sendErr)
		return sendErr
	}

	if req.Command == StopCmd {
		s.isStarted = false
	}
	return err
}

func (s *Server) Close() {
	s.isStarted = false
	if s.ln != nil {
		s.ln.Close()
	}
}
