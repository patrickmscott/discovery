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

func (c *Client) Join(service *ServiceDef) error {
	return c.client.Call("Discovery.Join", service, &Void{})
}

func (c *Client) Leave(service *ServiceDef) error {
	return c.client.Call("Discovery.Leave", service, &Void{})
}

func (c *Client) Snapshot(group string) ([]*ServiceDef, error) {
	var services []*ServiceDef
	err := c.client.Call("Discovery.Snapshot", group, &services)
	return services, err
}
