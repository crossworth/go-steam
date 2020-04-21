package protocol

import (
	"io"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
	"google.golang.org/protobuf/proto"
)

// Interface for all messages, typically outgoing. They can also be created by
// using the Read* methods in a PacketMsg.
type IMsg interface {
	Serializer
	IsProto() bool
	GetMsgType() steamlang.EMsg
	GetTargetJobID() JobID
	SetTargetJobID(JobID)
	GetSourceJobID() JobID
	SetSourceJobID(JobID)
}

// Interface for client messages, i.e. messages that are sent after logging in.
// ClientMsgProtobuf and ClientMsg implement this.
type IClientMsg interface {
	IMsg
	GetSessionID() int32
	SetSessionID(int32)
	GetSteamID() steamid.SteamID
	SetSteamID(steamid.SteamID)
}

// Represents a protobuf backed client message with session data.
type ClientMsgProtobuf struct {
	Header *steamlang.MsgHdrProtoBuf
	Body   proto.Message
}

func NewClientMsgProtobuf(eMsg steamlang.EMsg, body proto.Message) *ClientMsgProtobuf {
	hdr := steamlang.NewMsgHdrProtoBuf()
	hdr.Msg = eMsg
	return &ClientMsgProtobuf{
		Header: hdr,
		Body:   body,
	}
}

func (c *ClientMsgProtobuf) IsProto() bool {
	return true
}

func (c *ClientMsgProtobuf) GetMsgType() steamlang.EMsg {
	return steamlang.MakeEMsg(uint32(c.Header.Msg))
}

func (c *ClientMsgProtobuf) GetSessionID() int32 {
	return c.Header.Proto.GetClientSessionid()
}

func (c *ClientMsgProtobuf) SetSessionID(session int32) {
	c.Header.Proto.ClientSessionid = &session
}

func (c *ClientMsgProtobuf) GetSteamID() steamid.SteamID {
	return steamid.SteamID(c.Header.Proto.GetSteamid())
}

func (c *ClientMsgProtobuf) SetSteamID(s steamid.SteamID) {
	c.Header.Proto.Steamid = proto.Uint64(uint64(s))
}

func (c *ClientMsgProtobuf) GetTargetJobID() JobID {
	return JobID(c.Header.Proto.GetJobidTarget())
}

func (c *ClientMsgProtobuf) SetTargetJobID(job JobID) {
	c.Header.Proto.JobidTarget = proto.Uint64(uint64(job))
}

func (c *ClientMsgProtobuf) GetSourceJobID() JobID {
	return JobID(c.Header.Proto.GetJobidSource())
}

func (c *ClientMsgProtobuf) SetSourceJobID(job JobID) {
	c.Header.Proto.JobidSource = proto.Uint64(uint64(job))
}

func (c *ClientMsgProtobuf) Serialize(w io.Writer) error {
	err := c.Header.Serialize(w)
	if err != nil {
		return err
	}
	body, err := proto.Marshal(c.Body)
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

// Represents a struct backed client message.
type ClientMsg struct {
	Header  *steamlang.ExtendedClientMsgHdr
	Body    MessageBody
	Payload []byte
}

func NewClientMsg(body MessageBody, payload []byte) *ClientMsg {
	hdr := steamlang.NewExtendedClientMsgHdr()
	hdr.Msg = body.GetEMsg()
	return &ClientMsg{
		Header:  hdr,
		Body:    body,
		Payload: payload,
	}
}

func (c *ClientMsg) IsProto() bool {
	return true
}

func (c *ClientMsg) GetMsgType() steamlang.EMsg {
	return c.Header.Msg
}

func (c *ClientMsg) GetSessionID() int32 {
	return c.Header.SessionID
}

func (c *ClientMsg) SetSessionID(session int32) {
	c.Header.SessionID = session
}

func (c *ClientMsg) GetSteamID() steamid.SteamID {
	return steamid.SteamID(c.Header.SteamID)
}

func (c *ClientMsg) SetSteamID(s steamid.SteamID) {
	c.Header.SteamID = s.Uint64()
}

func (c *ClientMsg) GetTargetJobID() JobID {
	return JobID(c.Header.TargetJobID)
}

func (c *ClientMsg) SetTargetJobID(job JobID) {
	c.Header.TargetJobID = uint64(job)
}

func (c *ClientMsg) GetSourceJobID() JobID {
	return JobID(c.Header.SourceJobID)
}

func (c *ClientMsg) SetSourceJobID(job JobID) {
	c.Header.SourceJobID = uint64(job)
}

func (c *ClientMsg) Serialize(w io.Writer) error {
	err := c.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = c.Body.Serialize(w)
	if err != nil {
		return err
	}
	_, err = w.Write(c.Payload)
	return err
}

type Msg struct {
	Header  *steamlang.MsgHdr
	Body    MessageBody
	Payload []byte
}

func NewMsg(body MessageBody, payload []byte) *Msg {
	hdr := steamlang.NewMsgHdr()
	hdr.Msg = body.GetEMsg()

	return &Msg{
		Header:  hdr,
		Body:    body,
		Payload: payload,
	}
}

func (m *Msg) GetMsgType() steamlang.EMsg {
	return m.Header.Msg
}

func (m *Msg) IsProto() bool {
	return false
}

func (m *Msg) GetTargetJobID() JobID {
	return JobID(m.Header.TargetJobID)
}

func (m *Msg) SetTargetJobID(job JobID) {
	m.Header.TargetJobID = uint64(job)
}

func (m *Msg) GetSourceJobID() JobID {
	return JobID(m.Header.SourceJobID)
}

func (m *Msg) SetSourceJobID(job JobID) {
	m.Header.SourceJobID = uint64(job)
}

func (m *Msg) Serialize(w io.Writer) error {
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
