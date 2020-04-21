package gc

type PacketHandler interface {
	HandleGCPacket(*Packet)
}
