package discovery

import "testing"

func TestServiceDefCompare(t *testing.T) {
	var a, b ServiceDef
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
	a.Group = "group1"
	b.Group = "group2"
	if a.compare(&b) >= 0 {
		t.Error("a should be less than b: group")
	}

	a.Group = "group"
	b.Group = "group"
	a.Port = 1
	b.Port = 2
	if a.compare(&b) >= 0 {
		t.Error("a should be less than b: port")
	}
}
