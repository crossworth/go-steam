package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/13k/go-steam-resources/steamlang"
	"google.golang.org/protobuf/proto"
)

type PacketHandler interface {
	HandlePacket(*Packet)
}

// Packet represents an incoming, partially decoded message.
type Packet struct {
	Header  MessageHeader // decoded header
	Payload io.Reader     // unread data
	Data    []byte        // whole packet data, including header
}

func NewPacket(data []byte) (*Packet, error) {
	var rawEMsg uint32

	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &rawEMsg); err != nil {
		return nil, fmt.Errorf("protocol/Packet: error reading packet EMsg: %w", err)
	}

	isProto := steamlang.IsProto(rawEMsg)
	emsg := steamlang.MakeEMsg(rawEMsg)

	var header MessageHeader

	if isProto {
		header = NewProtoMessageHeader()
	} else {
		switch emsg {
		case steamlang.EMsg_ChannelEncryptRequest, steamlang.EMsg_ChannelEncryptResult:
			header = NewStructMessageHeader()
		default:
			header = NewClientStructMessageHeader()
		}
	}

	buf := bytes.NewReader(data)
	body := &bytes.Buffer{}

	if err := header.Deserialize(buf); err != nil {
		return nil, fmt.Errorf("protocol/Packet: error deserializing header: %w", err)
	}

	if _, err := buf.WriteTo(body); err != nil {
		return nil, err
	}

	p := &Packet{
		Header:  header,
		Data:    data,
		Payload: body,
	}

	return p, nil
}

func (p *Packet) EMsg() steamlang.EMsg { return p.Header.EMsg() }
func (p *Packet) IsProto() bool        { return p.Header.IsProto() }
func (p *Packet) SourceJobID() JobID   { return p.Header.SourceJobID() }
func (p *Packet) TargetJobID() JobID   { return p.Header.TargetJobID() }

func (p *Packet) String() string {
	return fmt.Sprintf(
		"Packet{EMsg=%s, Proto=%v, Len=%d, TargetJobID=%d, SourceJobID=%d}",
		p.EMsg(),
		p.IsProto(),
		len(p.Data),
		p.TargetJobID(),
		p.SourceJobID(),
	)
}

func (p *Packet) ReadMsg(body StructMessageBody) (*StructMessage, error) {
	header, ok := p.Header.(*StructMessageHeader)

	if !ok {
		return nil, fmt.Errorf("protocol/ReadMsg: invalid packet header %T", p.Header)
	}

	if header == nil {
		return nil, fmt.Errorf("protocol/ReadClientMsg: packet header is nil")
	}

	if err := body.Deserialize(p.Payload); err != nil {
		return nil, fmt.Errorf("protocol/ReadMsg: error deserializing body: %w", err)
	}

	payload, err := ioutil.ReadAll(p.Payload)

	if err != nil {
		return nil, fmt.Errorf("protocol/ReadMsg: error reading payload: %w", err)
	}

	msg := &StructMessage{
		Header:  header,
		Body:    body,
		Payload: payload,
	}

	return msg, nil
}

func (p *Packet) ReadClientMsg(body StructMessageBody) (*ClientStructMessage, error) {
	header, ok := p.Header.(*ClientStructMessageHeader)

	if !ok {
		return nil, fmt.Errorf("protocol/ReadClientMsg: invalid packet header %T", p.Header)
	}

	if header == nil {
		return nil, fmt.Errorf("protocol/ReadClientMsg: packet header is nil")
	}

	if err := body.Deserialize(p.Payload); err != nil {
		return nil, fmt.Errorf("protocol/ReadClientMsg: error deserializing body: %w", err)
	}

	payload, err := ioutil.ReadAll(p.Payload)

	if err != nil {
		return nil, fmt.Errorf("protocol/ReadClientMsg: error reading payload: %w", err)
	}

	msg := &ClientStructMessage{
		Header:  header,
		Body:    body,
		Payload: payload,
	}

	return msg, nil
}

func (p *Packet) ReadProtoMsg(body proto.Message) (*ProtoMessage, error) {
	header, ok := p.Header.(*ProtoMessageHeader)

	if !ok {
		return nil, fmt.Errorf("protocol/ReadProtoMsg: invalid packet header %T", p.Header)
	}

	if header == nil {
		return nil, fmt.Errorf("protocol/ReadProtoMsg: packet header is nil")
	}

	payload, err := ioutil.ReadAll(p.Payload)

	if err != nil {
		return nil, fmt.Errorf("protocol/ReadProtoMsg: error reading payload: %w", err)
	}

	if err := proto.Unmarshal(payload, body); err != nil {
		return nil, fmt.Errorf("protocol/ReadProtoMsg: error unmarshaling protobuf message: %w", err)
	}

	msg := &ProtoMessage{
		Header: header,
		Body:   body,
	}

	return msg, nil
}
