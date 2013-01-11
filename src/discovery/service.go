package discovery

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Discovery struct {
	server *Server
	conn   net.Conn
	id     int32
	// TODO(pscott): Make this a channel that takes an error code?
	resultChan chan bool
	client     *rpc.Client
}

func newDiscoveryService(server *Server) *Discovery {
	return &Discovery{server: server, resultChan: make(chan bool, 1)}
}

func (d *Discovery) init(conn net.Conn, id int32) {
	d.conn = conn
	d.id = id
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

type Void struct{}

func (d *Discovery) Join(service *ServiceDef, v *Void) error {
	service.connId = d.id
	d.server.eventChan <- func() { d.resultChan <- d.server.join(service) }
	if !<-d.resultChan {
		return errors.New("Unable to add service")
	}
	return nil
}

func (d *Discovery) Leave(service *ServiceDef, v *Void) error {
	service.connId = d.id
	d.server.eventChan <- func() { d.resultChan <- d.server.leave(service) }
	if !<-d.resultChan {
		return errors.New("Unable to remove service")
	}
	return nil
}

func (d *Discovery) Snapshot(group string, snapshot *[]*ServiceDef) error {
	d.server.eventChan <- func() {
		services := d.server.snapshot(group)
		// TODO(pscott): Reuse an internal buffer, resizing if necessary. Might
		// require locking as multiple snapshot requests can be sent in parallel.
		*snapshot = make([]*ServiceDef, services.Len())
		i := 0
		for iter := services.Front(); iter != nil; iter = iter.Next() {
			(*snapshot)[i] = iter.Value.(*ServiceDef)
			i++
		}
		d.resultChan <- true
	}
	if !<-d.resultChan {
		return errors.New("Snapshot failed")
	}
	return nil
}

// Start watching changes to the given group. Never returns an error.
func (d *Discovery) Watch(group string, v *Void) error {
	d.server.eventChan <- func() {
		// Do the client check in the server event loop to avoid any locking or race
		// conditions.
		client := d.rpcClient()
		success := client != nil
		if success {
			d.server.watch(group, client)
		}
		d.resultChan <- success
	}
	if !<-d.resultChan {
		return errors.New(
			"Watch failed, unable to connect to DiscoveryClient service")
	}
	return nil
}

// Stop watching changes to the given group. Due to the asynchronous nature of
// this method, changes in route to the connection may be sent after this method
// is called. Never returns an error.
func (d *Discovery) Ignore(group string, v *Void) error {
	d.server.eventChan <- func() {
		if d.client != nil {
			d.server.ignore(group, d.client)
		}
		d.resultChan <- true
	}
	<-d.resultChan
	return nil
}
