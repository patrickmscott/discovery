package discovery

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	mutex       sync.Mutex
	connections list.List
}

// Small default size to avoid allocating too much for small groups.
const defaultSnapshotSize = 32

func (s *Server) Snapshot(group string) *list.List {
	log.Printf("Snapshot: '%s'\n", group)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var services list.List
	for iter := s.connections.Front(); iter != nil; iter = iter.Next() {
		conn := iter.Value.(*connection)
		conn.mutex.Lock()
		servicesIter := conn.services.Iterator(group)
		for servicesIter.HasMore() {
			services.PushBack(servicesIter.Next())
		}
		conn.mutex.Unlock()
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

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		connection := newConnection(s)
		s.mutex.Lock()
		s.connections.PushBack(connection)
		s.mutex.Unlock()
		go connection.Process(conn)
	}
	return nil
}
