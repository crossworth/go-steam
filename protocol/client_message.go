package protocol

import (
	"io"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
	"google.golang.org/protobuf/proto"
)

// ClientMessage is the interface for client messages, i.e. messages that are sent after logging in.
//
// ClientProtoMessage and ClientStructMessage implement this.
type ClientMessage interface {
	Message

	SessionID() int32
	SetSessionID(int32)
	SteamID() steamid.SteamID
	SetSteamID(steamid.SteamID)
}

// ClientProtoMessage represents a protobuf backed client message with session data.
type ClientProtoMessage struct {
	Header *steamlang.MsgHdrProtoBuf
	Body   proto.Message
}

var _ ClientMessage = (*ClientProtoMessage)(nil)

func NewClientProtoMessage(msgType steamlang.EMsg, pb proto.Message) *ClientProtoMessage {
	hdr := steamlang.NewMsgHdrProtoBuf()
	hdr.Msg = msgType

	return &ClientProtoMessage{
		Header: hdr,
		Body:   pb,
	}
}

func (c *ClientProtoMessage) IsProto() bool {
	return true
}

func (c *ClientProtoMessage) Type() steamlang.EMsg {
	return steamlang.MakeEMsg(uint32(c.Header.Msg))
}

func (c *ClientProtoMessage) SessionID() int32 {
	return c.Header.Proto.GetClientSessionid()
}

func (c *ClientProtoMessage) SetSessionID(session int32) {
	c.Header.Proto.ClientSessionid = &session
}

func (c *ClientProtoMessage) SteamID() steamid.SteamID {
	return steamid.SteamID(c.Header.Proto.GetSteamid())
}

func (c *ClientProtoMessage) SetSteamID(s steamid.SteamID) {
	c.Header.Proto.Steamid = proto.Uint64(uint64(s))
}

func (c *ClientProtoMessage) TargetJobID() JobID {
	return JobID(c.Header.Proto.GetJobidTarget())
}

func (c *ClientProtoMessage) SetTargetJobID(job JobID) {
	c.Header.Proto.JobidTarget = proto.Uint64(uint64(job))
}

func (c *ClientProtoMessage) SourceJobID() JobID {
	return JobID(c.Header.Proto.GetJobidSource())
}

func (c *ClientProtoMessage) SetSourceJobID(job JobID) {
	c.Header.Proto.JobidSource = proto.Uint64(uint64(job))
}

func (c *ClientProtoMessage) Serialize(w io.Writer) error {
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

// ClientStructMessage represents a struct backed client message.
type ClientStructMessage struct {
	Header  *steamlang.ExtendedClientMsgHdr
	Body    MessageBody
	Payload []byte
}

var _ ClientMessage = (*ClientStructMessage)(nil)

func NewClientStructMessage(body MessageBody, payload []byte) *ClientStructMessage {
	hdr := steamlang.NewExtendedClientMsgHdr()
	hdr.Msg = body.GetEMsg()

	return &ClientStructMessage{
		Header:  hdr,
		Body:    body,
		Payload: payload,
	}
}

func (c *ClientStructMessage) IsProto() bool {
	return true
}

func (c *ClientStructMessage) Type() steamlang.EMsg {
	return c.Header.Msg
}

func (c *ClientStructMessage) SessionID() int32 {
	return c.Header.SessionID
}

func (c *ClientStructMessage) SetSessionID(session int32) {
	c.Header.SessionID = session
}

func (c *ClientStructMessage) SteamID() steamid.SteamID {
	return steamid.SteamID(c.Header.SteamID)
}

func (c *ClientStructMessage) SetSteamID(s steamid.SteamID) {
	c.Header.SteamID = s.Uint64()
}

func (c *ClientStructMessage) TargetJobID() JobID {
	return JobID(c.Header.TargetJobID)
}

func (c *ClientStructMessage) SetTargetJobID(job JobID) {
	c.Header.TargetJobID = uint64(job)
}

func (c *ClientStructMessage) SourceJobID() JobID {
	return JobID(c.Header.SourceJobID)
}

func (c *ClientStructMessage) SetSourceJobID(job JobID) {
	c.Header.SourceJobID = uint64(job)
}

func (c *ClientStructMessage) Serialize(w io.Writer) error {
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
