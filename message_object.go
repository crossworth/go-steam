package steam

import (
	"github.com/13k/go-steam/kv"
)

const (
	messageObjectRootKey = "MessageObject"
)

type MessageObject struct {
	kv.KeyValue
}

func NewMessageObject() *MessageObject {
	return &MessageObject{KeyValue: kv.NewKeyValueRoot(messageObjectRootKey)}
}

func (o *MessageObject) AddObject(key string) *MessageObject {
	o.KeyValue.AddObject(key)
	return o
}

func (o *MessageObject) AddString(key, value string) *MessageObject {
	o.KeyValue.AddString(key, value)
	return o
}

func (o *MessageObject) AddInt32(key, value string) *MessageObject {
	o.KeyValue.AddInt32(key, value)
	return o
}

func (o *MessageObject) AddInt64(key, value string) *MessageObject {
	o.KeyValue.AddInt64(key, value)
	return o
}

func (o *MessageObject) AddUint64(key, value string) *MessageObject {
	o.KeyValue.AddUint64(key, value)
	return o
}

func (o *MessageObject) AddFloat32(key, value string) *MessageObject {
	o.KeyValue.AddFloat32(key, value)
	return o
}

func (o *MessageObject) AddColor(key, value string) *MessageObject {
	o.KeyValue.AddColor(key, value)
	return o
}

func (o *MessageObject) AddPointer(key, value string) *MessageObject {
	o.KeyValue.AddPointer(key, value)
	return o
}
