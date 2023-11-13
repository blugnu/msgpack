package msgpack

import (
	"fmt"
	"io"
	"math"
)

// Encoder provides an api for streaming msgpack data.  To obtain an
// Encoder use NewEncoder, specifying an initial io.Writer.  The
// Writer can be changed at any time using SetWriter.
//
// The Encoder type is not safe for concurrent use.
type Encoder struct {
	out io.Writer
	err error
}

// NewEncoder returns a neenc Encoder that writes to the specified
// io.Writer.
func NewEncoder(out io.Writer) Encoder {
	return Encoder{out: out}
}

// WriteArrayHeader writes the msgpack type and length of an array to the
// current writer using the most efficient msgpack encoding possible
// according to the number of elements in the array (len).
//
// This function is primarily intended for use by other Encoder
// functions and in optimised streaming scenarios where it would
// typically be immediately followed by a call (or calls) to write
// the array elements.
//
// The EncodeArray method is usually more appropriate for encoding an array.
func (enc Encoder) WriteArrayHeader(len int) error {
	switch {
	case len == 0:
		_ = enc.Write(atomEmptyArray)
	case len < 16:
		_ = enc.Write(maskFixArray | byte(len))
	case len < 65536:
		_ = enc.Write(typeArray16)
		_ = enc.Write(uint16(len))
	default:
		_ = enc.Write(typeArray32)
		_ = enc.Write(uint32(len))
	}
	return enc.err
}

// WriteMapHeader writes the msgpack type and length of a map to the
// current writer using the most efficient msgpack encoding possible
// according to the number of entries in the map (n).
//
// This function is primarily intended for use by other Encoder
// functions and in optimised streaming scenarios where it would
// typically be immediately followed by a call (or calls) to write
// the map entries.
//
// The EncodeMap method is usually more appropriate for encoding a map.
func (enc Encoder) WriteMapHeader(n int) error {
	switch {
	case n == 0:
		_ = enc.Write(atomEmptyMap)
	case n < 16:
		_ = enc.Write(maskFixMap | byte(n))
	case n < 65536:
		_ = enc.Write(typeMap16)
		_ = enc.Write(uint16(n))
	default:
		_ = enc.Write(typeMap32)
		_ = enc.Write(uint32(n))
	}
	return enc.err
}

// WriteStringHeader writes the msgpack type and length of a string to the
// current writer using the most efficient msgpack encoding possible
// according to the length specified.
//
// The length of the string must be specified in bytes, not runes.
//
// This function is primarily intended for use by other Encoder
// functions and in optimised streaming scenarios where it would
// typically be immediately followed by a call (or calls) to write
// the bytes corresponding to the string content.
//
// The EncodeString method is usually more appropriate for encoding a string.
func (enc Encoder) WriteStringHeader(len int) error {
	switch {
	case len < 32:
		_ = enc.Write(maskFixString | byte(len))
	case len < 256:
		_ = enc.Write(typeString8)
		_ = enc.Write(byte(len))
	case len < 65536:
		_ = enc.Write(typeString16)
		_ = enc.Write(uint16(len))
	default:
		_ = enc.Write(typeString32)
		_ = enc.Write(uint32(len))
	}
	return enc.err
}

// Encode writes a msgpack encoded value to the writer. The value
// can be of any type supported by the Encoder.
//
// The types supported are:
//
//   - bool
//   - int family (int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64)
//   - string
//
// TODO: float family
// TODO: binary ([]byte)
func (enc Encoder) Encode(v any) error {
	switch v := v.(type) {
	// nil
	case nil:
		return enc.Write(atomNil)

	// bool
	case bool:
		if v {
			return enc.Write(atomTrue)
		}
		return enc.Write(atomFalse)

	// int family
	case int:
		return enc.EncodeInt(v)
	case int8:
		return enc.EncodeInt8(v)
	case int16:
		return enc.EncodeInt16(v)
	case int32:
		return enc.EncodeInt32(v)
	case int64:
		return enc.EncodeInt64(v)
	case uint:
		return enc.EncodeUint(v)
	case uint8:
		return enc.EncodeUint8(v)
	case uint16:
		return enc.EncodeUint16(v)
	case uint32:
		return enc.EncodeUint32(v)
	case uint64:
		return enc.EncodeUint64(v)

	case []int:
		return EncodeArray(enc, v, func(enc Encoder, v int) error { return enc.EncodeInt(v) })

	// string
	case string:
		return enc.EncodeString(v)

	default:
		panic(fmt.Errorf("Encode: %w: %T", ErrUnsupportedType, v))
	}
}

// EncodeBool encodes a boolean value to the current Writer.
func (enc Encoder) EncodeBool(b bool) error {
	if b {
		return enc.Write(atomTrue)
	}
	return enc.Write(atomFalse)
}

