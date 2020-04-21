package steam

import (
	"bytes"

	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/protocol"
	"github.com/13k/go-steam/protocol/gc"
	"google.golang.org/protobuf/proto"
)

type GameCoordinator struct {
	client   *Client
	handlers []GCPacketHandler
}

func newGC(client *Client) *GameCoordinator {
	return &GameCoordinator{
		client:   client,
		handlers: make([]GCPacketHandler, 0),
	}
}

type GCPacketHandler interface {
	HandleGCPacket(*gc.GCPacket)
}

func (g *GameCoordinator) RegisterPacketHandler(handler GCPacketHandler) {
	g.handlers = append(g.handlers, handler)
}

func (g *GameCoordinator) HandlePacket(packet *protocol.Packet) {
	if packet.EMsg != steamlang.EMsg_ClientFromGC {
		return
	}

	msg := &pb.CMsgGCClient{}

	packet.ReadProtoMsg(msg)

	p, err := gc.NewGCPacket(msg)

	if err != nil {
		g.client.Errorf("Error reading GC message: %v", err)
		return
	}

	for _, handler := range g.handlers {
		handler.HandleGCPacket(p)
	}
}

func (g *GameCoordinator) Write(msg gc.IGCMsg) {
	buf := &bytes.Buffer{}

	msg.Serialize(buf)

	msgType := msg.GetMsgType()

	if msg.IsProto() {
		msgType = msgType | 0x80000000 // mask with protoMask
	}

	g.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientToGC, &pb.CMsgGCClient{
		Msgtype: proto.Uint32(msgType),
		Appid:   proto.Uint32(msg.GetAppID()),
		Payload: buf.Bytes(),
	}))
}

// Sets you in the given games. Specify none to quit all games.
func (g *GameCoordinator) SetGamesPlayed(appIDs ...uint64) {
	games := make([]*pb.CMsgClientGamesPlayed_GamePlayed, len(appIDs))

	for i, appID := range appIDs {
		games[i] = &pb.CMsgClientGamesPlayed_GamePlayed{
			GameId: proto.Uint64(appID),
		}
	}

	g.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientGamesPlayed, &pb.CMsgClientGamesPlayed{
		GamesPlayed: games,
	}))
}
