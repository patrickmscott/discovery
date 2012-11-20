package discovery

import (
	"bytes"
	"math"
	"testing"
)

func TestBadSizeLength(t *testing.T) {
	_, err := readRequest(&bytes.Buffer{})
	if err == nil {
		t.Error()
	}

	var buffer bytes.Buffer
	buffer.WriteByte(0)
	_, err = readRequest(&buffer)
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
	var buffer bytes.Buffer
	writeLength(&buffer, 2*1024*1024)
	_, err := readRequest(&buffer)
	if err == nil {
		t.Error()
	}
	if err.Error() != "Invalid message size, too large" {
		t.Error("Wrong error")
	}
}

func TestMessageTooShort(t *testing.T) {
	var buffer bytes.Buffer
	writeLength(&buffer, 20)
	_, err := readRequest(&buffer)
	if err == nil {
		t.Error()
	}

	buffer.Reset()
	writeLength(&buffer, 20)
	buffer.WriteString("data")
	_, err = readRequest(&buffer)
	if err == nil {
		t.Error()
	}
}

func TestMessageTypes(t *testing.T) {
	var buffer bytes.Buffer

	var i MessageType
	for i = 0; i < lastMessageType; i++ {
		buffer.Reset()
		writeLength(&buffer, 1)
		buffer.WriteByte(byte(i))
		req, err := readRequest(&buffer)
		if err != nil && err.Error() == "Invalid message type" {
			t.Error()
		}
		if req.Type() != i {
			t.Error("Type mismatch", i, req.Type())
		}
	}

	for i = lastMessageType; i < math.MaxUint8; i++ {
		buffer.Reset()
		writeLength(&buffer, 1)
		buffer.WriteByte(byte(i))
		_, err := readRequest(&buffer)
		if err == nil || err.Error() != "Invalid message type" {
			t.Error(err)
		}
	}
}
