package discovery

import (
	"bytes"
	"encoding/json"
	"errors"
	"hash/crc32"
	"io"
)

type request struct {
	messageType MessageType
	message     interface{}
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

type Protocol struct {
	buffer [512]byte
}

// Read an integer value as 4 big-endian bytes from in.
func (p *Protocol) readInt(in io.Reader) (int, error) {
	buffer := p.buffer[0:4]
	n, err := in.Read(buffer)
	if err != nil {
		return 0, err
	}
	if n != len(buffer) {
		return 0, io.EOF
	}
	return int(uint(buffer[0])<<24 |
		uint(buffer[1])<<16 |
		uint(buffer[2])<<8 |
		uint(buffer[3])), nil
}

// Write the integer value as a sequence of 4 big-endian bytes to out.
func (p *Protocol) writeInt(out io.Writer, i int) error {
	p.buffer[0] = byte(i >> 24)
	p.buffer[1] = byte(i >> 16)
	p.buffer[2] = byte(i >> 8)
	p.buffer[3] = byte(i)
	_, err := out.Write(p.buffer[0:4])
	return err
}

const magicNumber int = 1412959279

var ErrMagicNumber = errors.New("Invalid magic number")
var ErrChecksum = errors.New("Invalid checksum")
var ErrMessageSize = errors.New("Invalid message size")
var ErrMessageType = errors.New("Invalid message type")

func (p *Protocol) readRequest(in io.Reader) (*request, error) {
	// Read the 4 byte magic number
	magic, err := p.readInt(in)
	if err != nil || magic != magicNumber {
		return nil, ErrMagicNumber
	}

	// Read the 4 byte checksum.
	checksum, err := p.readInt(in)
	if err != nil {
		return nil, ErrChecksum
	}

	// Read 4 bytes for the size of the message.
	size, err := p.readInt(in)
	if err != nil || size > maxMessageSize {
		return nil, ErrMessageSize
	}

	// Allocate a buffer if needed and read fully.
	buffer := p.buffer[0:]
	if size > len(buffer) {
		buffer = make([]byte, size)
	}
	offset := 0
	for {
		n, err := in.Read(buffer[offset:])
		offset += n
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	if offset != size {
		return nil, io.EOF
	}

	// Verify the checksum is correct.
	if checksum != int(crc32.ChecksumIEEE(buffer[0:size])) {
		return nil, ErrChecksum
	}

	var req request
	// The first byte should be a valid type.
	req.messageType = MessageType(buffer[0])
	if req.Type() < 0 || req.Type() >= lastMessageType {
		return &req, ErrMessageType
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

func (p *Protocol) writeSnapshot(
	out io.Writer, snapshot []ServiceDefinition) error {
	buffer := bytes.NewBuffer(p.buffer[0:0])
	if snapshot != nil {
		enc := json.NewEncoder(buffer)
		err := enc.Encode(snapshot)
		if err != nil {
			return err
		}
	}
	var length [4]byte
	length[0] = byte(buffer.Len() >> 24)
	length[1] = byte(buffer.Len() >> 16)
	length[2] = byte(buffer.Len() >> 8)
	length[3] = byte(buffer.Len())
	_, err := out.Write(length[0:])
	if err != nil {
		return err
	}
	if buffer.Len() != 0 {
		_, err = out.Write(buffer.Bytes())
	}
	return err
}
