package gc

import (
	"bytes"

	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/protocol"
	"google.golang.org/protobuf/proto"
)

// An incoming, partially unread message from the Game Coordinator.
type GCPacket struct {
	AppID       uint32
	MsgType     uint32
	IsProto     bool
	GCName      string
	Body        []byte
	TargetJobID protocol.JobID
}

func NewGCPacket(wrapper *pb.CMsgGCClient) (*GCPacket, error) {
	packet := &GCPacket{
		AppID:   wrapper.GetAppid(),
		MsgType: wrapper.GetMsgtype(),
		GCName:  wrapper.GetGcname(),
	}

	r := bytes.NewReader(wrapper.GetPayload())

	if steamlang.IsProto(wrapper.GetMsgtype()) {
		packet.MsgType = packet.MsgType & steamlang.EMsgMask
		packet.IsProto = true

		header := steamlang.NewMsgGCHdrProtoBuf()
		err := header.Deserialize(r)

		if err != nil {
			return nil, err
		}

		packet.TargetJobID = protocol.JobID(header.Proto.GetJobidTarget())
	} else {
		header := steamlang.NewMsgGCHdr()

		if err := header.Deserialize(r); err != nil {
			return nil, err
		}

		packet.TargetJobID = protocol.JobID(header.TargetJobID)
	}

	body := make([]byte, r.Len())

	if _, err := r.Read(body); err != nil {
		return nil, err
	}

	packet.Body = body

	return packet, nil
}

func (g *GCPacket) ReadProtoMsg(body proto.Message) {
	proto.Unmarshal(g.Body, body)
}

func (g *GCPacket) ReadMsg(body protocol.MessageBody) {
	body.Deserialize(bytes.NewReader(g.Body))
}
