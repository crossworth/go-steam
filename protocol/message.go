package protocol

import (
	"io"

	"github.com/13k/go-steam-resources/steamlang"
)

// Message is the interface for all messages, typically outgoing.
//
// They can also be created by using the Read* methods in a PacketMsg.
type Message interface {
	Serializer

	IsProto() bool
	Type() steamlang.EMsg
	TargetJobID() JobID
	SetTargetJobID(JobID)
	SourceJobID() JobID
	SetSourceJobID(JobID)
}

type MessageBody interface {
	Serializable

	GetEMsg() steamlang.EMsg
}

type StructMessage struct {
	Header  *steamlang.MsgHdr
	Body    MessageBody
	Payload []byte
}

var _ Message = (*StructMessage)(nil)

func NewStructMessage(body MessageBody, payload []byte) *StructMessage {
	hdr := steamlang.NewMsgHdr()
	hdr.Msg = body.GetEMsg()

	return &StructMessage{
		Header:  hdr,
		Body:    body,
		Payload: payload,
	}
}

func (m *StructMessage) Type() steamlang.EMsg {
	return m.Header.Msg
}

func (m *StructMessage) IsProto() bool {
	return false
}

func (m *StructMessage) TargetJobID() JobID {
	return JobID(m.Header.TargetJobID)
}

func (m *StructMessage) SetTargetJobID(job JobID) {
	m.Header.TargetJobID = uint64(job)
}

func (m *StructMessage) SourceJobID() JobID {
	return JobID(m.Header.SourceJobID)
}

func (m *StructMessage) SetSourceJobID(job JobID) {
	m.Header.SourceJobID = uint64(job)
}

func (m *StructMessage) Serialize(w io.Writer) error {
	err := m.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = m.Body.Serialize(w)
	if err != nil {
		return err
	}
	_, err = w.Write(m.Payload)
	return err
}
