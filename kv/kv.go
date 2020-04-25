package kv

import (
	"bytes"
	"encoding"
)

// KeyValue represents a recursive string key to arbitrary value container.
type KeyValue interface {
	// Type returns the node's Type.
	Type() Type
	// SetType sets the node's Type and returns the receiver.
	SetType(Type) KeyValue
	// Key returns the node's Key.
	Key() string
	// SetKey sets the node's Key and returns the receiver.
	SetKey(key string) KeyValue
	// Value returns the node's Value.
	Value() string
	// SetValue sets the node's Value and returns the receiver.
	SetValue(value string) KeyValue
	// Parent returns the parent node.
	Parent() KeyValue
	// SetParent sets the node's parent node and returns the receiver.
	SetParent(KeyValue) KeyValue
	// Children returns all child nodes
	Children() []KeyValue
	// SetChildren sets the node's children and returns the receiver.
	SetChildren(...KeyValue) KeyValue
	// Child finds a child node with the given key.
	Child(key string) KeyValue
	// AddChild adds a child node and returns the receiver.
	AddChild(KeyValue) KeyValue
	// AddObject adds an Object child node and returns the receiver.
	AddObject(key string) KeyValue
	// AddString adds a String child node and returns the receiver.
	AddString(key, value string) KeyValue
	// AddInt32 adds an Int32 child node and returns the receiver.
	AddInt32(key, value string) KeyValue
	// AddInt64 adds an Int64 child node and returns the receiver.
	AddInt64(key, value string) KeyValue
	// AddUint64 adds an Uint64 child node and returns the receiver.
	AddUint64(key, value string) KeyValue
	// AddFloat32 adds a Float32 child node and returns the receiver.
	AddFloat32(key, value string) KeyValue
	// AddColor adds a Color child node and returns the receiver.
	AddColor(key, value string) KeyValue
	// AddPointer adds a Pointer child node and returns the receiver.
	AddPointer(key, value string) KeyValue

	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type keyValue struct {
	typ      Type
	key      string
	value    string
	parent   KeyValue
	children []KeyValue
}

func NewKeyValue(t Type, key, value string, parent KeyValue) KeyValue {
	return &keyValue{
		typ:    t,
		key:    key,
		value:  value,
		parent: parent,
	}
}

func NewKeyValueEmpty() KeyValue {
	return NewKeyValue(TypeInvalid, "", "", nil)
}

func NewKeyValueRoot(key string) KeyValue {
	return NewKeyValueObject(key, nil)
}

func NewKeyValueObject(key string, parent KeyValue) KeyValue {
	return NewKeyValue(TypeObject, key, "", parent)
}

func NewKeyValueString(key, value string, parent KeyValue) KeyValue {
	return NewKeyValue(TypeString, key, value, parent)
}

func NewKeyValueInt32(key, value string, parent KeyValue) KeyValue {
	return NewKeyValue(TypeInt32, key, value, parent)
}

func NewKeyValueInt64(key, value string, parent KeyValue) KeyValue {
	return NewKeyValue(TypeInt64, key, value, parent)
}

func NewKeyValueUint64(key, value string, parent KeyValue) KeyValue {
	return NewKeyValue(TypeUint64, key, value, parent)
}

func NewKeyValueFloat32(key, value string, parent KeyValue) KeyValue {
	return NewKeyValue(TypeFloat32, key, value, parent)
}

func NewKeyValueColor(key, value string, parent KeyValue) KeyValue {
	return NewKeyValue(TypeColor, key, value, parent)
}

func NewKeyValuePointer(key, value string, parent KeyValue) KeyValue {
	return NewKeyValue(TypePointer, key, value, parent)
}

func (kv *keyValue) Type() Type { return kv.typ }
func (kv *keyValue) SetType(t Type) KeyValue {
	kv.typ = t
	return kv
}

func (kv *keyValue) Key() string { return kv.key }
func (kv *keyValue) SetKey(k string) KeyValue {
	kv.key = k
	return kv
}

func (kv *keyValue) Value() string { return kv.value }
func (kv *keyValue) SetValue(v string) KeyValue {
	kv.value = v
	return kv
}

func (kv *keyValue) Parent() KeyValue { return kv.parent }
func (kv *keyValue) SetParent(p KeyValue) KeyValue {
	kv.parent = p
	return kv
}

func (kv *keyValue) Children() []KeyValue { return kv.children }
func (kv *keyValue) SetChildren(children ...KeyValue) KeyValue {
	for _, c := range children {
		c.SetParent(kv)
	}

	kv.children = children

	return kv
}

func (kv *keyValue) Child(key string) KeyValue {
	for _, c := range kv.children {
		if c.Key() == key {
			return c
		}
	}

	return nil
}

func (kv *keyValue) AddChild(c KeyValue) KeyValue {
	c.SetParent(kv)

	kv.children = append(kv.children, c)

	return kv
}

func (kv *keyValue) AddObject(key string) KeyValue {
	return kv.AddChild(NewKeyValueObject(key, kv))
}

func (kv *keyValue) AddString(key, value string) KeyValue {
	return kv.AddChild(NewKeyValueString(key, value, kv))
}

func (kv *keyValue) AddInt32(key, value string) KeyValue {
	return kv.AddChild(NewKeyValueInt32(key, value, kv))
}

func (kv *keyValue) AddInt64(key, value string) KeyValue {
	return kv.AddChild(NewKeyValueInt64(key, value, kv))
}

func (kv *keyValue) AddUint64(key, value string) KeyValue {
	return kv.AddChild(NewKeyValueUint64(key, value, kv))
}

func (kv *keyValue) AddFloat32(key, value string) KeyValue {
	return kv.AddChild(NewKeyValueFloat32(key, value, kv))
}

func (kv *keyValue) AddColor(key, value string) KeyValue {
	return kv.AddChild(NewKeyValueColor(key, value, kv))
}

func (kv *keyValue) AddPointer(key, value string) KeyValue {
	return kv.AddChild(NewKeyValuePointer(key, value, kv))
}

func (kv *keyValue) MarshalText() ([]byte, error) {
	b := &bytes.Buffer{}

	if err := NewTextEncoder(b).Encode(kv); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (kv *keyValue) MarshalBinary() ([]byte, error) {
	b := &bytes.Buffer{}

	if err := NewBinaryEncoder(b).Encode(kv); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (kv *keyValue) UnmarshalBinary(data []byte) error {
	return NewBinaryDecoder(bytes.NewReader(data)).Decode(kv)
}
