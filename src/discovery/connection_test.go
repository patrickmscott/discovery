package discovery

import (
	"testing"
)

func TestConnectionJoin(t *testing.T) {
	conn := &connection{server: &Server{}}
	conn.server.eventChan = make(chan func(), 2)

	conn.Join(&JoinRequest{Host: "host"})
	conn.Join(&JoinRequest{Host: "host1"})

	(<-conn.server.eventChan)()
	if conn.server.services.Len() != 1 {
		t.Error("Join request failed to add service")
	}
	service := conn.server.services.Get(0)
	if service.Host != "host" {
		t.Error("First service incorrect", service)
	}

	(<-conn.server.eventChan)()
	if conn.server.services.Len() != 2 {
		t.Error("Second join request failed to add service")
	}
	service = conn.server.services.Get(1)
	if service.Host != "host1" {
		t.Error("Second service incorrect", service)
	}
}

func TestConnectionLeave(t *testing.T) {
	conn := &connection{server: &Server{}}
	conn.server.eventChan = make(chan func(), 3)

	conn.Join(&JoinRequest{Host: "host"})
	conn.Join(&JoinRequest{Host: "host1"})
	conn.Leave(&LeaveRequest{Host: "host"})

	(<-conn.server.eventChan)()
	(<-conn.server.eventChan)()
	(<-conn.server.eventChan)()

	if conn.server.services.Len() != 1 {
		t.Error("Wrong number of services")
	}
	service := conn.server.services.Get(0)
	if service.Host != "host1" {
		t.Error("Wrong service after leave event", service)
	}

	conn.Leave(&LeaveRequest{Group: "group"})
	(<-conn.server.eventChan)()

	if conn.server.services.Len() != 1 {
		t.Error("Wrong number of services after invalid leave")
	}
	service = conn.server.services.Get(0)
	if service.Host != "host1" || service.group != "" {
		t.Error("Wrong service", service)
	}
}
