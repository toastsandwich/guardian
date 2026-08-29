package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"text/tabwriter"
)

func DialAttach(ao AttachOptions) (int, error) {
	buf, err := json.Marshal(&ao)
	if err != nil {
		return -1, err
	}
	conn, err := net.Dial("unix", path)
	if err != nil {
		return -1, err
	}

	c := NewConn(conn)
	defer c.Close()

	err = c.Send(&Request{
		Version: byte(0),
		Command: AttachCmd,
		Body:    buf,
	})
	if err != nil {
		return -1, err
	}

	resp := Response{}
	err = c.Recieve(&resp)
	if err != nil {
		return -1, err
	}
	return int(resp.Code), nil
}

func DialDetach() (int, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return -1, err
	}

	c := NewConn(conn)
	defer c.Close()

	err = c.Send(&Request{
		Command: DettachCmd,
	})
	if err != nil {
		return -1, err
	}

	resp := Response{}
	err = c.Recieve(&resp)
	if err != nil {
		return -1, err
	}
	return int(resp.Code), nil
}

func DialStop() (int, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return -1, err
	}

	c := NewConn(conn)
	defer c.Close()

	err = c.Send(&Request{
		Command: StopCmd,
	})
	if err != nil {
		return -1, err
	}

	resp := Response{}
	err = c.Recieve(&resp)
	if err != nil {
		return -1, err
	}
	return int(resp.Code), nil
}

func DialList(b *strings.Builder, lo ListOptions) (int, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return -1, err
	}

	c := NewConn(conn)
	defer c.Close()

	rb, err := json.Marshal(&lo)
	if err != nil {
		return -1, err
	}

	err = c.Send(&Request{
		Command: ListCmd,
		Body:    rb,
	})

	ips := ListResp{}
	res := Response{}
	err = c.Recieve(&res)
	if err != nil {
		return -1, nil
	}

	err = json.Unmarshal(res.Body, &ips)
	if err != nil {
		return -1, err
	}

	status := func(p bool) string {
		if p {
			return "Allowed"
		}
		return "Denied"
	}
	tw := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "\nIP\tSTATUS")
	fmt.Fprintln(tw, "--\t------")
	for _, ip := range ips.IPs {
		fmt.Fprintf(
			tw,
			"%s\t%s\n", ip.IP, status(ip.Allow),
		)
	}
	tw.Flush()

	return int(res.Code), nil
}

func DialDeny(do DenyOptions) (int, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return -1, err
	}

	c := NewConn(conn)
	defer c.Close()

	buf, err := json.Marshal(&do)
	if err != nil {
		return -1, err
	}

	req := Request{
		Command: DenyCmd,
		Body:    buf,
	}
	err = c.Send(&req)
	if err != nil {
		return -1, err
	}

	resp := Response{}
	err = c.Recieve(&resp)
	if err != nil {
		return -1, err
	}
	return int(resp.Code), nil
}

func DialAllow(ao AllowOptions) (int, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return -1, err
	}

	c := NewConn(conn)
	defer c.Close()

	buf, err := json.Marshal(&ao)
	if err != nil {
		return -1, err
	}

	req := Request{
		Command: AllowCmd,
		Body:    buf,
	}
	err = c.Send(&req)
	if err != nil {
		return -1, err
	}

	resp := Response{}
	err = c.Recieve(&resp)
	if err != nil {
		return -1, err
	}
	return int(resp.Code), nil

}
