package discovery

import (
	"container/list"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync/atomic"
)

type Server struct {
	connections list.List
	services    serviceList
	eventChan   chan func()
	servicePool chan *Discovery
	nextConnId  int32
	watchers    map[string](map[*rpc.Client]bool)
}

func (s *Server) snapshot(group string) *list.List {
	log.Printf("Snapshot: '%s'\n", group)
	var services list.List
	iter := s.services.Iterator()
	for {
		service := iter.Next()
		if service == nil {
			break
		}
		// Services are ordered by group first so once we go beyond a group, we
		// don't need to keep iterating.
		diff := strcmp(service.Group, group)
		if diff > 0 {
			break
		} else if diff == 0 {
			services.PushBack(service)
		}
	}
	return &services
}

func (s *Server) join(service *ServiceDef) bool {
	if !s.services.Add(service) {
		return false
	}
	log.Println("Join:", service.toString())
	clients, ok := s.watchers[service.Group]
	if ok {
		for client := range clients {
			// Although client.Go does not wait for the response, it does block
			// sending the request. Invoke the Go method in a separate go routine to
			// not block.
			go client.Go("DiscoveryClient.Join", service, &Void{}, nil)
		}
	}
	return true
}

func (s *Server) leave(service *ServiceDef) bool {
	if !s.services.Remove(service) {
		return false
	}
	s.sendLeave(service)
	return true
}

func (s *Server) sendLeave(service *ServiceDef) {
	log.Println("Leave:", service.toString())
	clients, ok := s.watchers[service.Group]
	if ok {
		for client := range clients {
			// Although client.Go does not wait for the response, it does block
			// sending the request. Invoke the Go method in a separate go routine to
			// not block.
			go client.Go("DiscoveryClient.Leave", service, &Void{}, nil)
		}
	}
}

func (s *Server) removeAll(d *Discovery) {
	// Get rid of any watchers on this connection.
	if d.client != nil {
		for group, val := range s.watchers {
			delete(val, d.client)
			if len(val) == 0 {
				delete(s.watchers, group)
			}
		}
	}

	iter := s.services.Iterator()
	for {
		service := iter.Next()
		if service == nil {
			break
		}
		// Not the right connection id, just keep iterating.
		if service.connId != d.id {
			continue
		}
		iter.Remove()
		s.sendLeave(service)
	}
}

func (s *Server) watch(group string, client *rpc.Client) {
	m, ok := s.watchers[group]
	if !ok {
		s.watchers[group] = make(map[*rpc.Client]bool)
		m = s.watchers[group]
	}
	m[client] = true
}

func (s *Server) ignore(group string, client *rpc.Client) {
	if m, ok := s.watchers[group]; ok {
		delete(m, client)
	}
}

func NewServer() *Server {
	return &Server{
		// TODO(pscott): Add flags for event and service buffer size.
		eventChan:   make(chan func(), 1024),
		servicePool: make(chan *Discovery, 128),
		watchers:    make(map[string]map[*rpc.Client]bool)}
}

func (s *Server) processEvents() {
	log.Println("Event loop start...")
	for {
		(<-s.eventChan)()
	}
}

// Listen for connections on the given port.
func (s *Server) Serve(port uint16) (err error) {
	log.Println("Listening on port", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}

	go s.processEvents()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
	return nil
}

var debug = flag.Bool(
	"debugRpc", false, "Enable debug output of all rpc traffic")

type debugInput struct {
	rwc io.ReadWriteCloser
}

func (in *debugInput) Read(p []byte) (n int, err error) {
	n, err = in.rwc.Read(p)
	log.Println("rpc[r]:", string(p), err)
	return
}

func (in *debugInput) Write(p []byte) (n int, err error) {
	n, err = in.rwc.Write(p)
	log.Println("rpc[w]:", string(p), err)
	return
}

func (in *debugInput) Close() error {
	return in.rwc.Close()
}

func (s *Server) handleConnection(conn net.Conn) {
	// We create a new server each time so that we can have access to the
	// underlying connection. The standard rpc package does not give us access
	// to the calling connection :/
	var server *rpc.Server = rpc.NewServer()

	// Get a free service from the pool.
	var service *Discovery
	select {
	case service = <-s.servicePool:
		// Success
	default:
		service = newDiscoveryService(s)
	}

	// Set up the service variables.
	service.init(conn, atomic.AddInt32(&s.nextConnId, 1))

	// If debugging is enabled, log all rpc traffic.
	var rwc io.ReadWriteCloser = conn
	if *debug {
		rwc = &debugInput{conn}
	}

	// Set up the rpc service and start serving the connection.
	server.Register(service)
	server.ServeCodec(jsonrpc.NewServerCodec(rwc))

	// Connection has disconnected. Remove any registered services.
	s.removeAll(service)

	select {
	case s.servicePool <- service:
		// Success
	default:
		// Buffer is full
	}
}
