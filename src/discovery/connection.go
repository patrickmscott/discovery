package discovery

import (
	"log"
	"net"
	"sync"
)

type connection struct {
	server   *Server
	mutex    sync.Mutex
	services serviceList
	ip       string
}

func (c *connection) Join(req *JoinRequest) {
	var service serviceDefinition
	service.Host = req.Host
	service.Port = req.Port
	service.CustomData = req.CustomData
	service.group = req.Group

	// If the request did not include a host, use the connection ip.
	if service.Host == "" {
		service.Host = c.ip
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.services.Add(&service)
}

func (c *connection) Leave(req *LeaveRequest) {
	var service serviceDefinition
	service.Host = req.Host
	service.Port = req.Port
	service.group = req.Group

	// If the request did not include a host, use the connection ip.
	if service.Host == "" {
		service.Host = c.ip
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.services.Remove(&service)
}

func newConnection(server *Server) *connection {
	return &connection{server: server}
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
	conn.Close()
}
