package discovery

import (
	"container/list"
	"fmt"
	"log"
	"net"
)

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

	s.eventChan = make(chan event, 1024)
	go func() {
		log.Println("Event loop start...")
		for {
			e := <-s.eventChan
			e.Dispatch(s)
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
