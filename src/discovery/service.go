package discovery

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

type Discovery struct {
	server *Server
	conn   net.Conn
	id     int32
	client *rpc.Client
}

func newDiscoveryService(server *Server) *Discovery {
	return &Discovery{server: server}
}

func (d *Discovery) init(conn net.Conn, id int32) {
	d.conn = conn
	d.id = id
	d.client = nil
}

func (d *Discovery) rpcClient() *rpc.Client {
	if d.client == nil {
		tcpAddr, ok := d.conn.RemoteAddr().(*net.TCPAddr)
		if !ok {
			return nil
		}
		// TODO(pscott): Figure out how to multiplex this connection. If we could
		// reuse the same connection, we could avoid having to assume the client
		// is running the DiscoveryClient service on the default port.
		address := fmt.Sprintf("%s:%d", tcpAddr.IP.String(), DefaultPort)

		// A failure in connecting will return a nil client.
		d.client, _ = jsonrpc.Dial("tcp", address)
	}
	return d.client
}

// run takes a closure that returns an error. It runs the function in the main
// server event loop and returns any error that the function returns. If the
// function times out, run will return a timeout error.
func (d *Discovery) run(f func() error) error {
	result := make(chan error, 1)
	d.server.eventChan <- func() { result <- f() }
	select {
	case err := <-result:
		return err
	// TODO(pscott): make this configurable
	case <-time.After(2 * time.Second):
		return errors.New("Method timeout")
	}
	panic("Unreachable statement")
}

type Void struct{}

func (d *Discovery) Join(service *ServiceDef, v *Void) error {
	service.connId = d.id
	return d.run(func() error {
		if !d.server.join(service) {
			return errors.New("Unable to add service")
		}
		return nil
	})
}

func (d *Discovery) Leave(service *ServiceDef, v *Void) error {
	service.connId = d.id
	return d.run(func() error {
		if !d.server.leave(service) {
			return errors.New("Unable to remove service")
		}
		return nil
	})
}

func (d *Discovery) Snapshot(group string, snapshot *[]*ServiceDef) error {
	return d.run(func() error {
		services := d.server.snapshot(group)
		// TODO(pscott): Reuse an internal buffer, resizing if necessary. Might
		// require locking as multiple snapshot requests can be sent in parallel.
		*snapshot = make([]*ServiceDef, services.Len())
		i := 0
		for iter := services.Front(); iter != nil; iter = iter.Next() {
			(*snapshot)[i] = iter.Value.(*ServiceDef)
			i++
		}
		return nil
	})
}

// Start watching changes to the given group. Never returns an error.
func (d *Discovery) Watch(group string, v *Void) error {
	return d.run(func() error {
		// Do the client check in the server event loop to avoid any locking or race
		// conditions.
		client := d.rpcClient()
		success := client != nil
		if success {
			d.server.watch(group, client)
			return nil
		}
		return errors.New("Watch failed: unable to connect to client")
	})
}

// Stop watching changes to the given group. Due to the asynchronous nature of
// this method, changes in route to the connection may be sent after this method
// is called. Never returns an error.
func (d *Discovery) Ignore(group string, v *Void) error {
	return d.run(func() error {
		if d.client != nil {
			d.server.ignore(group, d.client)
		}
		return nil
	})
}
