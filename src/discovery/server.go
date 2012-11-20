package discovery

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type serviceEntry struct {
	ipAddress net.IP
	port      uint16
	data      []byte
}

type serviceGroup struct {
	entries list.List
}

type Server struct {
	mutex  sync.Mutex
	groups map[string]serviceGroup
}

func (s *Server) Join(req *JoinMessage) {
}

func (s *Server) Serve(port uint16) (err error) {
	log.Println("Listening on port", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(connection)
	}
	return nil
}

func getIpAddress(addr net.Addr) net.IP {
	tcpAddr := addr.(*net.TCPAddr)
	ip := tcpAddr.IP.To4()
	if tcpAddr.IP.IsLoopback() {
		ip = net.IPv4(127, 0, 0, 1)
	}
	return ip
}

func (s *Server) handleConnection(connection net.Conn) {
	log.Println("Handling connection from", getIpAddress(connection.RemoteAddr()))
	for {
		connection.SetReadDeadline(time.Now().Add(1 * time.Minute))
		req, err := readRequest(connection)
		if err != nil {
			log.Println("Error parsing request", err)
			break
		}

		switch req.Type() {
		case joinMessage:
			msg := req.ToJoin()
			log.Println("JOIN:", msg.Group, msg.Port)
		case leaveMessage:
			msg := req.ToLeave()
			log.Println("LEAVE:", msg.Group, msg.Port)
		case snapshotMessage:
			msg := req.ToSnapshot()
			log.Println("SNAPSHOT:", msg.Group)
		case watchMessage:
			msg := req.ToWatch()
			log.Println("WATCH:", msg.Groups)
		case heartbeatMessage:
		}
	}
	connection.Close()
}
