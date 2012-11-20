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

type Message interface{}

type JoinMessage struct {
	Message
	Port       uint16
	Group      string
	CustomData []byte
}

type LeaveMessage struct {
	Message
	Port  uint16
	Group string
}

// A snapshot only applies to a single group.
type SnapshotMessage struct {
	Message
	Group string
}

// A watch message contains a list of groups to watch on the connection.
type WatchMessage struct {
	Message
	Groups []string
}

type HeartbeatMessage struct {
	Message
}
