package protocol

import (
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
	"google.golang.org/protobuf/proto"
)

type StructMessageHeader struct {
	*steamlang.MsgHdr
}

var _ MessageHeader = (*StructMessageHeader)(nil)

func NewStructMessageHeader() *StructMessageHeader {
	return &StructMessageHeader{MsgHdr: steamlang.NewMsgHdr()}
}

func (h *StructMessageHeader) EMsg() steamlang.EMsg {
	return h.MsgHdr.Msg
}

func (h *StructMessageHeader) SetEMsg(emsg steamlang.EMsg) {
	h.MsgHdr.Msg = emsg
}

func (h *StructMessageHeader) IsProto() bool {
	return false
}

func (h *StructMessageHeader) SourceJobID() JobID {
	return JobID(h.MsgHdr.SourceJobID)
}

func (h *StructMessageHeader) SetSourceJobID(job JobID) {
	h.MsgHdr.SourceJobID = uint64(job)
}

func (h *StructMessageHeader) TargetJobID() JobID {
	return JobID(h.MsgHdr.TargetJobID)
}

func (h *StructMessageHeader) SetTargetJobID(job JobID) {
	h.MsgHdr.TargetJobID = uint64(job)
}

type ClientStructMessageHeader struct {
	*steamlang.ExtendedClientMsgHdr
}

var _ MessageHeader = (*ClientStructMessageHeader)(nil)

func NewClientStructMessageHeader() *ClientStructMessageHeader {
	return &ClientStructMessageHeader{ExtendedClientMsgHdr: steamlang.NewExtendedClientMsgHdr()}
}

func (h *ClientStructMessageHeader) EMsg() steamlang.EMsg {
	return h.ExtendedClientMsgHdr.Msg
}

func (h *ClientStructMessageHeader) SetEMsg(emsg steamlang.EMsg) {
	h.ExtendedClientMsgHdr.Msg = emsg
}

func (h *ClientStructMessageHeader) IsProto() bool {
	return false
}

func (h *ClientStructMessageHeader) SessionID() int32 {
	return h.ExtendedClientMsgHdr.SessionID
}

func (h *ClientStructMessageHeader) SetSessionID(id int32) {
	h.ExtendedClientMsgHdr.SessionID = id
}

func (h *ClientStructMessageHeader) SteamID() steamid.SteamID {
	return steamid.SteamID(h.ExtendedClientMsgHdr.SteamID)
}

func (h *ClientStructMessageHeader) SetSteamID(id steamid.SteamID) {
	h.ExtendedClientMsgHdr.SteamID = id.Uint64()
}

func (h *ClientStructMessageHeader) SourceJobID() JobID {
	return JobID(h.ExtendedClientMsgHdr.SourceJobID)
}

func (h *ClientStructMessageHeader) SetSourceJobID(job JobID) {
	h.ExtendedClientMsgHdr.SourceJobID = uint64(job)
}

func (h *ClientStructMessageHeader) TargetJobID() JobID {
	return JobID(h.ExtendedClientMsgHdr.TargetJobID)
}

func (h *ClientStructMessageHeader) SetTargetJobID(job JobID) {
	h.ExtendedClientMsgHdr.TargetJobID = uint64(job)
}

type ProtoMessageHeader struct {
	*steamlang.MsgHdrProtoBuf
}

var _ MessageHeader = (*ProtoMessageHeader)(nil)

func NewProtoMessageHeader() *ProtoMessageHeader {
	return &ProtoMessageHeader{MsgHdrProtoBuf: steamlang.NewMsgHdrProtoBuf()}
}

func (h *ProtoMessageHeader) EMsg() steamlang.EMsg {
	return h.MsgHdrProtoBuf.Msg
}

func (h *ProtoMessageHeader) SetEMsg(emsg steamlang.EMsg) {
	h.MsgHdrProtoBuf.Msg = emsg
}

func (h *ProtoMessageHeader) IsProto() bool {
	return true
}

func (h *ProtoMessageHeader) SessionID() int32 {
	return h.MsgHdrProtoBuf.Proto.GetClientSessionid()
}

func (h *ProtoMessageHeader) SetSessionID(id int32) {
	h.MsgHdrProtoBuf.Proto.ClientSessionid = proto.Int32(id)
}

func (h *ProtoMessageHeader) SteamID() steamid.SteamID {
	return steamid.SteamID(h.MsgHdrProtoBuf.Proto.GetSteamid())
}

func (h *ProtoMessageHeader) SetSteamID(id steamid.SteamID) {
	h.MsgHdrProtoBuf.Proto.Steamid = proto.Uint64(id.Uint64())
}

func (h *ProtoMessageHeader) SourceJobID() JobID {
	return JobID(h.MsgHdrProtoBuf.Proto.GetJobidSource())
}

func (h *ProtoMessageHeader) SetSourceJobID(job JobID) {
	h.MsgHdrProtoBuf.Proto.JobidSource = proto.Uint64(uint64(job))
}

func (h *ProtoMessageHeader) TargetJobID() JobID {
	return JobID(h.MsgHdrProtoBuf.Proto.GetJobidTarget())
}

func (h *ProtoMessageHeader) SetTargetJobID(job JobID) {
	h.MsgHdrProtoBuf.Proto.JobidTarget = proto.Uint64(uint64(job))
}
