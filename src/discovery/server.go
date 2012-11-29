package discovery

import (
	"container/list"
	"fmt"
	"log"
	"net"
)

type event interface {
	Run(*Server)
}

type connectionEvent struct {
	event
	conn net.Conn
}

func (e *connectionEvent) Run(s *Server) {
	conn := newConnection(s)
	s.connections.PushBack(conn)
	go conn.Process(e.conn)
}

type serviceEvent struct {
	event
	join    bool
	service *serviceDefinition
}

func (e *serviceEvent) Run(s *Server) {
	if e.join {
		s.services.Add(e.service)
	} else {
		s.services.Remove(e.service)
	}
}

type Server struct {
	connections list.List
	services    serviceList
	eventChan   chan event
}

// Small default size to avoid allocating too much for small groups.
const defaultSnapshotSize = 32

func (s *Server) Snapshot(group string) *list.List {
	log.Printf("Snapshot: '%s'\n", group)
	var services list.List
	for iter := s.services.Iterator(group); iter.HasMore(); {
		services.PushBack(iter.Next())
	}
	return &services
}

// Listen for connections on the given port.
func (s *Server) Serve(port uint16) (err error) {
	log.Println("Listening on port", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}

	s.eventChan = make(chan event)
	go func() {
		log.Println("Event loop start...")
		for {
			e := <-s.eventChan
			e.Run(s)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		s.eventChan <- &connectionEvent{conn: conn}
	}
	return nil
}
