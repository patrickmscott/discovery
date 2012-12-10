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

func (d *Discovery) Join(req *JoinRequest, v *Void) error {
	d.server.eventChan <- func() {
		d.resultChan <- d.server.join(&serviceDefinition{
			Host:       d.getHost(req.Host),
			Port:       req.Port,
			CustomData: req.CustomData,
			group:      req.Group,
			connId:     d.id})
	}
	// TODO(pscott): Make this a channel that takes an error code?
	if !<-d.resultChan {
		return errors.New("Unable to add service")
	}
	return nil
}

func (d *Discovery) Leave(req *LeaveRequest, v *Void) error {
	d.server.eventChan <- func() {
		d.resultChan <- d.server.leave(&serviceDefinition{
			Host:   d.getHost(req.Host),
			Port:   req.Port,
			group:  req.Group,
			connId: d.id})
	}
	if !<-d.resultChan {
		return errors.New("Unable to remove service")
	}
	return nil
}

func (d *Discovery) sendJoin(service *serviceDefinition)  {}
func (d *Discovery) sendLeave(service *serviceDefinition) {}
