package kv

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

type BinaryEncoder struct {
	w io.Writer
}

func NewBinaryEncoder(w io.Writer) *BinaryEncoder {
	return &BinaryEncoder{w: w}
}

func (e *BinaryEncoder) Encode(kv KeyValue) error {
	switch kv.Type() {
	case TypeInvalid, TypeEnd, TypeWideString:
		return fmt.Errorf("cannot encode nodes of type %s", kv.Type())
	}

	if err := e.writeType(kv.Type()); err != nil {
		return err
	}

	if err := e.writeString(kv.Key()); err != nil {
		return err
	}

	switch kv.Type() {
	case TypeObject:
		for _, c := range kv.Children() {
			if err := e.Encode(c); err != nil {
				return err
			}
		}

		if err := e.writeType(TypeEnd); err != nil {
			return err
		}

		if err := e.writeType(TypeEnd); err != nil {
			return err
		}
	case TypeString:
		if err := e.writeString(kv.Value()); err != nil {
			return err
		}
	case TypeInt32, TypeColor, TypePointer:
		n, err := strconv.ParseInt(kv.Value(), 10, 32)

		if err != nil {
			return err
		}

		if err := e.writeInt32(int32(n)); err != nil {
			return err
		}
	case TypeInt64:
		n, err := strconv.ParseInt(kv.Value(), 10, 64)

		if err != nil {
			return err
		}

		if err := e.writeInt64(n); err != nil {
			return err
		}
	case TypeUint64:
		n, err := strconv.ParseUint(kv.Value(), 10, 64)

		if err != nil {
			return err
		}

		if err := e.writeUint64(n); err != nil {
			return err
		}
	case TypeFloat32:
		n, err := strconv.ParseFloat(kv.Value(), 32)

		if err != nil {
			return err
		}

		if err := e.writeFloat32(float32(n)); err != nil {
			return err
		}
	}

	return nil
}

func (e *BinaryEncoder) writeType(t Type) error {
	_, err := e.w.Write([]byte{t.Byte()})
	return err
}

func (e *BinaryEncoder) writeString(s string) error {
	data := append([]byte(s), byte(0x0))

	if _, err := e.w.Write(data); err != nil {
		return err
	}

	return nil
}

func (e *BinaryEncoder) writeInt32(n int32) error {
	return binary.Write(e.w, binary.LittleEndian, n)
}

func (e *BinaryEncoder) writeInt64(n int64) error {
	return binary.Write(e.w, binary.LittleEndian, n)
}

func (e *BinaryEncoder) writeUint64(n uint64) error {
	return binary.Write(e.w, binary.LittleEndian, n)
}

func (e *BinaryEncoder) writeFloat32(n float32) error {
	return binary.Write(e.w, binary.LittleEndian, n)
}

type TextEncoder struct {
	ident string
	w     io.Writer
}

func NewTextEncoder(w io.Writer) *TextEncoder {
	return &TextEncoder{
		ident: "  ",
		w:     w,
	}
}

func (e *TextEncoder) Encode(kv KeyValue) error {
	fmt.Fprintf(e.w, "%s%s ", e.ident, kv.Key())

	switch kv.Type() {
	case TypeObject:
		if _, err := io.WriteString(e.w, "{\n"); err != nil {
			return err
		}

		for _, c := range kv.Children() {
			enc := NewTextEncoder(e.w)
			enc.ident = e.ident + e.ident

			if err := enc.Encode(c); err != nil {
				return err
			}

			if _, err := io.WriteString(e.w, "\n"); err != nil {
				return err
			}
		}

		fmt.Fprintf(e.w, "%s}\n", e.ident)
	default:
		if _, err := io.WriteString(e.w, strconv.Quote(kv.Value())); err != nil {
			return err
		}
	}

	return nil
}
