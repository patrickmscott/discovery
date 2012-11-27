package discovery

import (
	"fmt"
	"net"
)

type Client struct {
	proto Protocol
	conn net.Conn
}

func (c *Client) Connect(host string, port uint16) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	c.conn = conn
	return err
}

func (c *Client) Join(join *JoinRequest) error {
	return c.proto.writeRequest(c.conn, join)
}

func (c *Client) Leave(leave *LeaveRequest) error {
	return c.proto.writeRequest(c.conn, leave)
}
