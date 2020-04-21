package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/13k/go-steam-resources/steamlang"
	"google.golang.org/protobuf/proto"
)

// TODO: Headers are always deserialized twice.

// Represents an incoming, partially unread message.
type Packet struct {
	EMsg        steamlang.EMsg
	IsProto     bool
	TargetJobID JobID
	SourceJobID JobID
	Data        []byte
}

func NewPacket(data []byte) (*Packet, error) {
	var rawEMsg uint32

	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &rawEMsg); err != nil {
		return nil, err
	}

	eMsg := steamlang.MakeEMsg(rawEMsg)
	buf := bytes.NewReader(data)

	if eMsg == steamlang.EMsg_ChannelEncryptRequest || eMsg == steamlang.EMsg_ChannelEncryptResult {
		header := steamlang.NewMsgHdr()
		header.Msg = eMsg

		if err := header.Deserialize(buf); err != nil {
			return nil, err
		}

		return &Packet{
			EMsg:        eMsg,
			IsProto:     false,
			TargetJobID: JobID(header.TargetJobID),
			SourceJobID: JobID(header.SourceJobID),
			Data:        data,
		}, nil
	} else if steamlang.IsProto(rawEMsg) {
		header := steamlang.NewMsgHdrProtoBuf()
		header.Msg = eMsg

		if err := header.Deserialize(buf); err != nil {
			return nil, err
		}

		return &Packet{
			EMsg:        eMsg,
			IsProto:     true,
			TargetJobID: JobID(header.Proto.GetJobidTarget()),
			SourceJobID: JobID(header.Proto.GetJobidSource()),
			Data:        data,
		}, nil
	} else {
		header := steamlang.NewExtendedClientMsgHdr()
		header.Msg = eMsg

		if err := header.Deserialize(buf); err != nil {
			return nil, err
		}

		return &Packet{
			EMsg:        eMsg,
			IsProto:     false,
			TargetJobID: JobID(header.TargetJobID),
			SourceJobID: JobID(header.SourceJobID),
			Data:        data,
		}, nil
	}
}

func (p *Packet) String() string {
	return fmt.Sprintf(
		"Packet{EMsg = %v, Proto = %v, Len = %v, TargetJobID = %v, SourceJobID = %v}",
		p.EMsg,
		p.IsProto,
		len(p.Data),
		p.TargetJobID,
		p.SourceJobID,
	)
}

func (p *Packet) ReadProtoMsg(body proto.Message) (*ClientProtoMessage, error) {
	header := steamlang.NewMsgHdrProtoBuf()
	buf := bytes.NewBuffer(p.Data)

	if err := header.Deserialize(buf); err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(buf.Bytes(), body); err != nil {
		return nil, err
	}

	msg := &ClientProtoMessage{
		Header: header,
		Body:   body,
	}

	return msg, nil
}

func (p *Packet) ReadClientMsg(body MessageBody) (*ClientStructMessage, error) {
	header := steamlang.NewExtendedClientMsgHdr()
	buf := bytes.NewReader(p.Data)

	if err := header.Deserialize(buf); err != nil {
		return nil, err
	}

	if err := body.Deserialize(buf); err != nil {
		return nil, err
	}

	payload := make([]byte, buf.Len())

	if _, err := buf.Read(payload); err != nil {
		return nil, err
	}

	msg := &ClientStructMessage{
		Header:  header,
		Body:    body,
		Payload: payload,
	}

	return msg, nil
}

func (p *Packet) ReadMsg(body MessageBody) (*StructMessage, error) {
	header := steamlang.NewMsgHdr()
	buf := bytes.NewReader(p.Data)

	if err := header.Deserialize(buf); err != nil {
		return nil, err
	}

	if err := body.Deserialize(buf); err != nil {
		return nil, err
	}

	payload := make([]byte, buf.Len())

	if _, err := buf.Read(payload); err != nil {
		return nil, err
	}

	msg := &StructMessage{
		Header:  header,
		Body:    body,
		Payload: payload,
	}

	return msg, nil
}
