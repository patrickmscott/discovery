package discovery

import (
	"container/list"
	"fmt"
)

type serviceDefinition struct {
	Host       string `json:"host"`
	Port       uint16 `json:"port"`
	CustomData []byte `json:"customData,omitempty"`
	group      string
	connId     int32
}

type serviceList list.List

func (a *serviceDefinition) compare(b *serviceDefinition) int {
	res := strcmp(a.group, b.group)
	if res == 0 {
		res = strcmp(a.Host, b.Host)
	}
	if res == 0 {
		res = int(a.Port) - int(b.Port)
	}
	return res
}

func (s *serviceDefinition) String() string {
	return fmt.Sprintf("'%s' %s:%d %v conn#%d",
		s.group, s.Host, s.Port, s.CustomData, s.connId)
}

// Add a new service definition to the list. If the definition is added or
// updated, return true.
func (l *serviceList) Add(service *serviceDefinition) bool {
	list := (*list.List)(l)
	for iter := list.Front(); iter != nil; iter = iter.Next() {
		e := iter.Value.(*serviceDefinition)
		res := service.compare(e)
		if res > 0 {
			continue
		} else if res < 0 {
			list.InsertBefore(service, iter)
			return true
		} else if e.connId == service.connId {
			// Replace the definition if it is from the same connection.
			iter.Value = service
			return true
		}
		// Equal entries but from a different connection.
		return false
	}
	list.PushBack(service)
	return true
}

// Remove a service definition from the list. If a service has been removed,
// return true. Different connections cannot remove services they did not add.
func (l *serviceList) Remove(service *serviceDefinition) bool {
	list := (*list.List)(l)
	for iter := list.Front(); iter != nil; iter = iter.Next() {
		e := iter.Value.(*serviceDefinition)
		res := service.compare(e)
		if res < 0 {
			continue
		} else if res == 0 && e.connId == service.connId {
			list.Remove(iter)
			return true
		}
		// Did not find the service.
		break
	}
	return false
}

func (l *serviceList) RemoveAll(connId int32) {
	list := (*list.List)(l)
	for iter := list.Front(); iter != nil; {
		e := iter.Value.(*serviceDefinition)
		next := iter.Next()
		if e.connId == connId {
			list.Remove(iter)
		}
		iter = next
	}
}

func (l *serviceList) Get(index int) *serviceDefinition {
	if index < 0 || index >= l.Len() {
		return nil
	}
	for iter := (*list.List)(l).Front(); iter != nil; iter = iter.Next() {
		if index == 0 {
			return iter.Value.(*serviceDefinition)
		}
		index--
	}
	// Impossible to reach.
	panic("Unreachable")
}

func (l *serviceList) Len() int { return (*list.List)(l).Len() }

type iterator struct {
	iter    *list.Element
	cmp     func(service *serviceDefinition) int
	hasMore bool // shortcut to avoid calling cmp
}

// Returns true if the iterator has more data.
func (i *iterator) HasMore() bool {
	if i.hasMore {
		return true
	}
	i.hasMore = false
	for ; i.iter != nil; i.iter = i.iter.Next() {
		diff := i.cmp(i.iter.Value.(*serviceDefinition))
		if diff == 0 {
			i.hasMore = true
			break
		} else if diff > 0 {
			i.iter = nil
			break
		}
	}
	return i.hasMore
}

// Returns the next *serviceDefinition in the service group. Nil if this
// iterator has no more data.
func (i *iterator) Next() *serviceDefinition {
	if !i.HasMore() {
		return nil
	}
	def := i.iter.Value.(*serviceDefinition)
	i.iter = i.iter.Next()
	i.hasMore = false // more like unknown
	return def
}

// Create an interator over the serviceDefinitions in a service group. Never
// returns nil.
func (l *serviceList) GroupIterator(group string) *iterator {
	return &iterator{
		(*list.List)(l).Front(),
		func(service *serviceDefinition) int {
			return strcmp(service.group, group)
		},
		false}
}

func (l *serviceList) ConnIterator(id int32) *iterator {
	return &iterator{
		(*list.List)(l).Front(),
		func(service *serviceDefinition) int {
			if service.connId == id {
				return 0
			}
			return -1
		},
		false}
}
