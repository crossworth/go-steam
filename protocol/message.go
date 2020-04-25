package protocol

import (
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

// Message is the interface for all messages, typically outgoing.
//
// They can also be created by using the Read* methods of Packet.
type Message interface {
	Serializer

	IsProto() bool
	Type() steamlang.EMsg
	SourceJobID() JobID
	SetSourceJobID(JobID)
	TargetJobID() JobID
	SetTargetJobID(JobID)
}

type MessageHeader interface {
	Serializable

	EMsg() steamlang.EMsg
	SetEMsg(steamlang.EMsg)
	IsProto() bool
	SourceJobID() JobID
	SetSourceJobID(JobID)
	TargetJobID() JobID
	SetTargetJobID(JobID)
}

type StructMessageBody interface {
	Serializable

	GetEMsg() steamlang.EMsg
}

// ClientMessage is the interface for client messages, i.e. messages that are sent after logging in.
//
// ClientStructMessage and ProtoMessage implement this.
type ClientMessage interface {
	Message

	SessionID() int32
	SetSessionID(int32)
	SteamID() steamid.SteamID
	SetSteamID(steamid.SteamID)
}
