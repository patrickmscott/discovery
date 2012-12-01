package discovery

import (
	"log"
	"net"
)

// Generic event interface for running an event on the main server event thread.
type event interface {
	Dispatch(*Server)
}

type connectionEvent struct {
	event
	conn net.Conn
}

func (e *connectionEvent) Dispatch(s *Server) {
	conn := &connection{server: s}
	s.connections.PushBack(conn)
	go conn.Process(e.conn)
}

type serviceEvent struct {
	event
	// True if this is a join event, false if it is a leave.
	join    bool
	service *serviceDefinition
}

func (e *serviceEvent) Dispatch(s *Server) {
	if e.join {
		log.Println("Join:  ", e.service)
		s.services.Add(e.service)
	} else {
		log.Println("Leave: ", e.service)
		s.services.Remove(e.service)
	}
}
