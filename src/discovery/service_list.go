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
	list *list.List
	iter *list.Element
}

// Returns the current *serviceDefinition and increments the iterator.
func (i *iterator) Next() *serviceDefinition {
	if i.iter == nil {
		return nil
	}

	service := i.iter.Value.(*serviceDefinition)
	i.iter = i.iter.Next()
	return service
}

// Removes the current value and increments the iterator.
func (i *iterator) Remove() *serviceDefinition {
	if i.iter == nil {
		return nil
	}

	service := i.iter.Value.(*serviceDefinition)
	next := i.iter.Next()
	i.list.Remove(i.iter)
	i.iter = next
	return service
}

// Create a simple iterator over all services.
func (l *serviceList) Iterator() *iterator {
	ll := (*list.List)(l)
	return &iterator{ll, ll.Front()}
}
