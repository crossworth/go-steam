package protocol

import (
	"io"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
	"google.golang.org/protobuf/proto"
)

// ProtoMessage represents a protobuf backed client message with session data.
type ProtoMessage struct {
	Header *ProtoMessageHeader
	Body   proto.Message
}

var _ ClientMessage = (*ProtoMessage)(nil)

func NewProtoMessage(emsg steamlang.EMsg, pb proto.Message) *ProtoMessage {
	header := NewProtoMessageHeader()

	header.SetEMsg(emsg)

	return &ProtoMessage{
		Header: header,
		Body:   pb,
	}
}

func (m *ProtoMessage) Type() steamlang.EMsg {
	return m.Header.EMsg()
}

func (m *ProtoMessage) IsProto() bool {
	return m.Header.IsProto()
}

func (m *ProtoMessage) SessionID() int32 {
	return m.Header.SessionID()
}

func (m *ProtoMessage) SetSessionID(id int32) {
	m.Header.SetSessionID(id)
}

func (m *ProtoMessage) SteamID() steamid.SteamID {
	return m.Header.SteamID()
}

func (m *ProtoMessage) SetSteamID(id steamid.SteamID) {
	m.Header.SetSteamID(id)
}

func (m *ProtoMessage) SourceJobID() JobID {
	return m.Header.SourceJobID()
}

func (m *ProtoMessage) SetSourceJobID(job JobID) {
	m.Header.SetSourceJobID(job)
}

func (m *ProtoMessage) TargetJobID() JobID {
	return m.Header.TargetJobID()
}

func (m *ProtoMessage) SetTargetJobID(job JobID) {
	m.Header.SetTargetJobID(job)
}

func (m *ProtoMessage) Serialize(w io.Writer) error {
	if err := m.Header.Serialize(w); err != nil {
		return err
	}

	body, err := proto.Marshal(m.Body)

	if err != nil {
		return err
	}

	_, err = w.Write(body)

	return err
}
