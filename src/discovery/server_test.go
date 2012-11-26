package discovery

import (
	"container/list"
	"testing"
)

func TestAddEntry(t *testing.T) {
	var list list.List
	var join JoinMessage
	addEntry(&list, &join)
	if list.Len() != 1 {
		t.Error()
	}
	join.Host = "host"
	addEntry(&list, &join)
	if list.Len() != 2 {
		t.Error()
	}
	e := list.Front().Value.(*JoinMessage)
	if e.Host != "" {
		t.Error(e.Host)
	}
	e = list.Back().Value.(*JoinMessage)
	if e.Host != "host" {
		t.Error(e.Host)
	}
	addEntry(&list, &join)
	if list.Len() != 2 {
		t.Error()
	}
	join.CustomData = make([]byte, 5)
	addEntry(&list, &join)
	if list.Len() != 2 {
		t.Error()
	}
	e = list.Back().Value.(*JoinMessage)
	if e.CustomData == nil {
		t.Error()
	}

	// Clear the list
	list.Init()
	for i := 0; i < 100; i++ {
		join.Port = uint16(i)
		addEntry(&list, &join)
	}
	var i uint16 = 0
	for elem := list.Front(); elem != nil; elem = elem.Next() {
		e = elem.Value.(*JoinMessage)
		if e.Port != i {
			t.Error()
		}
		i++
	}
}

func TestCompareHostAndPort(t *testing.T) {
	var a, b JoinMessage
	a.Host = "host1"
	b.Host = "host2"
	if compareHostAndPort(&a, &b) != -1 {
		t.Error()
	}
	b.Host = "host1"
	if compareHostAndPort(&a, &b) != 0 {
		t.Error()
	}
	b.Host = "host0"
	if compareHostAndPort(&a, &b) != 1 {
		t.Error()
	}
	b.Host = "host1"
	a.Port = 1000
	b.Port = 1020
	if compareHostAndPort(&a, &b) != -20 {
		t.Error()
	}
	b.Port = 1000
	if compareHostAndPort(&a, &b) != 0 {
		t.Error()
	}
}

func TestServerInit(t *testing.T) {
	var server Server
	if server.groups != nil {
		t.Error()
	}
	server.Init()
	if server.groups == nil || len(server.groups) != 0 {
		t.Error()
	}
	server.groups["group"] = &list.List{}
	if len(server.groups) != 1 {
		t.Error()
	}
	server.Init()
	if len(server.groups) != 0 {
		t.Error()
	}
}

func TestServerJoin(t *testing.T) {
	var server Server
	server.Init()
	var join JoinMessage
	join.Host = "host"
	join.Port = 8080
	join.Group = "my group"
	server.Join(&join)
	if server.groups["my group"].Len() != 1 {
		t.Error()
	}
	if len(server.groups) != 1 {
		t.Error()
	}
	join.CustomData = make([]byte, 5)
	server.Join(&join)
	if server.groups["my group"].Len() != 1 {
		t.Error()
	}
	if server.groups["my group"].Front().Value.(*JoinMessage).CustomData == nil {
		t.Error()
	}
}

func TestServerSnapshot(t *testing.T) {
	var server Server
	server.Init()
	result := server.Snapshot("group")
	if result != nil {
		t.Error()
	}
	var join JoinMessage
	join.Host = "host1"
	join.Port = 8080
	join.Group = "group1"
	server.Join(&join)
	join.Host = "host2"
	join.Group = "group2"
	server.Join(&join)
	result = server.Snapshot("group")
	if result != nil {
		t.Error()
	}
	result = server.Snapshot("group1")
	if len(result) != 1 {
		t.Error()
	}
	service := result[0]
	if service.Host != "host1" || service.Port != 8080 {
		t.Error()
	}
	result = server.Snapshot("group2")
	if len(result) != 1 {
		t.Error()
	}
	service = result[0]
	if service.Host != "host2" || service.Port != 8080 {
		t.Error()
	}
}

func TestServerLeave(t *testing.T) {
	var server Server
	server.Init()
	var join JoinMessage
	join.Host = "host"
	join.Port = 8080
	join.Group = "group"
	server.Join(&join)
	result := server.Snapshot("group")
	if result == nil || len(result) != 1 {
		t.Error()
	}

	var leave LeaveMessage
	leave.Host = "host"
	leave.Port = 8080
	leave.Group = "group1"
	server.Leave(&leave)
	result = server.Snapshot("group")
	if result == nil || len(result) != 1 {
		t.Error()
	}
	result = server.Snapshot("group1")
	if result != nil {
		t.Error()
	}
	leave.Group = "group"
	server.Leave(&leave)
	result = server.Snapshot("group")
	if result != nil {
		t.Error()
	}
}

func BenchmarkSnapshot(b *testing.B) {
	var server Server
	server.Init()
	var join JoinMessage
	join.Host = "host"
	join.Group = "group"
	for i := 0; i < b.N; i++ {
		join.Port = uint16(i)
		server.Join(&join)
		server.Snapshot("group")
	}
}
