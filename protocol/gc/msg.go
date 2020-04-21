package gc

import (
	"io"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/protocol"
	"google.golang.org/protobuf/proto"
)

// An outgoing message to the Game Coordinator.
type IGCMsg interface {
	protocol.Serializer
	IsProto() bool
	GetAppID() uint32
	GetMsgType() uint32

	GetTargetJobID() protocol.JobID
	SetTargetJobID(protocol.JobID)
	GetSourceJobID() protocol.JobID
	SetSourceJobID(protocol.JobID)
}

type GCMsgProtobuf struct {
	AppID  uint32
	Header *steamlang.MsgGCHdrProtoBuf
	Body   proto.Message
}

var _ IGCMsg = (*GCMsgProtobuf)(nil)

func NewGCMsgProtobuf(appID, msgType uint32, body proto.Message) *GCMsgProtobuf {
	hdr := steamlang.NewMsgGCHdrProtoBuf()
	hdr.Msg = msgType
	return &GCMsgProtobuf{
		AppID:  appID,
		Header: hdr,
		Body:   body,
	}
}

func (g *GCMsgProtobuf) IsProto() bool {
	return true
}

func (g *GCMsgProtobuf) GetAppID() uint32 {
	return g.AppID
}

func (g *GCMsgProtobuf) GetMsgType() uint32 {
	return g.Header.Msg
}

func (g *GCMsgProtobuf) GetTargetJobID() protocol.JobID {
	return protocol.JobID(g.Header.Proto.GetJobidTarget())
}

func (g *GCMsgProtobuf) SetTargetJobID(job protocol.JobID) {
	g.Header.Proto.JobidTarget = proto.Uint64(uint64(job))
}

func (g *GCMsgProtobuf) GetSourceJobID() protocol.JobID {
	return protocol.JobID(g.Header.Proto.GetJobidSource())
}

func (g *GCMsgProtobuf) SetSourceJobID(job protocol.JobID) {
	g.Header.Proto.JobidSource = proto.Uint64(uint64(job))
}

func (g *GCMsgProtobuf) Serialize(w io.Writer) error {
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

type GCMsg struct {
	AppID   uint32
	MsgType uint32
	Header  *steamlang.MsgGCHdr
	Body    protocol.Serializer
}

var _ IGCMsg = (*GCMsg)(nil)

func NewGCMsg(appID, msgType uint32, body protocol.Serializer) *GCMsg {
	return &GCMsg{
		AppID:   appID,
		MsgType: msgType,
		Header:  steamlang.NewMsgGCHdr(),
		Body:    body,
	}
}

func (g *GCMsg) GetMsgType() uint32 {
	return g.MsgType
}

func (g *GCMsg) GetAppID() uint32 {
	return g.AppID
}

func (g *GCMsg) IsProto() bool {
	return false
}

func (g *GCMsg) GetTargetJobID() protocol.JobID {
	return protocol.JobID(g.Header.TargetJobID)
}

func (g *GCMsg) SetTargetJobID(job protocol.JobID) {
	g.Header.TargetJobID = uint64(job)
}

func (g *GCMsg) GetSourceJobID() protocol.JobID {
	return protocol.JobID(g.Header.SourceJobID)
}

func (g *GCMsg) SetSourceJobID(job protocol.JobID) {
	g.Header.SourceJobID = uint64(job)
}

func (g *GCMsg) Serialize(w io.Writer) error {
	err := g.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = g.Body.Serialize(w)
	return err
}
