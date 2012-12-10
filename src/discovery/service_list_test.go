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

	if !list.Add(service) || list.Len() != 1 {
		t.Error("Single add failed")
	}

	service = &serviceDefinition{Host: "host1"}
	if !list.Add(service) || list.Len() != 2 {
		t.Error("Multiple add failed")
	}

	// Replaces CustomData.
	if !list.Add(service) || list.Len() != 2 {
		t.Error("Duplicate add failed", list.Len())
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
	if !list.Add(service) || list.Get(0).CustomData == nil {
		t.Error("Data replacement failed")
	}

	if list.Add(&serviceDefinition{Host: "host", connId: 2}) {
		t.Error("Cannot replace different connection data")
	}
}

func TestServiceListRemove(t *testing.T) {
	var list serviceList
	list.Add(&serviceDefinition{})
	if !list.Remove(&serviceDefinition{}) || list.Len() != 0 {
		t.Error("Empty definition failed")
	}

	list.Add(&serviceDefinition{Host: "host"})
	if list.Remove(&serviceDefinition{Host: "host", group: "group"}) {
		t.Error("Removing unknown service failed")
	}
	if list.Len() != 1 {
		t.Error("Mismatched remove failed")
	}
	if list.Remove(&serviceDefinition{Host: "host", connId: 1}) {
		t.Error("Removing from a different connection should fail")
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

func TestServiceListGroupIterator(t *testing.T) {
	var list serviceList
	list.Add(&serviceDefinition{Host: "host1", group: "group1"})
	list.Add(&serviceDefinition{Host: "host2", group: "group1"})
	list.Add(&serviceDefinition{Host: "host1", group: "group2"})
	list.Add(&serviceDefinition{Host: "host2", group: "group2"})
	list.Add(&serviceDefinition{Host: "host3", group: "group2"})

	iter := list.GroupIterator("group1")
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

	iter = list.GroupIterator("group2")
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

	if list.GroupIterator("group3").Next() != nil ||
		list.GroupIterator("group3").HasMore() {
		t.Error("group3 should not have entries")
	}
}

func TestServiceListConnIterator(t *testing.T) {
	var list serviceList
	list.Add(&serviceDefinition{Host: "h0", connId: 0})
	list.Add(&serviceDefinition{Host: "h1", connId: 1})
	list.Add(&serviceDefinition{Host: "h2", connId: 2})
	list.Add(&serviceDefinition{Host: "h4", connId: 0})
	list.Add(&serviceDefinition{Host: "h5", connId: 1})
	list.Add(&serviceDefinition{Host: "h6", connId: 2})

	var i int32
	for i = 0; i < 3; i++ {
		iter := list.ConnIterator(i)
		if !iter.HasMore() {
			t.Error("Expected connection to have a valid iterator", i)
		}
		service := iter.Next()
		host := fmt.Sprintf("h%d", i)
		if service.Host != host || service.connId != i {
			t.Error("Wrong first entry in iterator", i, service)
		}
		if !iter.HasMore() {
			t.Error("Expected connection to have more entries", i)
		}
		service = iter.Next()
		host = fmt.Sprintf("h%d", i+4)
		if service.Host != host || service.connId != i {
			t.Error("Wrong second entry in iterator", i, service)
		}
		if iter.HasMore() {
			t.Error("Too many entries in iterator", i)
		}
	}

	if list.ConnIterator(3).HasMore() {
		t.Error("Invalid connection has a valid iterator")
	}
}
