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

func (s *Server) processConnection(connection net.Conn) {
	connection.SetReadDeadline(time.Now().Add(1 * time.Minute))
	//proto.SetIpFromAddr(connection.RemoteAddr())
	//log.Println("Connection from", proto.IpAddress)
	msg, err := readRequest(connection)
	if err != nil {
		log.Println("Error parsing request", err)
	}
	log.Printf("%#v", msg)
	connection.Close()
}
