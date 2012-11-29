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
	service := &serviceDefinition{"host", 8080, nil, "group"}

	if list.Len() != 0 {
		t.Error("Empty list should have 0 length")
	}

	list.Add(service)
	if list.Len() != 1 {
		t.Error("Single add failed")
	}

	service = &serviceDefinition{"host1", 8080, nil, "group"}
	list.Add(service)
	if list.Len() != 2 {
		t.Error("Multiple add failed")
	}

	list.Add(service)
	if list.Len() != 2 {
		t.Error("Duplicate add failed")
	}

	service = list.Get(0)
	if service.Host != "host" || service.Port != 8080 ||
		service.group != "group" || service.CustomData != nil {
		t.Error("First entry invalid", service)
	}

	service = list.Get(1)
	if service.Host != "host1" || service.Port != 8080 ||
		service.group != "group" || service.CustomData != nil {
		t.Error("Second entry invalid", service)
	}

	service = &serviceDefinition{"host", 8080, make([]byte, 1), "group"}
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
	list.Add(&serviceDefinition{"host1", 0, nil, "group1"})
	list.Add(&serviceDefinition{"host2", 0, nil, "group1"})
	list.Add(&serviceDefinition{"host1", 0, nil, "group2"})
	list.Add(&serviceDefinition{"host2", 0, nil, "group2"})
	list.Add(&serviceDefinition{"host3", 0, nil, "group2"})

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