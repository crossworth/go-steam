package steam

import (
	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/protocol"
	"github.com/13k/go-steam/steamid"
	"google.golang.org/protobuf/proto"
)

type TradeRequestID uint32

// Trading provides access to the Steam client's part of Steam Trading, that is bootstrapping the
// trade.
//
// The trade itself is not handled by the Steam client itself, but it's a part of the Steam website.
//
// You'll receive a TradeProposedEvent when a friend proposes a trade. You can accept it with the
// RespondRequest method. You can request a trade yourself with RequestTrade.
type Trading struct {
	client *Client
}

func (t *Trading) HandlePacket(packet *protocol.Packet) {
	switch packet.EMsg {
	case steamlang.EMsg_EconTrading_InitiateTradeProposed:
		msg := &pb.CMsgTrading_InitiateTradeRequest{}

		if _, err := packet.ReadProtoMsg(msg); err != nil {
			t.client.Errorf("error reading message: %v", err)
			return
		}

		t.client.Emit(&TradeProposedEvent{
			RequestID: TradeRequestID(msg.GetTradeRequestId()),
			Other:     steamid.SteamID(msg.GetOtherSteamid()),
		})
	case steamlang.EMsg_EconTrading_InitiateTradeResult:
		msg := &pb.CMsgTrading_InitiateTradeResponse{}

		if _, err := packet.ReadProtoMsg(msg); err != nil {
			t.client.Errorf("error reading message: %v", err)
			return
		}

		t.client.Emit(&TradeResultEvent{
			RequestID: TradeRequestID(msg.GetTradeRequestId()),
			Response:  steamlang.EEconTradeResponse(msg.GetResponse()),
			Other:     steamid.SteamID(msg.GetOtherSteamid()),

			NumDaysSteamGuardRequired:            msg.GetSteamguardRequiredDays(),
			NumDaysNewDeviceCooldown:             msg.GetNewDeviceCooldownDays(),
			DefaultNumDaysPasswordResetProbation: msg.GetDefaultPasswordResetProbationDays(),
			NumDaysPasswordResetProbation:        msg.GetPasswordResetProbationDays(),
		})
	case steamlang.EMsg_EconTrading_StartSession:
		msg := &pb.CMsgTrading_StartSession{}

		if _, err := packet.ReadProtoMsg(msg); err != nil {
			t.client.Errorf("error reading message: %v", err)
			return
		}

		t.client.Emit(&TradeSessionStartEvent{
			Other: steamid.SteamID(msg.GetOtherSteamid()),
		})
	}
}

// Requests a trade. You'll receive a TradeResultEvent if the request fails or
// if the friend accepted the trade.
func (t *Trading) RequestTrade(other steamid.SteamID) {
	pbmsg := &pb.CMsgTrading_InitiateTradeRequest{
		OtherSteamid: proto.Uint64(uint64(other)),
	}

	msg := protocol.NewClientProtoMessage(steamlang.EMsg_EconTrading_InitiateTradeRequest, pbmsg)

	t.client.Write(msg)
}

// Responds to a TradeProposedEvent.
func (t *Trading) RespondRequest(requestID TradeRequestID, accept bool) {
	var resp uint32

	if !accept {
		resp = 1
	}

	pbmsg := &pb.CMsgTrading_InitiateTradeResponse{
		TradeRequestId: proto.Uint32(uint32(requestID)),
		Response:       proto.Uint32(resp),
	}

	msg := protocol.NewClientProtoMessage(steamlang.EMsg_EconTrading_InitiateTradeResponse, pbmsg)

	t.client.Write(msg)
}

// This cancels a request made with RequestTrade.
func (t *Trading) CancelRequest(other steamid.SteamID) {
	pbmsg := &pb.CMsgTrading_CancelTradeRequest{
		OtherSteamid: proto.Uint64(uint64(other)),
	}

	msg := protocol.NewClientProtoMessage(steamlang.EMsg_EconTrading_CancelTradeRequest, pbmsg)

	t.client.Write(msg)
}
