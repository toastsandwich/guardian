package daemon

import (
	"encoding/json"
	"net"
)

type Client struct {
	Path string
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) socketPath() string {
	if c != nil && c.Path != "" {
		return c.Path
	}
	return path
}

func (c *Client) call(cmd Command, body any) (Response, error) {
	var buf []byte
	if body != nil {
		var err error
		buf, err = json.Marshal(body)
		if err != nil {
			return Response{}, err
		}
	}

	conn, err := net.Dial("unix", c.socketPath())
	if err != nil {
		return Response{}, err
	}
	dc := NewConn(conn)
	defer dc.Close()

	err = dc.Send(&Request{
		Version: byte(0),
		Command: cmd,
		Body:    buf,
	})
	if err != nil {
		return Response{}, err
	}

	resp := Response{}
	if err := dc.Receive(&resp); err != nil {
		return Response{}, err
	}
	return resp, nil
}

func (c *Client) Attach(ao AttachOptions) (byte, error) {
	resp, err := c.call(AttachCmd, ao)
	if err != nil {
		return 0, err
	}
	return resp.Code, nil
}

func (c *Client) Detach() (byte, error) {
	resp, err := c.call(DettachCmd, nil)
	if err != nil {
		return 0, err
	}
	return resp.Code, nil
}

func (c *Client) Stop() (byte, error) {
	resp, err := c.call(StopCmd, nil)
	if err != nil {
		return 0, err
	}
	return resp.Code, nil
}

func (c *Client) Allow(ao AllowOptions) (byte, error) {
	resp, err := c.call(AllowCmd, ao)
	if err != nil {
		return 0, err
	}
	return resp.Code, nil
}

func (c *Client) Deny(do DenyOptions) (byte, error) {
	resp, err := c.call(DenyCmd, do)
	if err != nil {
		return 0, err
	}
	return resp.Code, nil
}

func (c *Client) GetMode() (string, byte, error) {
	resp, err := c.call(GetModeCmd, nil)
	if err != nil {
		return "", 0, err
	}

	mr := ModeResp{}
	if len(resp.Body) > 0 {
		if err := json.Unmarshal(resp.Body, &mr); err != nil {
			return "", resp.Code, err
		}
	}
	return mr.Mode, resp.Code, nil
}

func (c *Client) SetMode(so SetModeOptions) (byte, error) {
	resp, err := c.call(SetModeCmd, so)
	if err != nil {
		return 0, err
	}
	return resp.Code, nil
}

func (c *Client) List(lo ListOptions) (ListResp, byte, error) {
	resp, err := c.call(ListCmd, lo)
	if err != nil {
		return ListResp{}, 0, err
	}

	ips := ListResp{}
	if len(resp.Body) > 0 {
		if err := json.Unmarshal(resp.Body, &ips); err != nil {
			return ListResp{}, resp.Code, err
		}
	}
	return ips, resp.Code, nil
}
