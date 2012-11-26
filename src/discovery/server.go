package discovery

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func compareHostAndPort(a, b *JoinMessage) int {
	if a.Host == b.Host {
		return int(a.Port) - int(b.Port)
	}
	if a.Host < b.Host {
		return -1
	}
	return 1
}

func addEntry(list *list.List, entry *JoinMessage) {
	entry = entry.Copy()
	for iter := list.Front(); iter != nil; iter = iter.Next() {
		e := iter.Value.(*JoinMessage)
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

// Add the given JoinMessage as an entry in the set of services. If the host and
// port already exist in the group, the entry is replaced.
func (s *Server) Join(req *JoinMessage) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	services := s.groups[req.Group]
	if services == nil {
		services = &list.List{}
		s.groups[req.Group] = services
	}
	addEntry(services, req)
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

func getIpAddress(addr net.Addr) string {
	tcpAddr := addr.(*net.TCPAddr)
	ip := tcpAddr.IP.To4().String()
	if tcpAddr.IP.IsLoopback() {
		ip = "127.0.0.1"
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
