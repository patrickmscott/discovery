package discovery

import "container/list"

type serviceList list.List

// Add a new service definition to the list. If the definition is added or
// updated, return true.
func (l *serviceList) Add(service *ServiceDef) bool {
	list := (*list.List)(l)
	for iter := list.Front(); iter != nil; iter = iter.Next() {
		e := iter.Value.(*ServiceDef)
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
func (l *serviceList) Remove(service *ServiceDef) bool {
	list := (*list.List)(l)
	for iter := list.Front(); iter != nil; iter = iter.Next() {
		e := iter.Value.(*ServiceDef)
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

func (l *serviceList) Get(index int) *ServiceDef {
	if index < 0 || index >= l.Len() {
		return nil
	}
	for iter := (*list.List)(l).Front(); iter != nil; iter = iter.Next() {
		if index == 0 {
			return iter.Value.(*ServiceDef)
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
	curr *list.Element
}

// Returns the current *ServiceDef and increments the iterator.
func (i *iterator) Next() *ServiceDef {
	if i.iter == nil {
		return nil
	}

	i.curr = i.iter
	service := i.iter.Value.(*ServiceDef)
	i.iter = i.iter.Next()
	return service
}

// Removes the value returned by Next(). Does not increment the iterator.
func (i *iterator) Remove() *ServiceDef {
	if i.curr == nil {
		return nil
	}

	service := i.curr.Value.(*ServiceDef)
	i.list.Remove(i.curr)
	return service
}

// Create a simple iterator over all services.
func (l *serviceList) Iterator() *iterator {
	ll := (*list.List)(l)
	return &iterator{list: ll, iter: ll.Front()}
}
