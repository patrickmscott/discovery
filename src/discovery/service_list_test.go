package discovery

import (
	"fmt"
	"testing"
)

func TestServiceListAdd(t *testing.T) {
	var list serviceList
	service := &ServiceDef{Host: "host"}

	if list.Len() != 0 {
		t.Error("Empty list should have 0 length")
	}

	if !list.Add(service) || list.Len() != 1 {
		t.Error("Single add failed")
	}

	service = &ServiceDef{Host: "host1"}
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

	service = &ServiceDef{Host: "host", CustomData: make([]byte, 1)}
	if !list.Add(service) || list.Get(0).CustomData == nil {
		t.Error("Data replacement failed")
	}

	if list.Add(&ServiceDef{Host: "host", connId: 2}) {
		t.Error("Cannot replace different connection data")
	}
}

func TestServiceListRemove(t *testing.T) {
	var list serviceList
	list.Add(&ServiceDef{})
	if !list.Remove(&ServiceDef{}) || list.Len() != 0 {
		t.Error("Empty definition failed")
	}

	list.Add(&ServiceDef{Host: "host"})
	if list.Remove(&ServiceDef{Host: "host", Group: "group"}) {
		t.Error("Removing unknown service failed")
	}
	if list.Len() != 1 {
		t.Error("Mismatched remove failed")
	}
	if list.Remove(&ServiceDef{Host: "host", connId: 1}) {
		t.Error("Removing from a different connection should fail")
	}
}

func TestServiceListGet(t *testing.T) {
	var list serviceList
	if list.Get(0) != nil {
		t.Error("Empty list should return nil")
	}

	list.Add(&ServiceDef{})
	if list.Get(0) == nil {
		t.Error("Single element list failed")
	}
	if list.Get(-1) != nil || list.Get(1) != nil {
		t.Error("Invalid index failed")
	}

	list.Add(&ServiceDef{Host: "host"})
	if list.Get(1).Host != "host" {
		t.Error("Wrong definition returned")
	}
}

func TestServiceListIterator(t *testing.T) {
	var list serviceList
	list.Add(&ServiceDef{Host: "host1"})
	list.Add(&ServiceDef{Host: "host2"})
	list.Add(&ServiceDef{Host: "host3"})
	list.Add(&ServiceDef{Host: "host4"})
	list.Add(&ServiceDef{Host: "host5"})

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
	if iter.Remove() != nil {
		t.Error("Remove before Next() failed")
	}
	def := iter.Next()
	if def.Host != iter.Remove().Host {
		t.Error("Remove deleted the wrong entry")
	}
	def = iter.Next()
	if def.Host != "host2" {
		t.Error("Next after remove failed", def)
	}
	iter.Next()
	iter.Remove()
	iter.Next()
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
