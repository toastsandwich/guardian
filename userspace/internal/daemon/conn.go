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

func (c *Conn) Recieve(v any) error {
	buf := make([]byte, 4096)
	n, err := c.c.Read(buf)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf[:n], v)
	// return c.dec.Decode(v)
}

func (c *Conn) Close() error {
	return c.c.Close()
}