// EncodeBytes encodes a []byte value to the current Writer
// as binary data.
func (enc Encoder) EncodeBytes(b []byte) error {
	if b == nil {
		return enc.Write(atomNil)
	}

	switch {
	case len(b) < 256:
		_ = enc.Write(typeBin8)
		_ = enc.Write(byte(len(b)))
		return enc.Write(b)

	case len(b) < 65536:
		_ = enc.Write(typeBin16)
		_ = enc.Write(uint16(len(b)))
		return enc.Write(b)

	default:
		_ = enc.Write(typeBin32)
		_ = enc.Write(uint32(len(b)))
		return enc.Write(b)
	}
}

// EncodeFloat32 encodes a float32 value to the current Writer.
func (enc Encoder) EncodeFloat32(f float32) error {
	_ = enc.Write(typeFloat32)
	return enc.Write(f)
}

// EncodeFloat64 encodes a float64 value to the current Writer.
func (enc Encoder) EncodeFloat64(f float64) error {
	_ = enc.Write(typeFloat64)
	return enc.Write(f)
}

// EncodeString encodes a string to the current writer.
func (enc Encoder) EncodeString(s string) error {
	if err := enc.WriteStringHeader(len(s)); err == nil {
		_, enc.err = io.WriteString(enc.out, s)
	}
	return enc.err
}

// Reset returns any error on the encoder and clears the error state.
//
// When an encoder is in the error state, any calls to write values
// will be ignored.  The encoder will remain in the error state until
// Reset is called.  An encoder is in the error state when any attempt
// to write to the current io.Writer returns an error.  The io.Writer
// error is retained until Reset is called.
//
// This enables the caller to check the error state after each call
// to Encode if desired, or to check the error state only after all
// values have been encoded:
//
//	if err := enc.Write(i1); err != nil {
//	  return err
//	}
//	if err := enc.Write(i2); err != nil {
//	  return err
//	}
//
// or alternatively:
//
//	enc.Write(i1)
//	enc.Write(i2)
//	if err := enc.Reset(); err != nil {
//	  return err
//	}
func (e *Encoder) ResetError() (err error) {
	err = e.err
	e.err = nil
	return
}

// SetWriter changes the current io.Writer of the Encoder.
func (enc *Encoder) SetWriter(out io.Writer) {
	enc.out = out
}

// Using temporarily changes the io.Writer destination for the Encoder
// while the specified function is executed.  The original io.Writer
// destination is restored after the function returns.
func (enc *Encoder) Using(dest io.Writer, fn func() error) error {
	og := enc.out
	defer func() { enc.out = og }()

	enc.out = dest
	enc.err = fn()
	return enc.err
}

// Write writes a value to the writer as big-endian raw bytes,
// with no msgpack type indicator or other encoding.
//
// This method is provided as a more efficient alternative to
// binary.Write(), optimised for handling the limited types that
// a msgpack encoder is required to write.
//
// If an error is returned when attempting to write to the Writer,
// the error is retained and returned on subsequent calls to Write
// unless/until the error is cleared by calling Reset.
//
// Write supports only a limited number of types, being intended
// for use by other Encoder functions and in specialised streaming
// scenarios. It is not intended for general use.
//
// The types supported are:
//
//   - []byte
//   - byte / uint8
//   - int8 / int16 / int32 / int64
//   - uint16 / uint32 / uint64
//   - float32 / float64
//
// The function will panic if a value of any other type is specified.
//
// To encode a []byte as msgpack encoded binary data, use EncodeBytes.
func (enc Encoder) Write(b any) error {
	if enc.err != nil {
		return enc.err
	}

	switch v := b.(type) {
	// byte family
	case uint8: // a.k.a byte
		_, enc.err = enc.out.Write([]byte{v})
	case []byte:
		_, enc.err = enc.out.Write(v)

	// int family
	case int8:
		_, enc.err = enc.out.Write([]byte{byte(v)})
	case int16:
		_, enc.err = enc.out.Write([]byte{byte(v >> 8), byte(v)})
	case uint16:
		_, enc.err = enc.out.Write([]byte{byte(v >> 8), byte(v)})
	case int32:
		_, enc.err = enc.out.Write([]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
	case uint32:
		_, enc.err = enc.out.Write([]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
	case int64:
		_, enc.err = enc.out.Write([]byte{byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
	case uint64:
		_, enc.err = enc.out.Write([]byte{byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})

	// float family
	case float32:
		b := math.Float32bits(v)
		_, enc.err = enc.out.Write([]byte{byte(b >> 24), byte(b >> 16), byte(b >> 8), byte(b)})
	case float64:
		b := math.Float64bits(v)
		_, enc.err = enc.out.Write([]byte{byte(b >> 56), byte(b >> 48), byte(b >> 40), byte(b >> 32), byte(b >> 24), byte(b >> 16), byte(b >> 8), byte(b)})

	// unsupported
	default:
		panic(fmt.Errorf("Write: %w: %T", ErrUnsupportedType, v))
	}

	return enc.err
}
