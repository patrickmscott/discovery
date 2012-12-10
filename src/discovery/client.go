package discovery

import (
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
)

const DefaultPort uint16 = 3472 /* DISC */

type Client struct {
	client *rpc.Client
}

func (c *Client) Connect(host string, port uint16) error {
	client, err := jsonrpc.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	c.client = client
	return err
}

func (c *Client) Join(join *JoinRequest) error {
	return c.client.Call("Discovery.Join", join, &Void{})
}

func (c *Client) Leave(leave *LeaveRequest) error {
	return c.client.Call("Discovery.Leave", leave, &Void{})
}
