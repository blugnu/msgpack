package msgpack

import (
	"fmt"
	"math"
)

// EncodeFixedInt writes a fixed int to the current writer. The
// function will panic with ErrOutOfRange if the value is
// out of range for a msgpack fixed int encoding.
//
// The valid range for EncodeFixedInt is -32..127 (incl.)
//
// To encode an integer outside of this range use a function
// appropriate to the type you wish to encode or EncodeInt; these
// functions all select the most efficient packing for the
// value involved.
func (enc Encoder) EncodeFixedInt(i int) error {
	switch {
	case i < int(minFixedInt),
		i > int(maxFixedInt):
		panic(fmt.Errorf("EncodeFixedInt: %d: %w: -%d..%d", i, ErrValueOutOfRange, minFixedInt, maxFixedInt))

	default:
		return enc.Write(byte(i))
	}
}

// EncodeInt8 encodes a signed 8-bit integer to the current writer.
//
// The encoder will use the most efficient format for the value
// being encoded, which may be a fixed int.
func (enc Encoder) EncodeInt8(i int8) error {
	switch {
	case i < minFixedInt:
		_ = enc.Write(typeInt8)
		return enc.Write(i)

	default: // all int8 are <= maxFixedInt:
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt
	}
}

// EncodeInt16 encodes a signed 16-bit integer to the current writer.
//
// The encoder will use the most efficient format for the value
// being encoded, which may not be int16.
func (enc Encoder) EncodeInt16(i int16) error {
	switch {
	case i < int16(math.MinInt8):
		_ = enc.Write(typeInt16)
		return enc.Write(int16(i))

	case i < int16(minFixedInt):
		_ = enc.Write(typeInt8)
		return enc.Write(int8(i))

	case i <= int16(maxFixedInt):
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt

	case i <= math.MaxUint8:
		_ = enc.Write(typeUint8)
		return enc.Write(uint8(i))

	default:
		_ = enc.Write(typeInt16)
		return enc.Write(i)
	}
}

// EncodeInt32 encodes a signed 32-bit integer to the current writer.
//
// The encoder will use the most efficient format for the value
// being encoded, which may not be int32.
func (enc Encoder) EncodeInt32(i int32) error {
	switch {
	case i < int32(math.MinInt16):
		_ = enc.Write(typeInt32)
		return enc.Write(int32(i))

	case i < int32(math.MinInt8):
		_ = enc.Write(typeInt16)
		return enc.Write(int16(i))

	case i < int32(minFixedInt):
		_ = enc.Write(typeInt8)
		return enc.Write(int8(i))

	case i <= int32(maxFixedInt):
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt

	case i <= math.MaxUint8:
		_ = enc.Write(typeUint8)
		return enc.Write(uint8(i))

	case i <= math.MaxUint16:
		_ = enc.Write(typeUint16)
		return enc.Write(uint16(i))

	default:
		_ = enc.Write(typeInt32)
		return enc.Write(i)
	}
}

// EncodeInt64 encodes a signed 64-bit integer to the current writer.
//
// The encoder will use the most efficient format for the value
// being encoded, which may not be int64.
func (enc Encoder) EncodeInt64(i int64) error {
	switch {
	case i < math.MinInt32:
		_ = enc.Write(typeInt64)
		return enc.Write(i)

	case i < math.MinInt16:
		_ = enc.Write(typeInt32)
		return enc.Write(int32(i))

	case i < math.MinInt8:
		_ = enc.Write(typeInt16)
		return enc.Write(int16(i))

	case i < int64(minFixedInt):
		_ = enc.Write(typeInt8)
		return enc.Write(int8(i))

	case i <= int64(maxFixedInt):
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt

	case i <= math.MaxUint8:
		_ = enc.Write(typeUint8)
		return enc.Write(uint8(i))

	case i <= math.MaxUint16:
		_ = enc.Write(typeUint16)
		return enc.Write(uint16(i))

	case i <= math.MaxUint32:
		_ = enc.Write(typeUint32)
		return enc.Write(uint32(i))

	default:
		_ = enc.Write(typeUint64) // keeps sonarcloud happy by not duplicating the case for < MinInt32 (positive int64/uint64 are identical)
		return enc.Write(i)
	}
}

