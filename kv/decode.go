package kv

import (
	"bufio"
	"encoding/binary"
	"io"
	"strconv"
)

type BinaryDecoder struct {
	r  io.Reader
	br *bufio.Reader
}

func NewBinaryDecoder(r io.Reader) *BinaryDecoder {
	return &BinaryDecoder{
		r:  r,
		br: bufio.NewReader(r),
	}
}

func (d *BinaryDecoder) Decode(kv KeyValue) error {
	var (
		typ   Type
		value string
		err   error
	)

	typ, err = d.readType()

	if err != nil {
		return err
	}

	kv.SetType(typ)

	if typ == TypeEnd {
		return nil
	}

	key, err := d.readString()

	if err != nil {
		return err
	}

	kv.SetKey(key)

	switch typ {
	case TypeObject:
		var children []KeyValue

		if children, err = d.readObject(); err != nil {
			return err
		}

		kv.SetChildren(children...)
	case TypeString:
		if value, err = d.readString(); err != nil {
			return err
		}

		kv.SetValue(value)
	case TypeInt32, TypeColor, TypePointer:
		if value, err = d.readInt32String(); err != nil {
			return err
		}

		kv.SetValue(value)
	case TypeInt64:
		if value, err = d.readInt64String(); err != nil {
			return err
		}

		kv.SetValue(value)
	case TypeUint64:
		if value, err = d.readUint64String(); err != nil {
			return err
		}

		kv.SetValue(value)
	case TypeFloat32:
		if value, err = d.readFloat32String(); err != nil {
			return err
		}

		kv.SetValue(value)
	}

	return nil
}

func (d *BinaryDecoder) readObject() ([]KeyValue, error) {
	var kvs []KeyValue

	for {
		kv := NewKeyValueEmpty()

		if err := d.Decode(kv); err != nil {
			return nil, err
		}

		if kv.Type() == TypeEnd {
			break
		}

		kvs = append(kvs, kv)
	}

	return kvs, nil
}

func (d *BinaryDecoder) readType() (Type, error) {
	b, err := d.br.ReadByte()

	if err != nil {
		return 0, err
	}

	return TypeFromByte(b), nil
}

func (d *BinaryDecoder) readString() (string, error) {
	s, err := d.br.ReadString(0x0)

	if err != nil {
		return "", err
	}

	return s[:len(s)-1], nil
}

func (d *BinaryDecoder) readInt64String() (string, error) {
	var n int64

	if err := binary.Read(d.br, binary.LittleEndian, &n); err != nil {
		return "", err
	}

	return strconv.FormatInt(n, 10), nil
}

func (d *BinaryDecoder) readInt32String() (string, error) {
	var n int32

	if err := binary.Read(d.br, binary.LittleEndian, &n); err != nil {
		return "", err
	}

	return strconv.FormatInt(int64(n), 10), nil
}

func (d *BinaryDecoder) readUint64String() (string, error) {
	var n uint64

	if err := binary.Read(d.br, binary.LittleEndian, &n); err != nil {
		return "", err
	}

	return strconv.FormatUint(n, 10), nil
}

func (d *BinaryDecoder) readFloat32String() (string, error) {
	var n float32

	if err := binary.Read(d.br, binary.LittleEndian, &n); err != nil {
		return "", err
	}

	return strconv.FormatFloat(float64(n), 'f', -1, 32), nil
}
