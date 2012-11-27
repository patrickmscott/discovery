package discovery

import (
	"bytes"
	"encoding/json"
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

func readSize(buffer *bytes.Buffer) int {
	bytes := buffer.Next(4)
	return int(bytes[0]<<24 | bytes[1]<<16 | bytes[2]<<8 | bytes[3])
}

func TestWriteSnapshot(t *testing.T) {
	var output bytes.Buffer
	err := writeSnapshot(&output, nil)
	if err != nil {
		t.Error(err)
	}
	if output.Len() != 4 || readSize(&output) != 0 {
		t.Error()
	}
	output.Reset()
	slice := make([]ServiceDefinition, 1)
	slice[0].Host = "host"
	slice[0].Port = 8080
	err = writeSnapshot(&output, slice)
	if err != nil {
		t.Error()
	}
	readSize(&output)
	dec := json.NewDecoder(&output)
	var snapshot []ServiceDefinition
	err = dec.Decode(&snapshot)
	if err != nil {
		t.Error()
	}
	if len(snapshot) != 1 {
		t.Error()
	}
	if snapshot[0].Host != "host" || snapshot[0].Port != 8080 {
		t.Error()
	}
}
