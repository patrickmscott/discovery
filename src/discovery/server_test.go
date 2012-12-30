package discovery

import (
	"fmt"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"
	"time"
)

func TestServerSnapshot(t *testing.T) {
	server := NewServer()
	result := server.snapshot("group")
	if result.Len() != 0 {
		t.Error("Empty services list should return empty list")
	}

	server.services.Add(&ServiceDef{Host: "host1", Group: "group1"})
	server.services.Add(&ServiceDef{Host: "host2", Group: "group2"})

	result = server.snapshot("group")
	if result.Len() != 0 {
		t.Error("group snapshot is not empty")
	}

	result = server.snapshot("group1")
	if result.Len() != 1 {
		t.Error("group1 snapshot should be a single entry", result.Len())
	}
	service := result.Front().Value.(*ServiceDef)
	if service.Host != "host1" || service.Group != "group1" {
		t.Error("group1 entry has wrong values", service)
	}

	result = server.snapshot("group2")
	if result.Len() != 1 {
		t.Error("group2 snapshot should be a single entry", result.Len())
	}
	service = result.Front().Value.(*ServiceDef)
	if service.Host != "host2" || service.Group != "group2" {
		t.Error("group2 entry has wrong values", service)
	}
}

type testClientImpl struct {
	join, leave *ServiceDef
	signal      chan int
}

func (t *testClientImpl) Join(service *ServiceDef, v *Void) error {
	t.join = service
	t.signal <- 1
	return nil
}

func (t *testClientImpl) Leave(service *ServiceDef, v *Void) error {
	t.leave = service
	t.signal <- 1
	return nil
}

func serveTestImpl(impl *testClientImpl, conn io.ReadWriteCloser) {
	server := rpc.NewServer()
	server.RegisterName("DiscoveryClient", impl)
	go server.ServeCodec(jsonrpc.NewServerCodec(conn))
}

func TestServerJoin(t *testing.T) {
	server := NewServer()
	impl := &testClientImpl{signal: make(chan int)}
	read, write := net.Pipe()
	serveTestImpl(impl, read)
	server.watch("group1", jsonrpc.NewClient(write))

	if !server.join(&ServiceDef{Host: "h", Group: "group1"}) {
		t.Error("Server join failed")
	}
	<-impl.signal
	if impl.join == nil || impl.join.Host != "h" || impl.join.Group != "group1" {
		t.Error("Join callback incorrect", impl.join)
	}

	impl.join = nil
	if !server.join(&ServiceDef{Host: "h2", Port: 50, Group: "group1"}) {
		t.Error("Server join failed")
	}
	<-impl.signal
	if impl.join == nil || impl.join.Host != "h2" || impl.join.Port != 50 {
		t.Error("Join callback incorrect", impl.join)
	}

	impl.join = nil
	if !server.join(&ServiceDef{Host: "h", Group: "group2"}) {
		t.Error("Server join failed")
	}
	time.Sleep(20 * time.Millisecond)
	if impl.join != nil {
		t.Error("No join should have been sent")
	}

	if server.join(&ServiceDef{Host: "h", Group: "group1", connId: 1}) {
		t.Error("Server join should have failed")
	}
}

func BenchmarkSnapshot(b *testing.B) {
	server := NewServer()
	for i := 0; i < b.N; i++ {
		server.services.Add(
			&ServiceDef{Host: "host", Port: uint16(i), Group: "group"})
		if server.snapshot("group").Len() != i+1 {
			b.Error("Wrong snapshot size")
		}
	}
}

func BenchmarkRemoveAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		server := NewServer()
		for j := 0; j < 50; j++ {
			for k := 0; k < 5; k++ {
				host := fmt.Sprintf("host#%d-%d", j, k)
				server.services.Add(&ServiceDef{Host: host, connId: int32(k)})
			}
		}
		b.StartTimer()
		server.removeAll(&Discovery{id: 1})
	}
}
