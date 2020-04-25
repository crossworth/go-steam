package protocol

import (
	"io"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

// ClientStructMessage represents a struct backed client message.
type ClientStructMessage struct {
	Header  *ClientStructMessageHeader
	Body    StructMessageBody
	Payload []byte
}

var _ ClientMessage = (*ClientStructMessage)(nil)

func NewClientStructMessage(body StructMessageBody, payload []byte) *ClientStructMessage {
	header := NewClientStructMessageHeader()

	header.SetEMsg(body.GetEMsg())

	return &ClientStructMessage{
		Header:  header,
		Body:    body,
		Payload: payload,
	}
}

func (m *ClientStructMessage) Type() steamlang.EMsg {
	return m.Header.EMsg()
}

func (m *ClientStructMessage) IsProto() bool {
	return m.Header.IsProto()
}

func (m *ClientStructMessage) SessionID() int32 {
	return m.Header.SessionID()
}

func (m *ClientStructMessage) SetSessionID(session int32) {
	m.Header.SetSessionID(session)
}

func (m *ClientStructMessage) SteamID() steamid.SteamID {
	return m.Header.SteamID()
}

func (m *ClientStructMessage) SetSteamID(s steamid.SteamID) {
	m.Header.SetSteamID(s)
}

func (m *ClientStructMessage) SourceJobID() JobID {
	return m.Header.SourceJobID()
}

func (m *ClientStructMessage) SetSourceJobID(job JobID) {
	m.Header.SetSourceJobID(job)
}

func (m *ClientStructMessage) TargetJobID() JobID {
	return m.Header.TargetJobID()
}

func (m *ClientStructMessage) SetTargetJobID(job JobID) {
	m.Header.SetTargetJobID(job)
}

func (m *ClientStructMessage) Serialize(w io.Writer) error {
	if err := m.Header.Serialize(w); err != nil {
		return err
	}

	if err := m.Body.Serialize(w); err != nil {
		return err
	}

	_, err := w.Write(m.Payload)

	return err
}
