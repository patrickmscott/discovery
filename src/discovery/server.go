package discovery

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	connections list.List
	services    serviceList
	eventChan   chan func()
	nextConnId  int32
	watchers    map[string](map[*connection]bool)
}

// Small default size to avoid allocating too much for small groups.
const defaultSnapshotSize = 32

func (s *Server) Snapshot(group string) *list.List {
	log.Printf("Snapshot: '%s'\n", group)
	var services list.List
	for iter := s.services.GroupIterator(group); iter.HasMore(); {
		services.PushBack(iter.Next())
	}
	return &services
}

func (s *Server) join(service *serviceDefinition) {
	if !s.services.Add(service) {
		return
	}
	log.Println("Join: ", service)
	connections, ok := s.watchers[service.group]
	if !ok {
		return
	}
	for conn := range connections {
		conn.SendJoin(service)
	}
}

func (s *Server) leave(service *serviceDefinition) {
	if !s.services.Remove(service) {
		return
	}
	log.Println("Leave: ", service)
	connections, ok := s.watchers[service.group]
	if !ok {
		return
	}
	for conn := range connections {
		conn.SendLeave(service)
	}
}

func (s *Server) removeAll(conn *connection) {
	for iter := s.services.ConnIterator(conn.id); iter.HasMore(); {
		service := iter.Remove()
		log.Println("Leave: ", service)
		connections, ok := s.watchers[service.group]
		if !ok {
			continue
		}
		for c := range connections {
			c.SendLeave(service)
		}
	}
}

func (s *Server) watch(group string, conn *connection) {
	s.watchers[group][conn] = true
}

// Listen for connections on the given port.
func (s *Server) Serve(port uint16) (err error) {
	log.Println("Listening on port", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}

	s.eventChan = make(chan func(), 1024)
	go func() {
		log.Println("Event loop start...")
		for {
			(<-s.eventChan)()
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		s.eventChan <- func() {
			c := &connection{server: s, id: atomic.AddInt32(&s.nextConnId, 1)}
			s.connections.PushBack(c)
			go c.Process(conn)
		}
	}
	return nil
}
