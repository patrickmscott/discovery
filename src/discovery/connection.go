package discovery

import (
	"log"
	"net"
)

type connection struct {
	server *Server
	ip     string
	id     int32
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
	service.connId = c.id
	c.server.eventChan <- func() {
		log.Println("Join: ", service)
		c.server.services.Add(&service)
	}
}

func (c *connection) Leave(req *LeaveRequest) {
	var service serviceDefinition
	service.Host = c.serviceHost(req.Host)
	service.Port = req.Port
	service.group = req.Group
	// Only need Host/Port/group tuple for Remove.
	c.server.eventChan <- func() {
		log.Println("Leave: ", service)
		c.server.services.Remove(&service)
	}
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
	c.server.eventChan <- func() {
		log.Printf("Removing all services for conn #%d\n", c.id)
		c.server.services.RemoveAll(c.id)
	}
	conn.Close()
}
