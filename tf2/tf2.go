/*
Provides access to TF2 Game Coordinator functionality.
*/
package tf2

import (
	"github.com/13k/go-steam"
	pbtf2 "github.com/13k/go-steam-resources/protobuf/tf2"
	"github.com/13k/go-steam/protocol/gc"
	"github.com/13k/go-steam/tf2/protocol"
)

const AppID = 440

// To use any methods of this, you'll need to SetPlaying(true) and wait for
// the GCReadyEvent.
type TF2 struct {
	client *steam.Client
}

// Creates a new TF2 instance and registers it as a packet handler
func New(client *steam.Client) *TF2 {
	t := &TF2{client}
	client.GC.RegisterPacketHandler(t)
	return t
}

func (t *TF2) SetPlaying(playing bool) {
	if playing {
		t.client.GC.SetGamesPlayed(AppID)
	} else {
		t.client.GC.SetGamesPlayed()
	}
}

func (t *TF2) SetItemPosition(itemID, position uint64) error {
	msgType := uint32(pbtf2.EGCItemMsg_k_EMsgGCSetSingleItemPosition)
	gcMsg := gc.NewStructMessage(AppID, msgType, &protocol.MsgGCSetItemPosition{
		AssetID:  itemID,
		Position: position,
	})

	return t.client.GC.Write(gcMsg)
}

// recipe -2 = wildcard
func (t *TF2) CraftItems(items []uint64, recipe int16) error {
	msgType := uint32(pbtf2.EGCItemMsg_k_EMsgGCCraft)

	return t.client.GC.Write(gc.NewStructMessage(AppID, msgType, &protocol.MsgGCCraft{
		Recipe: recipe,
		Items:  items,
	}))
}

func (t *TF2) DeleteItem(itemID uint64) error {
	msgType := uint32(pbtf2.EGCItemMsg_k_EMsgGCDelete)
	gcMsg := gc.NewStructMessage(AppID, msgType, &protocol.MsgGCDeleteItem{
		ItemID: itemID,
	})

	return t.client.GC.Write(gcMsg)
}

func (t *TF2) NameItem(toolID, target uint64, name string) error {
	msgType := uint32(pbtf2.EGCItemMsg_k_EMsgGCNameItem)

	return t.client.GC.Write(gc.NewStructMessage(AppID, msgType, &protocol.MsgGCNameItem{
		Tool:   toolID,
		Target: target,
		Name:   name,
	}))
}

type GCReadyEvent struct{}

func (t *TF2) HandleGCPacket(packet *gc.Packet) {
	if packet.AppID != AppID {
		return
	}

	switch pbtf2.EGCBaseClientMsg(packet.MsgType) {
	case pbtf2.EGCBaseClientMsg_k_EMsgGCClientWelcome:
		t.handleWelcome(packet)
	}
}

func (t *TF2) handleWelcome(_ *gc.Packet) {
	// the packet's body is pretty useless
	t.client.Emit(&GCReadyEvent{})
}
