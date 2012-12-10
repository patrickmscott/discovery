package discovery

type JoinRequest struct {
	Host       string
	Port       uint16
	Group      string
	CustomData []byte
}

type LeaveRequest struct {
	Host  string
	Port  uint16
	Group string
}

// A snapshot only applies to a single group.
type SnapshotRequest struct {
	Group string
}

// A watch request contains a list of groups to watch on the connection.
type WatchRequest struct {
	Groups []string
}

type HeartbeatRequest struct{}
