package discovery

import "fmt"

// A ServiceDef is used when a service joins or leaves a group, an event is sent
// to watchers of a group, or when a group snapshot is requested.
type ServiceDef struct {
	Host  string `json:"host,omitempty"`
	Port  uint16 `json:"port"`
	Group string `json:"group"`
	// CustomData need not be present when a client calls Discovery.Leave.
	CustomData []byte `json:"custom_data,omitempty"`

	// Used internally to denote which connection the service is attached.
	connId int32
}

// Compare this service definition with b. Services are ordered by group, then
// by host, and finally by port.
// a < b  return < 0
// a == b return 0
// a > b  return > 0
func (a *ServiceDef) compare(b *ServiceDef) int {
	res := strcmp(a.Group, b.Group)
	if res == 0 {
		res = strcmp(a.Host, b.Host)
	}
	if res == 0 {
		res = int(a.Port) - int(b.Port)
	}
	return res
}

// Internal toString method that includes the connection number. Not exposed
// since go clients do not need to see a connection number.
func (def *ServiceDef) toString() string {
	return fmt.Sprintf("'%s' %s:%d conn#%d",
		def.Group, def.Host, def.Port, def.connId)
}

func (def *ServiceDef) String() string {
	return fmt.Sprintf("'%s' %s:%d %v",
		def.Group, def.Host, def.Port, def.CustomData)
}
