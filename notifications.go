package steam

import (
	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/protocol"
)

type Notifications struct {
	// Maps notification types to their count. If a type is not present in the map,
	// its count is zero.
	notifications map[NotificationType]uint
	client        *Client
}

var _ protocol.PacketHandler = (*Notifications)(nil)

func NewNotifications(client *Client) *Notifications {
	return &Notifications{
		make(map[NotificationType]uint),
		client,
	}
}

func (n *Notifications) HandlePacket(packet *protocol.Packet) {
	switch packet.EMsg() {
	case steamlang.EMsg_ClientUserNotifications:
		n.handleClientUserNotifications(packet)
	}
}

type NotificationType uint

const (
	TradeOffer NotificationType = 1
)

func (n *Notifications) handleClientUserNotifications(packet *protocol.Packet) {
	msg := &pb.CMsgClientUserNotifications{}

	if _, err := packet.ReadProtoMsg(msg); err != nil {
		n.client.Errorf("notifications/ClientUserNotifications: error reading message: %v", err)
		return
	}

	for _, notification := range msg.GetNotifications() {
		typ := NotificationType(*notification.UserNotificationType)
		count := uint(*notification.Count)
		n.notifications[typ] = count
		n.client.Emit(&NotificationEvent{typ, count})
	}

	// check if there is a notification in our map that isn't in the current packet
	for typ := range n.notifications {
		exists := false

		for _, t := range msg.GetNotifications() {
			if NotificationType(t.GetUserNotificationType()) == typ {
				exists = true
				break
			}
		}

		if !exists {
			delete(n.notifications, typ)
			n.client.Emit(&NotificationEvent{typ, 0})
		}
	}
}
