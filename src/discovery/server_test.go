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

func TestJoinMessageCompare(t *testing.T) {
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
