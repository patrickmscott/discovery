package discovery

type MessageType byte

const (
	joinMessage MessageType = iota
	leaveMessage
	watchMessage
	snapshotMessage
	heartbeatMessage
	lastMessageType
	maxMessageSize = 1 * 1024 * 1024
)

type JoinMessage struct {
	Host       string
	Port       uint16
	Group      string
	CustomData []byte
}

// Returns a shallow copy of the join message.
func (join *JoinMessage) Copy() *JoinMessage {
	return &JoinMessage{join.Host, join.Port, join.Group, join.CustomData}
}

type LeaveMessage struct {
	Host  string
	Port  uint16
	Group string
}

// A snapshot only applies to a single group.
type SnapshotMessage struct {
	Group string
}

// A watch message contains a list of groups to watch on the connection.
type WatchMessage struct {
	Groups []string
}

type HeartbeatMessage struct{}
