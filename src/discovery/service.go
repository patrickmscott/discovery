package discovery

import (
	"errors"
	"net"
)

type Discovery struct {
	server     *Server
	conn       net.Conn
	id         int32
	resultChan chan bool
}

func newDiscoveryService(server *Server) *Discovery {
	return &Discovery{server: server, resultChan: make(chan bool, 1)}
}

func (d *Discovery) init(conn net.Conn, id int32) {
	d.conn = conn
	d.id = id
}

func (d *Discovery) getHost(host string) string {
	if host != "" {
		return host
	}
	return d.conn.RemoteAddr().String()
}

type Void struct{}

func (d *Discovery) Join(service *ServiceDef, v *Void) error {
	service.connId = d.id
	d.server.eventChan <- func() { d.resultChan <- d.server.join(service) }
	// TODO(pscott): Make this a channel that takes an error code?
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
		// TODO(pscott): Reuse an internal buffer, resizing if necessary.
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

func (d *Discovery) sendJoin(service *ServiceDef)  {}
func (d *Discovery) sendLeave(service *ServiceDef) {}
