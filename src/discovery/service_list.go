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
	return fmt.Sprintf("'%s' %s:%d %v", s.group, s.Host, s.Port, s.CustomData)
}

func (l *serviceList) Add(service *serviceDefinition) {
	list := (*list.List)(l)
	for iter := list.Front(); iter != nil; iter = iter.Next() {
		e := iter.Value.(*serviceDefinition)
		res := service.compare(e)
		if res > 0 {
			continue
		} else if res < 0 {
			list.InsertBefore(service, iter)
		} else {
			iter.Value = service
		}
		return
	}
	list.PushBack(service)
}

func (l *serviceList) Remove(service *serviceDefinition) {
	list := (*list.List)(l)
	for iter := list.Front(); iter != nil; iter = iter.Next() {
		e := iter.Value.(*serviceDefinition)
		res := service.compare(e)
		if res < 0 {
			continue
		} else if res == 0 {
			list.Remove(iter)
		}
		break
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
	group   string
	hasMore bool // shortcut to avoid string comparisons
}

// Returns true if the iterator has more data.
func (i *iterator) HasMore() bool {
	if i.hasMore {
		return true
	}
	i.hasMore = false
	for ; i.iter != nil; i.iter = i.iter.Next() {
		def := i.iter.Value.(*serviceDefinition)
		diff := strcmp(def.group, i.group)
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
func (l *serviceList) Iterator(group string) *iterator {
	return &iterator{(*list.List)(l).Front(), group, false}
}
