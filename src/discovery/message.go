package discovery

// Basic struct containing the host ip and port.
type HostAndPort struct {
	Host string
	Port uint8
}

type JoinMessage struct {
	// The address information for connecting to the published service
	Address HostAndPort

	// The service group that is being joined. This can be an arbitrary string.
	// Services added to this group will be published to watchers of the group.
	Group string

	// Custom data sent by the publisher and interpreted by the watcher.
	CustomData []byte
}

type LeaveMessage struct {
	// The address of the service that is leaving the group.
	Address HostAndPort

	// The group name being watched. This is helpful when watching multiple groups
	// on the same connection.
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
