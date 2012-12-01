package discovery

import (
	"log"
	"net"
)

type connection struct {
	server   *Server
	services serviceList
	ip       string
}

func (c *connection) serviceHost(host string) string {
	if host == "" {
		return c.ip
	}
	return host
}

func (c *connection) Join(req *JoinRequest) {
	var service serviceDefinition
	service.Host = c.serviceHost(req.Host)
	service.Port = req.Port
	service.CustomData = req.CustomData
	service.group = req.Group
	c.server.eventChan <- &serviceEvent{join: true, service: &service}
}

func (c *connection) Leave(req *LeaveRequest) {
	var service serviceDefinition
	service.Host = c.serviceHost(req.Host)
	service.Port = req.Port
	service.group = req.Group
	c.server.eventChan <- &serviceEvent{join: false, service: &service}
}

func getIpAddress(addr net.Addr) string {
	tcpAddr := addr.(*net.TCPAddr)
	if tcpAddr.IP.IsLoopback() {
		return "127.0.0.1"
	}
	return tcpAddr.IP.String()
}

// Handle incoming requests until either the client disconnects or there is an
// error.
func (c *connection) Process(conn net.Conn) {
	var proto Protocol
	c.ip = getIpAddress(conn.RemoteAddr())
	log.Println("Processing connection from", c.ip)

	for {
		req, err := proto.readRequest(conn)
		if err != nil {
			break
		}

		switch req.Type() {
		case joinRequest:
			c.Join(req.(*JoinRequest))
		case leaveRequest:
			c.Leave(req.(*LeaveRequest))
		case snapshotRequest:
			req := req.(*SnapshotRequest)
			result := c.server.Snapshot(req.Group)
			if err := proto.writeJson(conn, result); err != nil {
				break
			}
		case watchRequest:
		}
	}
	log.Println("Closing connection from", c.ip)
	conn.Close()
}
