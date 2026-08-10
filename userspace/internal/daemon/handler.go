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

type ShowResult struct {
	IP      string
	Allowed bool
}

func (s *Server) HandleShow() *Response {
	ch := s.guardian.Load().Show()
	resp := &Response{}
	showRes := []ShowResult{}
	for res := range ch {
		showRes = append(showRes, ShowResult{res.IP, res.Allowed})
	}
	resp.Ips = showRes
	return resp
}
