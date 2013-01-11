package discovery

import (
	"bytes"
	"net"
	"net/rpc/jsonrpc"
	"testing"
)

func initDiscoveryTest(server *Server, id int32) *Discovery {
	read, _ := net.Pipe()
	disc := newDiscoveryService(server)
	disc.init(read, id)
	return disc
}

func TestDiscoveryJoin(t *testing.T) {
	server := NewServer()
	go server.processEvents()
	disc := initDiscoveryTest(server, 0)
	err := disc.Join(&ServiceDef{Host: "host", Group: "group"}, &Void{})
	if err != nil {
		t.Error(err)
	}

	if server.services.Len() != 1 {
		t.Error("Wrong number of services")
	}
	def := server.services.Get(0)
	if def.Host != "host" || def.CustomData != nil {
		t.Error("Wrong host entry", def)
	}

	custom := make([]byte, 1)
	custom[0] = 42
	err = disc.Join(
		&ServiceDef{Host: "host", Group: "group", CustomData: custom},
		&Void{})
	if err != nil {
		t.Error(err)
	}
	if server.services.Len() != 1 {
		t.Error("Wrong number of services")
	}
	def = server.services.Get(0)
	if def.Host != "host" || bytes.Compare(custom, def.CustomData) != 0 {
		t.Error("Wrong host entry", def)
	}

	disc = initDiscoveryTest(server, 1)
	err = disc.Join(&ServiceDef{Host: "host", Group: "group"}, &Void{})
	if err == nil {
		t.Error("Expected error adding host on diff connection")
	}
}

func TestDiscoveryLeave(t *testing.T) {
	server := NewServer()
	go server.processEvents()
	disc := initDiscoveryTest(server, 0)

	err := disc.Leave(&ServiceDef{Host: "host"}, &Void{})
	if err == nil {
		t.Error("Expected error removing undefined host")
	}

	server.services.Add(&ServiceDef{Host: "host"})
	err = disc.Leave(&ServiceDef{Host: "host", Group: "group"}, &Void{})
	if err == nil {
		t.Error("Expected error removing different group")
	}
	if server.services.Len() != 1 {
		t.Error("Invalid leave removed service entry")
	}

	err = disc.Leave(&ServiceDef{Host: "host"}, &Void{})
	if err != nil {
		t.Error(err)
	}
	if server.services.Len() != 0 {
		t.Error("Services is not empty")
	}

	disc = initDiscoveryTest(server, 1)
	// Default connection id is 0
	server.services.Add(&ServiceDef{Host: "host"})
	err = disc.Leave(&ServiceDef{Host: "host"}, &Void{})
	if err == nil {
		t.Error("Expected error from different connection")
	}
}

func TestDiscoverySnapshot(t *testing.T) {
	server := NewServer()
	go server.processEvents()
	disc := initDiscoveryTest(server, 0)

	server.services.Add(&ServiceDef{Host: "host1", Port: 1, Group: "a"})
	server.services.Add(&ServiceDef{Host: "host2", Port: 2, Group: "a"})
	server.services.Add(&ServiceDef{Host: "host3", Port: 1, Group: "b"})

	var snapshot []*ServiceDef
	err := disc.Snapshot("a", &snapshot)
	if err != nil {
		t.Error(err)
	}
	if len(snapshot) != 2 {
		t.Error("Snapshot length incorrect")
	}
	if def := snapshot[0]; def.Host != "host1" || def.Port != 1 {
		t.Error("Incorrect service", def)
	}
	if def := snapshot[1]; def.Host != "host2" || def.Port != 2 {
		t.Error("Incorrect service", def)
	}

	err = disc.Snapshot("b", &snapshot)
	if err != nil {
		t.Error(err)
	}
	if len(snapshot) != 1 {
		t.Error("Snapshot length incorrect")
	}
	if def := snapshot[0]; def.Host != "host3" || def.Port != 1 {
		t.Error("Incorrect service", def)
	}

	err = disc.Snapshot("c", &snapshot)
	if err != nil {
		t.Error(err)
	}
	if len(snapshot) != 0 {
		t.Error("Snapshot should be empty")
	}
}

func TestDiscoveryWatch(t *testing.T) {
	server := NewServer()
	go server.processEvents()
	disc := initDiscoveryTest(server, 0)

	err := disc.Watch("group", &Void{})
	if err == nil {
		t.Error("Watching without a client should fail")
	}
	if disc.client != nil {
		t.Error()
	}
	_, ok := server.watchers["group"]
	if ok {
		t.Error("Watcher should not be added")
	}

	_, write := net.Pipe()
	disc.client = jsonrpc.NewClient(write)

	err = disc.Watch("group", &Void{})
	if err != nil {
		t.Error(err)
	}

	if !server.watchers["group"][disc.client] {
		t.Error("Watcher not added")
	}
}

func TestDiscoveryIgnore(t *testing.T) {
	server := NewServer()
	go server.processEvents()
	disc := initDiscoveryTest(server, 0)

	_, write := net.Pipe()
	disc.client = jsonrpc.NewClient(write)
	disc.Watch("group", &Void{})
	if server.watchers["group"] == nil || !server.watchers["group"][disc.client] {
		t.Error("Watcher not registered")
	}

	disc.Ignore("diff_group", &Void{})
	if server.watchers["group"] == nil || !server.watchers["group"][disc.client] {
		t.Error("Watcher removed")
	}

	otherClient := initDiscoveryTest(server, 1)
	otherClient.client = jsonrpc.NewClient(write)
	otherClient.Watch("group", &Void{})
	if len(server.watchers["group"]) != 2 {
		t.Error("Wrong watcher count")
	}

	disc.Ignore("group", &Void{})
	if len(server.watchers["group"]) != 1 {
		t.Error("Watcher was not removed")
	}

	otherClient.Ignore("group", &Void{})
	if server.watchers["group"] != nil {
		t.Error("Watcher group not deleted")
	}
}
