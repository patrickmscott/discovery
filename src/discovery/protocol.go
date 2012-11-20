package discovery

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

type request struct {
	messageType MessageType
	message     Message
}

func (r *request) Type() MessageType {
	return r.messageType
}

func (r *request) ToJoin() *JoinMessage {
	if r.messageType != joinMessage {
		return nil
	}
	return r.message.(*JoinMessage)
}

func (r *request) ToLeave() *LeaveMessage {
	if r.messageType != leaveMessage {
		return nil
	}
	return r.message.(*LeaveMessage)
}

func (r *request) ToSnapshot() *SnapshotMessage {
	if r.messageType != snapshotMessage {
		return nil
	}
	return r.message.(*SnapshotMessage)
}

func (r *request) ToWatch() *WatchMessage {
	if r.messageType != watchMessage {
		return nil
	}
	return r.message.(*WatchMessage)
}

func readRequest(in io.Reader) (*request, error) {
	// Read 4 bytes for the size of the message.
	buffer := make([]byte, 4)
	err := fill(in, buffer)
	if err != nil {
		return nil, err
	}
	size := uint(buffer[0])<<24 |
		uint(buffer[1])<<16 |
		uint(buffer[2])<<8 |
		uint(buffer[3])
	if size > maxMessageSize {
		return nil, errors.New("Invalid message size, too large")
	}

	// Allocate a buffer of the message size and read fully.
	buffer = make([]byte, size)
	err = fill(in, buffer)
	if err != nil {
		return nil, err
	}

	var req request
	// The first byte should be a valid type.
	req.messageType = MessageType(buffer[0])
	if req.Type() < 0 || req.Type() >= lastMessageType {
		return &req, errors.New("Invalid message type")
	}

	dec := json.NewDecoder(bytes.NewBuffer(buffer[1:]))
	switch req.Type() {
	case joinMessage:
		req.message = &JoinMessage{}
	case leaveMessage:
		req.message = &LeaveMessage{}
	case snapshotMessage:
		req.message = &SnapshotMessage{}
	case watchMessage:
		req.message = &WatchMessage{}
	case heartbeatMessage:
		return &request{heartbeatMessage, nil}, nil
	}
	err = dec.Decode(req.message)
	return &req, err
}

func fill(in io.Reader, buffer []byte) error {
	total := 0
	for {
		n, err := in.Read(buffer[total:])
		if err != nil {
			return err
		}
		total += n
		if n == 0 || total == cap(buffer) {
			break
		}
	}
	return nil
}
