package discovery

import (
	"bytes"
	"container/list"
)

type serviceDefinition struct {
	Host       string `json:"host"`
	Port       uint16 `json:"port"`
	CustomData []byte `json:"customData,omitempty"`
	group      string
}

type serviceList list.List

func (a *serviceDefinition) compare(b *serviceDefinition) int {
	res := bytes.Compare([]byte(a.group), []byte(b.group))
	if res == 0 {
		res = bytes.Compare([]byte(a.Host), []byte(b.Host))
	}
	if res == 0 {
		res = int(a.Port) - int(b.Port)
	}
	return res
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
