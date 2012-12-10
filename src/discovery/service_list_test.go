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

func TestServiceListIterator(t *testing.T) {
	var list serviceList
	list.Add(&serviceDefinition{Host: "host1"})
	list.Add(&serviceDefinition{Host: "host2"})
	list.Add(&serviceDefinition{Host: "host3"})
	list.Add(&serviceDefinition{Host: "host4"})
	list.Add(&serviceDefinition{Host: "host5"})

	iter := list.Iterator()
	i := 0
	for {
		def := iter.Next()
		if def == nil {
			break
		}
		if def.Host != fmt.Sprintf("host%d", i+1) {
			t.Error("wrong entry", i, def)
		}
		i++
	}
	if i != 5 {
		t.Error("Wrong number of entries", i)
	}

	iter = list.Iterator()
	def := iter.Remove()
	if def.Host != "host1" {
		t.Error("Remove failed", def)
	}
	def = iter.Next()
	if def.Host != "host2" {
		t.Error("Next after remove failed", def)
	}
	iter.Remove()
	iter.Remove()
	def = iter.Next()
	if def.Host != "host5" {
		t.Error("Wrong entry after removing 2 entries", def)
	}
	if iter.Next() != nil {
		t.Error("Extra entries")
	}

	iter = list.Iterator()
	if iter.Next().Host != "host2" || iter.Next().Host != "host5" {
		t.Error("Leftover entries are incorrect")
	}
}
