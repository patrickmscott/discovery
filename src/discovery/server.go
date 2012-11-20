package discovery

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Server struct {
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
		go s.processConnection(connection)
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

func (s *Server) processConnection(connection net.Conn) {
	connection.SetReadDeadline(time.Now().Add(1 * time.Minute))
	log.Println("Connection from", getIpAddress(connection.RemoteAddr()))
	msg, err := readRequest(connection)
	if err != nil {
		log.Println("Error parsing request", err)
	} else {
		log.Printf("%#v", msg)
	}
	connection.Close()
}
