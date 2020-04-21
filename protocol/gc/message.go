package gc

import (
	"io"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/protocol"
	"google.golang.org/protobuf/proto"
)

// Message represents an outgoing message to the Game Coordinator.
type Message interface {
	protocol.Serializer

	IsProto() bool
	GetAppID() uint32
	GetMsgType() uint32
	GetTargetJobID() protocol.JobID
	SetTargetJobID(protocol.JobID)
	GetSourceJobID() protocol.JobID
	SetSourceJobID(protocol.JobID)
}

type ProtoMessage struct {
	AppID  uint32
	Header *steamlang.MsgGCHdrProtoBuf
	Body   proto.Message
}

var _ Message = (*ProtoMessage)(nil)

func NewProtoMessage(appID, msgType uint32, body proto.Message) *ProtoMessage {
	hdr := steamlang.NewMsgGCHdrProtoBuf()
	hdr.Msg = msgType
	return &ProtoMessage{
		AppID:  appID,
		Header: hdr,
		Body:   body,
	}
}

func (g *ProtoMessage) IsProto() bool {
	return true
}

func (g *ProtoMessage) GetAppID() uint32 {
	return g.AppID
}

func (g *ProtoMessage) GetMsgType() uint32 {
	return g.Header.Msg
}

func (g *ProtoMessage) GetTargetJobID() protocol.JobID {
	return protocol.JobID(g.Header.Proto.GetJobidTarget())
}

func (g *ProtoMessage) SetTargetJobID(job protocol.JobID) {
	g.Header.Proto.JobidTarget = proto.Uint64(uint64(job))
}

func (g *ProtoMessage) GetSourceJobID() protocol.JobID {
	return protocol.JobID(g.Header.Proto.GetJobidSource())
}

func (g *ProtoMessage) SetSourceJobID(job protocol.JobID) {
	g.Header.Proto.JobidSource = proto.Uint64(uint64(job))
}

func (g *ProtoMessage) Serialize(w io.Writer) error {
	err := g.Header.Serialize(w)
	if err != nil {
		return err
	}
	body, err := proto.Marshal(g.Body)
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

type StructMessage struct {
	AppID   uint32
	MsgType uint32
	Header  *steamlang.MsgGCHdr
	Body    protocol.Serializer
}

var _ Message = (*StructMessage)(nil)

func NewStructMessage(appID, msgType uint32, body protocol.Serializer) *StructMessage {
	return &StructMessage{
		AppID:   appID,
		MsgType: msgType,
		Header:  steamlang.NewMsgGCHdr(),
		Body:    body,
	}
}

func (g *StructMessage) GetMsgType() uint32 {
	return g.MsgType
}

func (g *StructMessage) GetAppID() uint32 {
	return g.AppID
}

func (g *StructMessage) IsProto() bool {
	return false
}

func (g *StructMessage) GetTargetJobID() protocol.JobID {
	return protocol.JobID(g.Header.TargetJobID)
}

func (g *StructMessage) SetTargetJobID(job protocol.JobID) {
	g.Header.TargetJobID = uint64(job)
}

func (g *StructMessage) GetSourceJobID() protocol.JobID {
	return protocol.JobID(g.Header.SourceJobID)
}

func (g *StructMessage) SetSourceJobID(job protocol.JobID) {
	g.Header.SourceJobID = uint64(job)
}

func (g *StructMessage) Serialize(w io.Writer) error {
	err := g.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = g.Body.Serialize(w)
	return err
}
