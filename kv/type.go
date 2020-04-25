package kv

type Type int8

const (
	TypeInvalid    Type = iota - 1 // -1
	TypeObject                     // 0x00
	TypeString                     // 0x01
	TypeInt32                      // 0x02
	TypeFloat32                    // 0x03
	TypePointer                    // 0x04
	TypeWideString                 // 0x05
	TypeColor                      // 0x06
	TypeUint64                     // 0x07
	TypeEnd                        // 0x08
	_                              // skip
	TypeInt64                      // 0x0a
)

func TypeFromByte(b byte) Type {
	t := Type(b)

	if t < TypeObject || t == 0x09 || t > TypeInt64 {
		return TypeInvalid
	}

	return t
}

func (t Type) Byte() byte {
	switch t {
	case TypeObject,
		TypeString,
		TypeInt32,
		TypeFloat32,
		TypePointer,
		TypeWideString,
		TypeColor,
		TypeUint64,
		TypeEnd,
		TypeInt64:
		return byte(t)
	default:
		return 0
	}
}

func (t Type) String() string {
	switch t {
	case TypeObject:
		return "Object"
	case TypeString:
		return "String"
	case TypeInt32:
		return "Int32"
	case TypeFloat32:
		return "Float32"
	case TypePointer:
		return "Pointer"
	case TypeWideString:
		return "WideString"
	case TypeColor:
		return "Color"
	case TypeUint64:
		return "Uint64"
	case TypeEnd:
		return "End"
	case TypeInt64:
		return "Int64"
	default:
		return "Invalid"
	}
}
