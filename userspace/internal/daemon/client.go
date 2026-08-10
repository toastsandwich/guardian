package daemon

import (
	"encoding/json"
	"net"
)

func PingDial(r Request) Response {
	conn, err := net.Dial("unix", Path)
	if err != nil {
		return Response{Error: err}
	}
	defer conn.Close()

	b, err := json.Marshal(&r)
	if err != nil {
		return Response{Error: err}
	}

	_, err = conn.Write(b)
	if err != nil {
		return Response{Error: err}
	}

	res := make([]byte, 1024)
	n, err := conn.Read(res)
	if err != nil {
		return Response{Error: err}
	}
	resp := Response{}
	err = json.Unmarshal(res[:n], &resp)
	if err != nil {
		return Response{Error: err}
	}
	return resp
}

func StopDial(r Request) error {
	conn, err := net.Dial("unix", Path)
	if err != nil {
		return err
	}
	defer conn.Close()

	b, err := json.Marshal(&r)
	if err != nil {
		return err
	}

	_, err = conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func AttachDial(r Request) Response {
	conn, err := net.Dial("unix", Path)
	if err != nil {
		return Response{Error: err}
	}
	defer conn.Close()

	b, err := json.Marshal(&r)
	if err != nil {
		return Response{Error: err}
	}

	_, err = conn.Write(b)
	if err != nil {
		return Response{Error: err}
	}

	res := make([]byte, 1024)
	n, err := conn.Read(res)
	if err != nil {
		return Response{Error: err}
	}
	resp := Response{}
	err = json.Unmarshal(res[:n], &resp)
	if err != nil {
		return Response{Error: err}
	}
	return resp
}
