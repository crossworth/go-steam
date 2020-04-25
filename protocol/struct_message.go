package protocol

import (
	"io"

	"github.com/13k/go-steam-resources/steamlang"
)

// StructMessage represents a struct backed message.
type StructMessage struct {
	Header  *StructMessageHeader
	Body    StructMessageBody
	Payload []byte
}

var _ Message = (*StructMessage)(nil)

func NewStructMessage(body StructMessageBody, payload []byte) *StructMessage {
	header := NewStructMessageHeader()

	header.SetEMsg(body.GetEMsg())

	return &StructMessage{
		Header:  header,
		Body:    body,
		Payload: payload,
	}
}

func (m *StructMessage) Type() steamlang.EMsg {
	return m.Header.EMsg()
}

func (m *StructMessage) IsProto() bool {
	return m.Header.IsProto()
}

func (m *StructMessage) SourceJobID() JobID {
	return m.Header.SourceJobID()
}

func (m *StructMessage) SetSourceJobID(job JobID) {
	m.Header.SetSourceJobID(job)
}

func (m *StructMessage) TargetJobID() JobID {
	return m.Header.TargetJobID()
}

func (m *StructMessage) SetTargetJobID(job JobID) {
	m.Header.SetTargetJobID(job)
}

func (m *StructMessage) Serialize(w io.Writer) error {
	if err := m.Header.Serialize(w); err != nil {
		return err
	}

	if err := m.Body.Serialize(w); err != nil {
		return err
	}

	_, err := w.Write(m.Payload)

	return err
}
