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
	handlers []gc.PacketHandler
}

var _ protocol.PacketHandler = (*GameCoordinator)(nil)

func NewGC(client *Client) *GameCoordinator {
	return &GameCoordinator{
		client:   client,
		handlers: make([]gc.PacketHandler, 0),
	}
}

func (g *GameCoordinator) RegisterPacketHandler(handler gc.PacketHandler) {
	g.handlers = append(g.handlers, handler)
}

func (g *GameCoordinator) HandlePacket(packet *protocol.Packet) {
	if packet.EMsg() != steamlang.EMsg_ClientFromGC {
		return
	}

	msg := &pb.CMsgGCClient{}

	if _, err := packet.ReadProtoMsg(msg); err != nil {
		g.client.Errorf("gc: error reading message: %v", err)
		return
	}

	p, err := gc.NewPacket(msg)

	if err != nil {
		g.client.Errorf("gc: error reading message: %v", err)
		return
	}

	for _, handler := range g.handlers {
		handler.HandleGCPacket(p)
	}
}

func (g *GameCoordinator) Write(msg gc.Message) error {
	buf := &bytes.Buffer{}

	if err := msg.Serialize(buf); err != nil {
		return err
	}

	msgType := msg.GetMsgType()

	if msg.IsProto() {
		msgType = steamlang.MaskProto(msgType)
	}

	g.client.Write(protocol.NewProtoMessage(steamlang.EMsg_ClientToGC, &pb.CMsgGCClient{
		Msgtype: proto.Uint32(msgType),
		Appid:   proto.Uint32(msg.GetAppID()),
		Payload: buf.Bytes(),
	}))

	return nil
}

// Sets you in the given games. Specify none to quit all games.
func (g *GameCoordinator) SetGamesPlayed(appIDs ...uint64) {
	games := make([]*pb.CMsgClientGamesPlayed_GamePlayed, len(appIDs))

	for i, appID := range appIDs {
		games[i] = &pb.CMsgClientGamesPlayed_GamePlayed{
			GameId: proto.Uint64(appID),
		}
	}

	g.client.Write(protocol.NewProtoMessage(steamlang.EMsg_ClientGamesPlayed, &pb.CMsgClientGamesPlayed{
		GamesPlayed: games,
	}))
}