// WriteUint8 encodes an unsigned 8-bit integer to the current writer.
//
// The encoder will use the most efficient format for the value
// being encoded: fixed int or uint8.
func (enc Encoder) EncodeUint8(i uint8) error {
	switch {
	case i <= maxFixedUint:
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt

	default:
		_ = enc.Write(typeUint8)
		return enc.Write(i)
	}
}

// EncodeUint16 encodes an unsigned 16-bit integer to the current writer.
//
// The encoder will use the most efficient format for the value
// being encoded: fixed int, uint8 or uint16.
func (enc Encoder) EncodeUint16(i uint16) error {
	switch {
	case i <= uint16(maxFixedUint):
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt``

	case i <= math.MaxUint8:
		_ = enc.Write(typeUint8)
		return enc.Write(uint8(i))

	default:
		_ = enc.Write(typeUint16)
		return enc.Write(i)
	}
}

// EncodeUint32 encodes an unsigned 32-bit integer to the current writer.
//
// The encoder will use the most efficient format for the value
// being encoded: fixed int, uint8, uint16 or uint32.
func (enc Encoder) EncodeUint32(i uint32) error {
	switch {
	case i <= uint32(maxFixedUint):
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt

	case i <= math.MaxUint8:
		_ = enc.Write(typeUint8)
		return enc.Write(uint8(i))

	case i <= math.MaxUint16:
		_ = enc.Write(typeUint16)
		return enc.Write(uint16(i))

	default:
		_ = enc.Write(typeUint32)
		return enc.Write(i)
	}
}

// EncodeUint64 encodes an unsigned 64-bit integer to the current writer.
//
// The encoder will use the most efficient format for the value
// being encoded: fixed int, uint8, uint16, uint32 or uint64.
func (enc Encoder) EncodeUint64(i uint64) error {
	switch {
	case i <= uint64(maxFixedUint):
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt

	case i <= math.MaxUint8:
		_ = enc.Write(typeUint8)
		return enc.Write(uint8(i))

	case i <= math.MaxUint16:
		_ = enc.Write(typeUint16)
		return enc.Write(uint16(i))

	case i <= math.MaxUint32:
		_ = enc.Write(typeUint32)
		return enc.Write(uint32(i))

	default:
		_ = enc.Write(typeUint64)
		return enc.Write(i)
	}
}

// EncodeInt encodes a signed integer to the current writer.
//
// The encoder packs using the smallest possible integer
// type for the value involved.
//
// To write values that exceed the MaxInt/MinInt range on a 32-bit
// platform you must explicitly use WriteInt64/WriteUint64.
func (enc Encoder) EncodeInt(i int) error {
	switch {
	case i < math.MinInt32:
		_ = enc.Write(typeInt64)
		return enc.Write(int64(i))

	case i < math.MinInt16:
		_ = enc.Write(typeInt32)
		return enc.Write(int32(i))

	case i < math.MinInt8:
		_ = enc.Write(typeInt16)
		return enc.Write(int16(i))

	case i < int(minFixedInt):
		_ = enc.Write(typeInt8)
		return enc.Write(int8(i))

	case i <= int(maxFixedInt):
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt

	case i <= math.MaxUint8:
		_ = enc.Write(typeUint8)
		return enc.Write(uint8(i))

	case i <= math.MaxUint16:
		_ = enc.Write(typeUint16)
		return enc.Write(uint16(i))

	case i <= math.MaxUint32:
		_ = enc.Write(typeUint32)
		return enc.Write(uint32(i))

	default:
		_ = enc.Write(typeUint64) // keeps sonarcloud happy by not duplicating the case for < MinInt32 (positive int64/uint64 are identical)
		return enc.Write(int64(i))
	}
}

// EncodeUint encodes an unsigned integer to the current writer.
//
// The encoder packs using the smallest possible integer
// type for the value involved.
func (enc Encoder) EncodeUint(i uint) error {
	switch {
	case i <= uint(maxFixedUint):
		return enc.Write(byte(i)) // bypass the range check in EncodeFixedInt
	case i <= math.MaxUint8:
		_ = enc.Write(typeUint8)
		return enc.Write(uint8(i))
	case i <= math.MaxUint16:
		_ = enc.Write(typeUint16)
		return enc.Write(uint16(i))
	case i <= math.MaxUint32:
		_ = enc.Write(typeUint32)
		return enc.Write(uint32(i))
	default:
		_ = enc.Write(typeUint64)
		return enc.Write(uint64(i))
	}

}
