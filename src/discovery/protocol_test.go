package discovery

import (
	"bytes"
	"encoding/json"
	"hash/crc32"
	"io"
	"math"
	"testing"
)

func TestProtocolReadRequestMagicNumber(t *testing.T) {
	var proto Protocol
	_, err := proto.readRequest(&bytes.Buffer{})
	if err == nil {
		t.Error()
	}

	var buffer bytes.Buffer
	proto.writeInt(&buffer, 1234)
	_, err = proto.readRequest(&buffer)
	if err == nil || err != ErrMagicNumber {
		t.Error(err)
	}

	buffer.Reset()
	proto.writeInt(&buffer, magicNumber)
	_, err = proto.readRequest(&buffer)
	if err != nil && err == ErrMagicNumber {
		t.Error(err)
	}
}

func TestProtocolReadRequestChecksum(t *testing.T) {
	var proto Protocol
	var buffer bytes.Buffer
	proto.writeInt(&buffer, magicNumber)
	_, err := proto.readRequest(&buffer)
	if err == nil || err != ErrChecksum {
		t.Error(err)
	}

	buffer.Reset()
	proto.writeInt(&buffer, magicNumber)
	proto.writeInt(&buffer, 0)
	proto.writeInt(&buffer, 1) // size
	buffer.WriteByte(0)
	_, err = proto.readRequest(&buffer)
	if err == nil || err != ErrChecksum {
		t.Error(err)
	}

	var bytes []byte = []byte{0, '{', '}'}
	buffer.Reset()
	proto.writeInt(&buffer, magicNumber)
	proto.writeInt(&buffer, int(crc32.ChecksumIEEE(bytes)))
	proto.writeInt(&buffer, len(bytes))
	buffer.Write(bytes)
	_, err = proto.readRequest(&buffer)
	if err != nil {
		t.Error(err)
	}
}

func TestProtocolReadRequestSize(t *testing.T) {
	var proto Protocol
	var buffer bytes.Buffer
	proto.writeInt(&buffer, magicNumber)
	proto.writeInt(&buffer, 0) // checksum
	buffer.WriteByte(0)
	_, err := proto.readRequest(&buffer)
	if err == nil || err != ErrRequestSize {
		t.Error(err)
	}
	buffer.Reset()
	proto.writeInt(&buffer, magicNumber)
	proto.writeInt(&buffer, 0) // checksum
	proto.writeInt(&buffer, 2*1024*1024)
	_, err = proto.readRequest(&buffer)
	if err == nil || err != ErrRequestSize {
		t.Error(err)
	}
}

func TestProtocolReadRequestEOF(t *testing.T) {
	var proto Protocol
	var buffer bytes.Buffer
	proto.writeInt(&buffer, magicNumber)
	proto.writeInt(&buffer, 0)  // checksum
	proto.writeInt(&buffer, 20) // size
	_, err := proto.readRequest(&buffer)
	if err == nil || err != io.EOF {
		t.Error(err)
	}

	buffer.Reset()
	proto.writeInt(&buffer, magicNumber)
	proto.writeInt(&buffer, 0)  // checksum
	proto.writeInt(&buffer, 20) // size
	buffer.WriteString("data")
	_, err = proto.readRequest(&buffer)
	if err == nil || err != io.EOF {
		t.Error(err)
	}
}

func TestProtocolReadRequestType(t *testing.T) {
	var proto Protocol
	var buffer bytes.Buffer

	var i RequestType
	for i = 0; i < lastRequestType; i++ {
		var bytes [1]byte
		bytes[0] = byte(i)
		buffer.Reset()
		proto.writeInt(&buffer, magicNumber)
		proto.writeInt(&buffer, int(crc32.ChecksumIEEE(bytes[0:])))
		proto.writeInt(&buffer, 1) // size
		buffer.Write(bytes[0:])
		req, err := proto.readRequest(&buffer)
		if req == nil {
			t.Error(err)
		}
		if req.Type() != i {
			t.Error("Type mismatch", i, req.Type())
		}
	}

	for i = lastRequestType; i < math.MaxUint8; i++ {
		var bytes [1]byte
		bytes[0] = byte(i)
		buffer.Reset()
		proto.writeInt(&buffer, magicNumber)
		proto.writeInt(&buffer, int(crc32.ChecksumIEEE(bytes[0:])))
		proto.writeInt(&buffer, 1) // size
		buffer.Write(bytes[0:])
		_, err := proto.readRequest(&buffer)
		if err == nil || err != ErrRequestType {
			t.Error(err)
		}
	}
}

type jsonTest struct {
	Name  string
	Value int
}

func TestProtocolWriteJson(t *testing.T) {
	var proto Protocol
	var output bytes.Buffer
	err := proto.writeJson(&output, nil)
	if err != nil {
		t.Error(err)
	}
	magic, err := proto.readInt(&output)
	if magic != magicNumber || err != nil {
		t.Error(magic, err)
	}
	checksum, err := proto.readInt(&output)
	if checksum != int(crc32.ChecksumIEEE(nil)) || err != nil {
		t.Error(checksum, err)
	}
	size, err := proto.readInt(&output)
	if size != 0 || err != nil {
		t.Error(size, err)
	}
	output.Reset()
	slice := make([]*jsonTest, 1)
	slice[0] = &jsonTest{"name", 42}
	err = proto.writeJson(&output, slice)
	if err != nil {
		t.Error()
	}
	proto.readInt(&output) // magic
	proto.readInt(&output) // checksum
	proto.readInt(&output) // size
	dec := json.NewDecoder(&output)
	var result []*jsonTest
	err = dec.Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Error()
	}
	if result[0].Name != "name" || result[0].Value != 42 {
		t.Error()
	}
}
