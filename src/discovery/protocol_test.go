package discovery

import (
	"bytes"
	"math"
	"net"
	"testing"
)

func TestSetIpFromAddr(t *testing.T) {
	var proto Protocol
	err := proto.SetIpFromAddr(&net.TCPAddr{})
	if err == nil {
		t.Error()
	}

	var addr net.TCPAddr
	addr.IP = net.ParseIP("2001:0db8:85a3:0042:0000:8a2e:0370:7334")
	err = proto.SetIpFromAddr(&addr)
	if err == nil {
		t.Error()
	}

	addr.IP = net.IPv6loopback
	err = proto.SetIpFromAddr(&addr)
	if err != nil {
		t.Error(err)
	}
	if !proto.IpAddress.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Error("Loopback address failed", proto.IpAddress)
	}

	addr.IP = net.ParseIP("134.22.31.43")
	err = proto.SetIpFromAddr(&addr)
	if err != nil {
		t.Error(err)
	}
	if !proto.IpAddress.Equal(net.IPv4(134, 22, 31, 43)) {
		t.Error("Failed to parse ip address", proto.IpAddress)
	}
}

func TestReadInt(t *testing.T) {
	var proto Protocol
	buffer := make([]byte, 6)
	buffer[0] = 0x00
	buffer[1] = 0x01
	buffer[2] = 0x10
	buffer[3] = 0x11
	buffer[4] = 0xfe
	buffer[5] = 0x03
	proto.message = bytes.NewBuffer(buffer)

	i, err := proto.ReadInt()
	if err != nil {
		t.Error(err)
	}
	if i != 0x00011011 {
		t.Error("Not the right int value", i)
	}

	s, err := proto.ReadShort()
	if err != nil {
		t.Error(err)
	}
	if s != -509 {
		t.Error("Not the right short value", s)
	}

	proto.message = nil
	i, err = proto.ReadInt()
	if err == nil {
		t.Error()
	}
	s, err = proto.ReadShort()
	if err == nil {
		t.Error()
	}

	proto.message = bytes.NewBuffer(make([]byte, 2))
	i, err = proto.ReadInt()
	if err == nil || err.Error() != "Invalid message" {
		t.Error()
	}

	proto.message = &bytes.Buffer{}
	s, err = proto.ReadShort()
	if err == nil || err.Error() != "Invalid message" {
		t.Error()
	}
}

func TestBadSizeLength(t *testing.T) {
	var proto Protocol
	err := proto.ReadRequest(&bytes.Buffer{})
	if err == nil {
		t.Error()
	}

	var buffer bytes.Buffer
	buffer.WriteByte(0)
	err = proto.ReadRequest(&buffer)
	if err == nil {
		t.Error()
	}
}

func writeLength(buffer *bytes.Buffer, length uint) {
	buffer.WriteByte(byte(length >> 24))
	buffer.WriteByte(byte(length >> 16))
	buffer.WriteByte(byte(length >> 8))
	buffer.WriteByte(byte(length))
}

func TestInvalidMessageSize(t *testing.T) {
	var proto Protocol
	var buffer bytes.Buffer
	writeLength(&buffer, 2*1024*1024)
	err := proto.ReadRequest(&buffer)
	if err == nil {
		t.Error()
	}
	if err.Error() != "Invalid message size, too large" {
		t.Error("Wrong error")
	}
}

func TestMessageTooShort(t *testing.T) {
	var proto Protocol
	var buffer bytes.Buffer
	writeLength(&buffer, 20)
	err := proto.ReadRequest(&buffer)
	if err == nil {
		t.Error()
	}

	buffer.Reset()
	writeLength(&buffer, 20)
	buffer.WriteString("data")
	err = proto.ReadRequest(&buffer)
	if err == nil {
		t.Error()
	}
}

func TestMessageTypes(t *testing.T) {
	var proto Protocol
	var buffer bytes.Buffer

	for _, r := range [...]MessageType{JOIN, LEAVE, WATCH, SNAPSHOT, HEARTBEAT} {
		buffer.Reset()
		writeLength(&buffer, 1)
		buffer.WriteByte(byte(r))
		err := proto.ReadRequest(&buffer)
		if err != nil {
			t.Error(err)
		}
		if proto.Type != r {
			t.Error("Type mismatch", r, proto.Type)
		}
	}

	for i := lastMessageType; i < math.MaxUint8; i++ {
		buffer.Reset()
		writeLength(&buffer, 1)
		buffer.WriteByte(byte(i))
		err := proto.ReadRequest(&buffer)
		if err == nil || err.Error() != "Invalid message type" {
			t.Error(err)
		}
	}
}
