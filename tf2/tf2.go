/*
Provides access to TF2 Game Coordinator functionality.
*/
package tf2

import (
	"github.com/13k/go-steam"
	pbtf2 "github.com/13k/go-steam-resources/protobuf/tf2"
	gc "github.com/13k/go-steam/protocol/gamecoordinator"
	"github.com/13k/go-steam/tf2/protocol"
)

const AppId = 440

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
		t.client.GC.SetGamesPlayed(AppId)
	} else {
		t.client.GC.SetGamesPlayed()
	}
}

func (t *TF2) SetItemPosition(itemID, position uint64) {
	gcMsg := gc.NewGCMsg(AppId, uint32(pbtf2.EGCItemMsg_k_EMsgGCSetSingleItemPosition), &protocol.MsgGCSetItemPosition{
		AssetId:  itemID,
		Position: position,
	})

	t.client.GC.Write(gcMsg)
}

// recipe -2 = wildcard
func (t *TF2) CraftItems(items []uint64, recipe int16) {
	t.client.GC.Write(gc.NewGCMsg(AppId, uint32(pbtf2.EGCItemMsg_k_EMsgGCCraft), &protocol.MsgGCCraft{
		Recipe: recipe,
		Items:  items,
	}))
}

func (t *TF2) DeleteItem(itemID uint64) {
	gcMsg := gc.NewGCMsg(AppId, uint32(pbtf2.EGCItemMsg_k_EMsgGCDelete), &protocol.MsgGCDeleteItem{
		ItemId: itemID,
	})

	t.client.GC.Write(gcMsg)
}

func (t *TF2) NameItem(toolID, target uint64, name string) {
	t.client.GC.Write(gc.NewGCMsg(AppId, uint32(pbtf2.EGCItemMsg_k_EMsgGCNameItem), &protocol.MsgGCNameItem{
		Tool:   toolID,
		Target: target,
		Name:   name,
	}))
}

type GCReadyEvent struct{}

func (t *TF2) HandleGCPacket(packet *gc.GCPacket) {
	if packet.AppId != AppId {
		return
	}
	switch pbtf2.EGCBaseClientMsg(packet.MsgType) {
	case pbtf2.EGCBaseClientMsg_k_EMsgGCClientWelcome:
		t.handleWelcome(packet)
	}
}

func (t *TF2) handleWelcome(_ *gc.GCPacket) {
	// the packet's body is pretty useless
	t.client.Emit(&GCReadyEvent{})
}
