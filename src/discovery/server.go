package discovery

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func compareHostAndPort(a, b *JoinRequest) int {
	if a.Host == b.Host {
		return int(a.Port) - int(b.Port)
	}
	if a.Host < b.Host {
		return -1
	}
	return 1
}

func addEntry(list *list.List, entry *JoinRequest) {
	entry = entry.copy()
	for iter := list.Front(); iter != nil; iter = iter.Next() {
		e := iter.Value.(*JoinRequest)
		res := compareHostAndPort(entry, e)
		if res < 0 {
			list.InsertBefore(entry, iter)
			return
		} else if res == 0 {
			iter.Value = entry
			return
		}
	}
	list.PushBack(entry)
}

// Remove the given entry from the list. Compares Host and Port until an entry
// is found.
func removeEntry(list *list.List, entry *LeaveRequest) {
	for iter := list.Front(); iter != nil; iter = iter.Next() {
		e := iter.Value.(*JoinRequest)
		if e.Host > entry.Host {
			// The list is sorted alphabetically so we know that the entry does not
			// exist.
			break
		} else if e.Host == entry.Host && e.Port == entry.Port {
			list.Remove(iter)
			break
		}
	}
}

type Server struct {
	mutex  sync.Mutex
	groups map[string]*list.List
}

// Initialize the server state. Also can be used to reset the server at any
// point.
func (s *Server) Init() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.groups = make(map[string]*list.List)
}

// Add the given JoinRequest as an entry in the set of services. If the host and
// port already exist in the group, the entry is replaced.
func (s *Server) Join(req *JoinRequest) {
	log.Printf("Join: '%s' %s:%d\n", req.Group, req.Host, req.Port)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	services := s.groups[req.Group]
	if services == nil {
		services = &list.List{}
		s.groups[req.Group] = services
	}
	addEntry(services, req)
}

// Handle a leave event. This mostly happens when multiple services are
// broadcast on the same connection and one service leaves a group.
func (s *Server) Leave(req *LeaveRequest) {
	log.Printf("Leave: '%s' %s:%d\n", req.Group, req.Host, req.Port)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	services := s.groups[req.Group]
	if services == nil {
		return
	}
	removeEntry(services, req)
}

// Return value for Snapshot. Contains individual service information suitable
// for json serialization to a client.
type ServiceDefinition struct {
	Host       string `json:"host"`
	Port       uint16 `json:"port"`
	CustomData []byte `json:"customData,omitempty"`
}

func (s *Server) Snapshot(group string) []ServiceDefinition {
	log.Printf("Snapshot: '%s'\n", group)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	services := s.groups[group]
	if services == nil || services.Len() == 0 {
		return nil
	}
	slice := make([]ServiceDefinition, services.Len())
	i := 0
	for iter := services.Front(); iter != nil; iter = iter.Next() {
		join := iter.Value.(*JoinRequest)
		broadcast := &slice[i]
		broadcast.Host = join.Host
		broadcast.Port = join.Port
		broadcast.CustomData = join.CustomData
		i++
	}
	return slice
}

// Listen for connections on the given port.
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

func (s *Server) handleConnection(connection net.Conn) {
	var proto Protocol
	log.Println("Handling connection from", getIpAddress(connection.RemoteAddr()))
	for {
		connection.SetReadDeadline(time.Now().Add(1 * time.Minute))
		req, err := proto.readRequest(connection)
		if err != nil {
			log.Println("Error parsing request", err)
			break
		}

		switch req.Type() {
		case joinRequest:
			s.Join(req.(*JoinRequest))
		case leaveRequest:
			s.Leave(req.(*LeaveRequest))
		case snapshotRequest:
			msg := req.(*SnapshotRequest)
			result := s.Snapshot(msg.Group)
			connection.SetWriteDeadline(time.Now().Add(1 * time.Minute))
			if err := proto.writeJson(connection, result); err != nil {
				break
			}
		case watchRequest:
			msg := req.(*WatchRequest)
			log.Println("WATCH:", msg.Groups)
		case heartbeatRequest:
		}
	}
	connection.Close()
}
