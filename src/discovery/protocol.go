package discovery

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
)

type MessageType byte

const (
	JOIN = iota
	LEAVE
	WATCH
	SNAPSHOT
	HEARTBEAT
	lastMessageType
	maxMessageSize = 1 * 1024 * 1024
)

type Protocol struct {
	Type      MessageType
	IpAddress net.IP
	message		*bytes.Buffer
}

func (p *Protocol) SetIpFromAddr(addr net.Addr) error {
	tcpAddr := addr.(*net.TCPAddr)
	p.IpAddress = tcpAddr.IP.To4()
	if tcpAddr.IP.IsLoopback() {
		p.IpAddress = net.IPv4(127, 0, 0, 1)
	}
	if p.IpAddress == nil {
		return errors.New(fmt.Sprintln("Invalid IpAddress from", addr))
	}
	return nil
}

func (p *Protocol) ReadRequest(in io.Reader) error {
	// Read 4 bytes for the size of the message.
	buffer := make([]byte, 4)
	err := fill(in, buffer)
	if err != nil {
		return err
	}
	size := uint(buffer[0])<<24 |
		uint(buffer[1])<<16 |
		uint(buffer[2])<<8 |
		uint(buffer[3])
	if size > maxMessageSize {
		return errors.New("Invalid message size, too large")
	}

	// Allocate a buffer of the message size and read fully.
	buffer = make([]byte, size)
	err = fill(in, buffer)
	if err != nil {
		return err
	}

	// The first byte should be a valid type.
	p.Type = MessageType(buffer[0])
	if p.Type < 0 || p.Type >= lastMessageType {
		return errors.New("Invalid message type")
	}

	p.message = bytes.NewBuffer(buffer[1:])
	return nil
}

func (p *Protocol) ReadInt() (int, error) {
	if p.message == nil || p.message.Len() < 4 {
		return 0, errors.New("Invalid message")
	}
	buffer := p.message.Next(4)
	return int(uint(buffer[0])<<24 |
		uint(buffer[1])<<16 |
		uint(buffer[2])<<8 |
		uint(buffer[3])), nil
}

func (p *Protocol) ReadShort() (int16, error) {
	if p.message == nil || p.message.Len() < 2 {
		return 0, errors.New("Invalid message")
	}
	buffer := p.message.Next(2)
	return int16(uint16(buffer[0])<<8 | uint16(buffer[1])), nil
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
