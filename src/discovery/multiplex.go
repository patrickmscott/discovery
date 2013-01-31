package discovery

import (
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"errors"
	"io"
	"net/rpc"
	"reflect"
)

type multiplexCodec struct {
	rpc.ServerCodec
	rpc.ClientCodec

	requestChan  chan *Message
	responseChan chan *Message
	errorChan    chan error
	rwc          io.ReadWriteCloser
}

func newMultiplexCodec(rwc io.ReadWriteCloser) *multiplexCodec {
	mux := &multiplexCodec{
		// TODO(pscott): Should these be buffered?
		requestChan:  make(chan *Message),
		responseChan: make(chan *Message),
		errorChan:    make(chan error),
		rwc:          rwc}
	go mux.input()
	return mux
}

const maxMessageSize = 1024 * 1024 // 1 MB limit

// TODO(pscott): Cap the size of the buffer when reused?
type byteBuffer []byte

func (b byteBuffer) ensureSize(size int) {
	if b == nil || size > cap(b) {
		b = make([]byte, size)
	}
	b = b[0:0]
}

// input() handles reading both Requests and Responses from the connection.
// Depending on the type of Message, the correct channel is used to send the
// object.
func (mux *multiplexCodec) input() {
	var buf byteBuffer
	for {
		var msg *Message = &Message{}
		// Read the fixed size of the message.
		var size int32
		err := binary.Read(mux.rwc, binary.BigEndian, &size)
		if err != nil {
			mux.errorChan <- err
			continue
		}

		if size > maxMessageSize {
			mux.errorChan <- errors.New("Max message size exceeded")
			continue
		}

		buf.ensureSize(int(size))
		_, err = io.ReadFull(mux.rwc, buf)
		// err will be non-nil if ReadFull fails to read len(buf) bytes
		if err != nil {
			mux.errorChan <- err
			continue
		}

		// Parse the message and send it to the appropriate channel.
		err = proto.Unmarshal(buf, msg)
		switch {
		case err != nil:
			mux.errorChan <- err
		case *msg.Type < MessageType___LAST_REQUEST:
			mux.requestChan <- msg
		default:
			mux.responseChan <- msg
		}
	}
}

func (mux *multiplexCodec) writeMessage(msg proto.Message) error {
	bytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	err = binary.Write(mux.rwc, binary.BigEndian, int32(len(bytes)))
	if err != nil {
		return err
	}
	_, err = mux.rwc.Write(bytes)
	return err
}

var typeMap map[string]MessageType

func typeOf(i interface{}) string {
	return reflect.TypeOf(i).Elem().String()
}

func init() {
	typeMap = make(map[string]MessageType)
	typeMap[typeOf((*JoinRequest)(nil))] = MessageType_JOIN_REQUEST
	typeMap[typeOf((*LeaveRequest)(nil))] = MessageType_LEAVE_REQUEST
	typeMap[typeOf((*SnapshotRequest)(nil))] = MessageType_SNAPSHOT_REQUEST
	typeMap[typeOf((*WatchRequest)(nil))] = MessageType_WATCH_REQUEST
	typeMap[typeOf((*IgnoreRequest)(nil))] = MessageType_IGNORE_REQUEST

	typeMap[typeOf((*ErrorResponse)(nil))] = MessageType_ERROR_RESPONSE
	typeMap[typeOf((*SnapshotResponse)(nil))] = MessageType_SNAPSHOT_RESPONSE
}

func (mux *multiplexCodec) WriteRequest(req *rpc.Request, i interface{}) error {
	var msg Message
	var err error
	msg.Sequence = proto.Uint64(req.Seq)
	msg.Type = typeMap[typeOf(i)].Enum()
	msg.Payload, err = proto.Marshal(i.(proto.Message))
	if err != nil {
		return err
	}
	return mux.writeMessage(&msg)
}

const requestMethod = "Discovery.Request"
const responseMethod = "Discovery.Response"

func (mux *multiplexCodec) ReadResponseHeader(res *rpc.Response) error {
	select {
	case err := <-mux.errorChan:
		return err
	case msg := <-mux.responseChan:
		res.ServiceMethod = responseMethod
		res.Seq = *msg.Sequence
	}
	return nil
}

func (m *multiplexCodec) ReadResponseBody(i interface{}) error {
	proto.Unmarshal(nil, &Message{})
	return nil
}

func (m *multiplexCodec) Close() error {
	return m.rwc.Close()
}

func (m *multiplexCodec) ReadRequestHeader(req *rpc.Request) error {
	return nil
}

func (m *multiplexCodec) ReadRequestBody(i interface{}) error {
	return nil
}

func (mux *multiplexCodec) WriteResponse(
	res *rpc.Response, i interface{}) error {
	var msg Message
	var err error
	msg.Sequence = proto.Uint64(res.Seq)
	msg.Type = typeMap[typeOf(i)].Enum()
	msg.Payload, err = proto.Marshal(i.(proto.Message))
	if err != nil {
		return err
	}
	return mux.writeMessage(&msg)
}
