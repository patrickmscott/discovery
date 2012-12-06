package discovery

import (
	"fmt"
	"testing"
)

func TestServiceDefinitionCompare(t *testing.T) {
	var a, b serviceDefinition
	if a.compare(&b) != 0 {
		t.Error("Empty definition compare failed")
	}

	a.Host = "host1"
	b.Host = "host2"
	if a.compare(&b) >= 0 {
		t.Error("a should be less than b: host")
	}

	a.Host = "host"
	b.Host = "host"
	a.group = "group1"
	b.group = "group2"
	if a.compare(&b) >= 0 {
		t.Error("a should be less than b: group")
	}

	a.group = "group"
	b.group = "group"
	a.Port = 1
	b.Port = 2
	if a.compare(&b) >= 0 {
		t.Error("a should be less than b: port")
	}
}

func TestServiceListAdd(t *testing.T) {
	var list serviceList
	service := &serviceDefinition{Host: "host"}

	if list.Len() != 0 {
		t.Error("Empty list should have 0 length")
	}

	list.Add(service)
	if list.Len() != 1 {
		t.Error("Single add failed")
	}

	service = &serviceDefinition{Host: "host1"}
	list.Add(service)
	if list.Len() != 2 {
		t.Error("Multiple add failed")
	}

	list.Add(service)
	if list.Len() != 2 {
		t.Error("Duplicate add failed")
	}

	service = list.Get(0)
	if service.Host != "host" {
		t.Error("First entry invalid", service)
	}

	service = list.Get(1)
	if service.Host != "host1" {
		t.Error("Second entry invalid", service)
	}

	service = &serviceDefinition{Host: "host", CustomData: make([]byte, 1)}
	list.Add(service)
	service = list.Get(0)
	if service.CustomData == nil {
		t.Error("Data replacement failed")
	}
}

func TestServiceListRemove(t *testing.T) {
	var list serviceList
	list.Add(&serviceDefinition{})
	list.Remove(&serviceDefinition{})
	if list.Len() != 0 {
		t.Error("Empty definition failed")
	}

	list.Add(&serviceDefinition{Host: "host"})
	list.Remove(&serviceDefinition{Host: "host", group: "group"})
	if list.Len() != 1 {
		t.Error("Mismatched remove failed")
	}
}

func TestServiceListGet(t *testing.T) {
	var list serviceList
	if list.Get(0) != nil {
		t.Error("Empty list should return nil")
	}

	list.Add(&serviceDefinition{})
	if list.Get(0) == nil {
		t.Error("Single element list failed")
	}
	if list.Get(-1) != nil || list.Get(1) != nil {
		t.Error("Invalid index failed")
	}

	list.Add(&serviceDefinition{Host: "host"})
	if list.Get(1).Host != "host" {
		t.Error("Wrong definition returned")
	}
}

func TestServiceListIterator(t *testing.T) {
	var list serviceList
	list.Add(&serviceDefinition{Host: "host1", group: "group1"})
	list.Add(&serviceDefinition{Host: "host2", group: "group1"})
	list.Add(&serviceDefinition{Host: "host1", group: "group2"})
	list.Add(&serviceDefinition{Host: "host2", group: "group2"})
	list.Add(&serviceDefinition{Host: "host3", group: "group2"})

	iter := list.Iterator("group1")
	i := 0
	for iter.HasMore() {
		def := iter.Next()
		if def.Host != fmt.Sprintf("host%d", i+1) {
			t.Error("group1 wrong entry", i, def)
		}
		i++
	}
	if i != 2 {
		t.Error("Wrong number of entries in group1", i)
	}

	iter = list.Iterator("group2")
	i = 0
	for iter.HasMore() {
		def := iter.Next()
		if def.Host != fmt.Sprintf("host%d", i+1) {
			t.Error("group2 wrong entry", i, def)
		}
		i++
	}
	if i != 3 {
		t.Error("Wrong number of entries in group2", i)
	}

	if list.Iterator("group3").Next() != nil ||
		list.Iterator("group3").HasMore() {
		t.Error("group3 should not have entries")
	}
}

func TestServiceListRemoveAll(t *testing.T) {
	var list serviceList
	list.Add(&serviceDefinition{Host: "host1", connId: 0})
	list.Add(&serviceDefinition{Host: "host2", connId: 0})
	list.Add(&serviceDefinition{Host: "host3", connId: 1})
	list.Add(&serviceDefinition{Host: "host5", connId: 1})
	list.Add(&serviceDefinition{Host: "host4", connId: 2})
	list.Add(&serviceDefinition{Host: "host6", connId: 2})

	if list.Len() != 6 {
		t.Error("Wrong number of services")
	}
	list.RemoveAll(-1)
	if list.Len() != 6 {
		t.Error("Invalid connection id removed entries")
	}

	list.RemoveAll(0)
	if list.Len() != 4 {
		t.Error("Removing valid connection failed")
	}
	if list.Get(0).Host != "host3" || list.Get(1).Host != "host4" ||
		list.Get(2).Host != "host5" || list.Get(3).Host != "host6" {
		t.Error("Wrong entries after removing connId 0")
	}

	list.RemoveAll(2)
	if list.Len() != 2 {
		t.Error("Removing connId 2 failed")
	}
	if list.Get(0).Host != "host3" || list.Get(1).Host != "host5" {
		t.Error("Wrong entries after removing connId 2")
	}
}
