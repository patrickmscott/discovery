package discovery

import (
	"bytes"
	"encoding/json"
	"errors"
	"hash/crc32"
	"io"
)

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
	// Not using the internal scratch space as we could be using it to write a
	// snapshot.
	var buffer [4]byte
	buffer[0] = byte(i >> 24)
	buffer[1] = byte(i >> 16)
	buffer[2] = byte(i >> 8)
	buffer[3] = byte(i)
	_, err := out.Write(buffer[0:])
	return err
}

const (
	magicNumber    = 1412959279
	maxRequestSize = 1 * 1024 * 1024
)

var ErrMagicNumber = errors.New("Invalid magic number")
var ErrChecksum = errors.New("Invalid checksum")
var ErrRequestSize = errors.New("Invalid request size")
var ErrRequestType = errors.New("Invalid request type")

func (p *Protocol) readRequest(in io.Reader) (Request, error) {
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

	// Read 4 bytes for the size of the request.
	size, err := p.readInt(in)
	if err != nil || size > maxRequestSize {
		return nil, ErrRequestSize
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

	// The first byte should be a valid type.
	requestType := RequestType(buffer[0])
	if requestType < 0 || requestType >= lastRequestType {
		return nil, ErrRequestType
	}

	var req Request
	dec := json.NewDecoder(bytes.NewBuffer(buffer[1:]))
	switch requestType {
	case joinRequest:
		req = &JoinRequest{}
	case leaveRequest:
		req = &LeaveRequest{}
	case snapshotRequest:
		req = &SnapshotRequest{}
	case watchRequest:
		req = &WatchRequest{}
	case heartbeatRequest:
		return &HeartbeatRequest{}, nil
	}
	err = dec.Decode(req)
	return req, err
}

func (p *Protocol) writeJson(out io.Writer, obj interface{}) error {
	buffer := bytes.NewBuffer(p.buffer[0:0])
	if obj != nil {
		enc := json.NewEncoder(buffer)
		err := enc.Encode(obj)
		if err != nil {
			return err
		}
	}
	return p.writeBytes(out, buffer.Bytes())
}

func (p *Protocol) writeRequest(out io.Writer, request Request) error {
	buffer := bytes.NewBuffer(p.buffer[0:0])
	buffer.WriteByte(byte(request.Type()))
	enc := json.NewEncoder(buffer)
	err := enc.Encode(request)
	if err != nil {
		return err
	}
	return p.writeBytes(out, buffer.Bytes())
}

func (p *Protocol) writeBytes(out io.Writer, bytes []byte) error {
	if err := p.writeInt(out, magicNumber); err != nil {
		return err
	}
	checksum := int(crc32.ChecksumIEEE(bytes))
	if err := p.writeInt(out, checksum); err != nil {
		return err
	}
	if err := p.writeInt(out, len(bytes)); err != nil {
		return err
	}
	if len(bytes) != 0 {
		n, err := out.Write(bytes)
		if err != nil {
			return err
		}
		if n != len(bytes) {
			return io.EOF
		}
	}
	return nil
}
