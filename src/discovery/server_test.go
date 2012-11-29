package discovery

import (
	"testing"
)

func TestServerSnapshot(t *testing.T) {
	var server Server
	result := server.Snapshot("group")
	if result.Len() != 0 {
		t.Error("Empty connection list should return empty list")
	}

	var conn1, conn2 = newConnection(&server), newConnection(&server)
	server.connections.PushBack(conn1)
	server.connections.PushBack(conn2)
	result = server.Snapshot("group")
	if result.Len() != 0 {
		t.Error("Empty services list should return empty list")
	}

	conn1.Join(&JoinRequest{Host: "host1", Group: "group1"})
	conn2.Join(&JoinRequest{Host: "host2", Group: "group2"})

	result = server.Snapshot("group")
	if result.Len() != 0 {
		t.Error("group snapshot is not empty")
	}

	result = server.Snapshot("group1")
	if result.Len() != 1 {
		t.Error("group1 snapshot should be a single entry", result.Len())
	}
	service := result.Front().Value.(*serviceDefinition)
	if service.Host != "host1" || service.group != "group1" {
		t.Error("group1 entry has wrong values", service)
	}

	result = server.Snapshot("group2")
	if result.Len() != 1 {
		t.Error("group2 snapshot should be a single entry", result.Len())
	}
	service = result.Front().Value.(*serviceDefinition)
	if service.Host != "host2" || service.group != "group2" {
		t.Error("group2 entry has wrong values", service)
	}
}

func BenchmarkSnapshot(b *testing.B) {
	var server Server
	var conn *connection = newConnection(&server)
	server.connections.PushBack(conn)
	var join JoinRequest
	join.Host = "host"
	join.Group = "group"
	for i := 0; i < b.N; i++ {
		join.Port = uint16(i)
		conn.Join(&JoinRequest{Host: "host", Port: uint16(i), Group: "group"})
		if server.Snapshot("group").Len() != i+1 {
			b.Error("Wrong snapshot size")
		}
	}
}
