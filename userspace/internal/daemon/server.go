package daemon

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync/atomic"
	"time"
)

const path = "/tmp/guardian.sock"

type Server struct {
	ln   net.Listener
	exec *Executor

	logger *slog.Logger

	isStarted atomic.Bool
	isClosed  atomic.Bool
}

func NewServer() *Server {
	ex := NewEmpty()
	ex.Init()
	return &Server{exec: ex, logger: slog.Default().With("component", "server")}
}

func (s *Server) Start() error {
	if s.isStarted.Load() {
		s.logger.Warn("start requested but daemon already running", "path", path)
		return fmt.Errorf("guardian already started")
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		s.logger.Warn("failed to remove stale socket", "path", path, "err", err)
	}

	s.logger.Info("starting unix listener", "path", path, "pid", os.Getpid())
	ln, err := net.Listen("unix", path)
	if err != nil {
		s.logger.Error("failed to start listener", "path", path, "err", err)
		return err
	}
	if ul, ok := ln.(*net.UnixListener); ok {
		ul.SetUnlinkOnClose(true)
	}

	s.isStarted.Store(true)
	s.ln = ln
	defer s.Close()

	s.logger.Info("listener ready", "path", path, "network", ln.Addr().Network(), "addr", ln.Addr().String())
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if !s.isStarted.Load() {
				s.logger.Info("accept loop stopped")
				return nil
			}
			s.logger.Error("failed to accept connection", "err", err)
			continue
		}

		if err := s.handleConn(conn); err != nil {
			s.logger.Error("failed to handle connection", "err", err)
		}
		if !s.isStarted.Load() {
			s.logger.Info("guardian stopped")
			return nil
		}
	}
}

func (s *Server) handleConn(c net.Conn) error {
	start := time.Now()
	s.logger.Info("accepted connection")

	conn := NewConn(c)
	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Warn("failed to close connection", "err", err)
		} else {
			s.logger.Debug("connection closed", "duration", time.Since(start))
		}
	}()

	req := Request{}
	if err := conn.Recieve(&req); err != nil {
		s.logger.Error("failed to receive request", "err", err, "duration", time.Since(start))
		return err
	}

	logger := s.logger.With("command", req.Command.String())
	logger.Info("received request")

	execFn, err := s.exec.Do(req.Command)
	if err != nil {
		logger.Error("unknown command", "err", err)
		resp := Response{Code: CodeInternal}
		if sendErr := conn.Send(&resp); sendErr != nil {
			logger.Error("failed to send response", "err", sendErr, "code", codeString(resp.Code))
			return sendErr
		}
		logger.Info("sent error response", "code", codeString(resp.Code), "duration", time.Since(start))
		return err
	}

	logger.Debug("dispatching command")
	resp, err := execFn(req)
	if err != nil {
		logger.Error("command execution failed",
			"err", err,
			"code", codeString(resp.Code),
			"duration", time.Since(start),
		)
	} else {
		logger.Info("command execution completed",
			"code", codeString(resp.Code),
			"duration", time.Since(start),
		)
	}

	if sendErr := conn.Send(&resp); sendErr != nil {
		logger.Error("failed to send response", "err", sendErr, "code", codeString(resp.Code))
		return sendErr
	}
	logger.Info("response sent", "code", codeString(resp.Code), "duration", time.Since(start))

	if req.Command == StopCmd {
		logger.Info("stop command received, shutting down")
		s.isStarted.Store(false)
	}
	return err
}

func (s *Server) Close() {
	if s.isClosed.Load() {
		s.logger.Debug("close skipped, already closed")
		return
	}

	s.logger.Info("closing listener", "path", path)
	s.isStarted.Store(false)
	if s.ln != nil {
		if err := s.ln.Close(); err != nil {
			s.logger.Error("failed to close listener", "path", path, "err", err)
		}
	}
	s.isClosed.Store(true)
	s.logger.Info("listener closed", "path", path)
}
