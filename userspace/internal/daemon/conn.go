package daemon

import (
	"encoding/json"
	"net"
)

type Conn struct {
	c   net.Conn
	enc *json.Encoder
	dec *json.Decoder
}

func NewConn(c net.Conn) *Conn {
	return &Conn{
		c:   c,
		enc: json.NewEncoder(c),
		dec: json.NewDecoder(c),
	}
}

func (c *Conn) Send(v any) error {
	return c.enc.Encode(v)
}

func (c *Conn) Receive(v any) error {
	return c.dec.Decode(v)
}

func (c *Conn) Close() error {
	return c.c.Close()
}
