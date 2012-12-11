package discovery

import (
	"fmt"
	"testing"
)

func TestServerSnapshot(t *testing.T) {
	var server Server
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

func BenchmarkSnapshot(b *testing.B) {
	var server Server
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
		var server Server
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
