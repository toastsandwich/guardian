package daemon

import "github.com/toastsandwich/guardian/internal/guardian"

func HandlePing(_ *Request) (*Response, error) {
	return &Response{Status: StatusOK}, nil
}

func (s *Server) HandleAttach(req *Request) (*Response, error) {
	g := guardian.New(req.IfaceName)
	s.guardian.Store(g)

	g = s.guardian.Load()
	err := g.Attach()
	if err != nil {
		return nil, err
	}

	return &Response{Status: StatusAttached}, nil
}
