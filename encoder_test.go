package msgpack

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func NewTestEncoder() (Encoder, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return NewEncoder(buf), buf
}

func TestEncoder(t *testing.T) {
	// ARRANGE
	enc, buf := NewTestEncoder()
	encerr := errors.New("encoder error")

	type expect struct {
		result []byte
		error
		panic error
	}

	testcases := []struct {
		spec       string // for information only, not part of the test
		errorState bool   // true if the test case runs with the encoder in an error state
		fn         func() error
		expect
	}{
		// Encode
		{spec: "Encode(struct{})", fn: func() error { return enc.Encode(struct{}{}) }, expect: expect{panic: ErrUnsupportedType}},
		{spec: "Encode(nil)", fn: func() error { return enc.Encode(nil) }, expect: expect{result: []byte{atomNil}}},
		{spec: "Encode(true)", fn: func() error { return enc.Encode(true) }, expect: expect{result: []byte{atomTrue}}},
		{spec: "Encode(false)", fn: func() error { return enc.Encode(false) }, expect: expect{result: []byte{atomFalse}}},
		{spec: "Encode(int(0))", fn: func() error { return enc.Encode(int(0)) }, expect: expect{result: []byte{0x00}}},
		{spec: "Encode(int8(127))", fn: func() error { return enc.Encode(int8(127)) }, expect: expect{result: []byte{0x7f}}},
		{spec: "Encode(int16(32767))", fn: func() error { return enc.Encode(int16(32767)) }, expect: expect{result: []byte{typeInt16, 0x7f, 0xff}}},
		{spec: "Encode(int32(2147483647))", fn: func() error { return enc.Encode(int32(2147483647)) }, expect: expect{result: []byte{typeInt32, 0x7f, 0xff, 0xff, 0xff}}},
		{spec: "Encode(int64(9223372036854775807))", fn: func() error { return enc.Encode(int64(9223372036854775807)) }, expect: expect{result: []byte{typeUint64, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "Encode(uint(0))", fn: func() error { return enc.Encode(uint(0)) }, expect: expect{result: []byte{0x00}}},
		{spec: "Encode(uint8(255))", fn: func() error { return enc.Encode(uint8(255)) }, expect: expect{result: []byte{typeUint8, 0xff}}},
		{spec: "Encode(uint16(65535))", fn: func() error { return enc.Encode(uint16(65535)) }, expect: expect{result: []byte{typeUint16, 0xff, 0xff}}},
		{spec: "Encode(uint32(4294967295))", fn: func() error { return enc.Encode(uint32(4294967295)) }, expect: expect{result: []byte{typeUint32, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "Encode(uint64(18446744073709551615))", fn: func() error { return enc.Encode(uint64(18446744073709551615)) }, expect: expect{result: []byte{typeUint64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "Encode(-32)", fn: func() error { return enc.Encode(-32) }, expect: expect{result: []byte{0xe0}}},
		{spec: "Encode(-33)", fn: func() error { return enc.Encode(-33) }, expect: expect{result: []byte{typeInt8, 0xdf}}},
		{spec: "Encode(-128)", fn: func() error { return enc.Encode(-128) }, expect: expect{result: []byte{typeInt8, 0x80}}},
		{spec: "Encode(-129)", fn: func() error { return enc.Encode(-129) }, expect: expect{result: []byte{typeInt16, 0xff, 0x7f}}},
		{spec: "Encode(-32768)", fn: func() error { return enc.Encode(-32768) }, expect: expect{result: []byte{typeInt16, 0x80, 0x00}}},
		{spec: "Encode(-32769)", fn: func() error { return enc.Encode(-32769) }, expect: expect{result: []byte{typeInt32, 0xff, 0xff, 0x7f, 0xff}}},
		{spec: "Encode(-2147483648)", fn: func() error { return enc.Encode(-2147483648) }, expect: expect{result: []byte{typeInt32, 0x80, 0x00, 0x00, 0x00}}},
		{spec: "Encode(-2147483649)", fn: func() error { return enc.Encode(-2147483649) }, expect: expect{result: []byte{typeInt64, 0xff, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff}}},
		{spec: "Encode(-9223372036854775808)", fn: func() error { return enc.Encode(-9223372036854775808) }, expect: expect{result: []byte{typeInt64, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "Encode(float32(3.1415927))", fn: func() error { return enc.Encode(float32(3.1415927)) }, expect: expect{result: []byte{typeFloat32, 0x40, 0x49, 0x0F, 0xDB}}},
		{spec: "Encode(3.1415927)", fn: func() error { return enc.Encode(3.1415927) }, expect: expect{result: []byte{typeFloat64, 0x40, 0x09, 0x21, 0xfb, 0x5a, 0x7e, 0xd1, 0x97}}},
		{spec: "Encode([]int{1,2})", fn: func() error { return enc.Encode([]int{1, 2}) }, expect: expect{result: []byte{maskFixArray | byte(2), 0x01, 0x02}}},
		{spec: "Encode([]byte{1,2})", fn: func() error { return enc.Encode([]byte{1, 2}) }, expect: expect{result: []byte{typeBin8, 0x02, 0x01, 0x02}}},

		// bool
		{spec: "EncodeBool(true)", fn: func() error { return enc.EncodeBool(true) }, expect: expect{result: []byte{atomTrue}}},
		{spec: "EncodeBool(false)", fn: func() error { return enc.EncodeBool(false) }, expect: expect{result: []byte{atomFalse}}},
		{spec: "EncodeBool(true) (error)", errorState: true, fn: func() error { return enc.EncodeBool(true) }, expect: expect{error: encerr}},

		// int family
		// fixed int
		{spec: "EncodeFixedInt(-33)", fn: func() error { return enc.EncodeFixedInt(-33) }, expect: expect{panic: ErrValueOutOfRange}},
		{spec: "EncodeFixedInt(-32)", fn: func() error { return enc.EncodeFixedInt(-32) }, expect: expect{result: []byte{0xe0}}},
		{spec: "EncodeFixedInt(0)", fn: func() error { return enc.EncodeFixedInt(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeFixedInt(127)", fn: func() error { return enc.EncodeFixedInt(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeFixedInt(128)", fn: func() error { return enc.EncodeFixedInt(128) }, expect: expect{panic: ErrValueOutOfRange}},
		{spec: "EncodeFixedInt(-33) (error)", errorState: true, fn: func() error { return enc.EncodeFixedInt(-33) }, expect: expect{panic: ErrValueOutOfRange}},
		{spec: "EncodeFixedInt(0) (error)", errorState: true, fn: func() error { return enc.EncodeFixedInt(0) }, expect: expect{error: encerr}},
		{spec: "EncodeFixedInt(128) (error)", errorState: true, fn: func() error { return enc.EncodeFixedInt(128) }, expect: expect{panic: ErrValueOutOfRange}},
		// int8
		{spec: "EncodeInt8(-128)", fn: func() error { return enc.EncodeInt8(-128) }, expect: expect{result: []byte{typeInt8, 0x80}}},
		{spec: "EncodeInt8(-33)", fn: func() error { return enc.EncodeInt8(-33) }, expect: expect{result: []byte{typeInt8, 0xdf}}},
		{spec: "EncodeInt8(-32)", fn: func() error { return enc.EncodeInt8(-32) }, expect: expect{result: []byte{0xe0}}},
		{spec: "EncodeInt8(0)", fn: func() error { return enc.EncodeInt8(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeInt8(127)", fn: func() error { return enc.EncodeInt8(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeInt8(-128) (error)", errorState: true, fn: func() error { return enc.EncodeInt8(-128) }, expect: expect{error: encerr}},
		{spec: "EncodeInt8(-32) (error)", errorState: true, fn: func() error { return enc.EncodeInt8(-32) }, expect: expect{error: encerr}},
		{spec: "EncodeInt8(0) (error)", errorState: true, fn: func() error { return enc.EncodeInt8(0) }, expect: expect{error: encerr}},
		// int16
		{spec: "EncodeInt16(-32768)", fn: func() error { return enc.EncodeInt16(-32768) }, expect: expect{result: []byte{typeInt16, 0x80, 0x00}}},
		{spec: "EncodeInt16(-129)", fn: func() error { return enc.EncodeInt16(-129) }, expect: expect{result: []byte{typeInt16, 0xff, 0x7f}}},
		{spec: "EncodeInt16(-128)", fn: func() error { return enc.EncodeInt16(-128) }, expect: expect{result: []byte{typeInt8, 0x80}}},
		{spec: "EncodeInt16(-33)", fn: func() error { return enc.EncodeInt16(-33) }, expect: expect{result: []byte{typeInt8, 0xdf}}},
		{spec: "EncodeInt16(-32)", fn: func() error { return enc.EncodeInt16(-32) }, expect: expect{result: []byte{0xe0}}},
		{spec: "EncodeInt16(0)", fn: func() error { return enc.EncodeInt16(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeInt16(127)", fn: func() error { return enc.EncodeInt16(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeInt16(128)", fn: func() error { return enc.EncodeInt16(128) }, expect: expect{result: []byte{typeUint8, 0x80}}},
		{spec: "EncodeInt16(255)", fn: func() error { return enc.EncodeInt16(255) }, expect: expect{result: []byte{typeUint8, 0xff}}},
		{spec: "EncodeInt16(256)", fn: func() error { return enc.EncodeInt16(256) }, expect: expect{result: []byte{typeInt16, 0x01, 0x00}}},
		{spec: "EncodeInt16(32767)", fn: func() error { return enc.EncodeInt16(32767) }, expect: expect{result: []byte{typeInt16, 0x7f, 0xff}}},
		{spec: "EncodeInt16(-32768) (error)", errorState: true, fn: func() error { return enc.EncodeInt16(-32768) }, expect: expect{error: encerr}},
		{spec: "EncodeInt16(-128) (error)", errorState: true, fn: func() error { return enc.EncodeInt16(-128) }, expect: expect{error: encerr}},
		{spec: "EncodeInt16(-32) (error)", errorState: true, fn: func() error { return enc.EncodeInt16(-32) }, expect: expect{error: encerr}},
		{spec: "EncodeInt16(0) (error)", errorState: true, fn: func() error { return enc.EncodeInt16(0) }, expect: expect{error: encerr}},
		{spec: "EncodeInt16(127) (error)", errorState: true, fn: func() error { return enc.EncodeInt16(127) }, expect: expect{error: encerr}},
		{spec: "EncodeInt16(32767) (error)", errorState: true, fn: func() error { return enc.EncodeInt16(32767) }, expect: expect{error: encerr}},
		// int32
		{spec: "EncodeInt32(-2147483648)", fn: func() error { return enc.EncodeInt32(-2147483648) }, expect: expect{result: []byte{typeInt32, 0x80, 0x00, 0x00, 0x00}}},
		{spec: "EncodeInt32(-32769)", fn: func() error { return enc.EncodeInt32(-32769) }, expect: expect{result: []byte{typeInt32, 0xff, 0xff, 0x7f, 0xff}}},
		{spec: "EncodeInt32(-32768)", fn: func() error { return enc.EncodeInt32(-32768) }, expect: expect{result: []byte{typeInt16, 0x80, 0x00}}},
		{spec: "EncodeInt32(-129)", fn: func() error { return enc.EncodeInt32(-129) }, expect: expect{result: []byte{typeInt16, 0xff, 0x7f}}},
		{spec: "EncodeInt32(-128)", fn: func() error { return enc.EncodeInt32(-128) }, expect: expect{result: []byte{typeInt8, 0x80}}},
		{spec: "EncodeInt32(-33)", fn: func() error { return enc.EncodeInt32(-33) }, expect: expect{result: []byte{typeInt8, 0xdf}}},
		{spec: "EncodeInt32(-32)", fn: func() error { return enc.EncodeInt32(-32) }, expect: expect{result: []byte{0xe0}}},
		{spec: "EncodeInt32(0)", fn: func() error { return enc.EncodeInt32(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeInt32(127)", fn: func() error { return enc.EncodeInt32(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeInt32(128)", fn: func() error { return enc.EncodeInt32(128) }, expect: expect{result: []byte{typeUint8, 0x80}}},
		{spec: "EncodeInt32(255)", fn: func() error { return enc.EncodeInt32(255) }, expect: expect{result: []byte{typeUint8, 0xff}}},
		{spec: "EncodeInt32(256)", fn: func() error { return enc.EncodeInt32(256) }, expect: expect{result: []byte{typeUint16, 0x01, 0x00}}},
		{spec: "EncodeInt32(65535)", fn: func() error { return enc.EncodeInt32(65535) }, expect: expect{result: []byte{typeUint16, 0xff, 0xff}}},
		{spec: "EncodeInt32(65536)", fn: func() error { return enc.EncodeInt32(65536) }, expect: expect{result: []byte{typeInt32, 0x00, 0x01, 0x00, 0x00}}},
		{spec: "EncodeInt32(2147483647)", fn: func() error { return enc.EncodeInt32(2147483647) }, expect: expect{result: []byte{typeInt32, 0x7f, 0xff, 0xff, 0xff}}},
		{spec: "EncodeInt32(-2147483648) (error)", errorState: true, fn: func() error { return enc.EncodeInt32(-2147483648) }, expect: expect{error: encerr}},
		{spec: "EncodeInt32(-32768) (error)", errorState: true, fn: func() error { return enc.EncodeInt32(-32768) }, expect: expect{error: encerr}},
		{spec: "EncodeInt32(-128) (error)", errorState: true, fn: func() error { return enc.EncodeInt32(-128) }, expect: expect{error: encerr}},
		{spec: "EncodeInt32(-32) (error)", errorState: true, fn: func() error { return enc.EncodeInt32(-32) }, expect: expect{error: encerr}},
		{spec: "EncodeInt32(127) (error)", errorState: true, fn: func() error { return enc.EncodeInt32(127) }, expect: expect{error: encerr}},
		{spec: "EncodeInt32(32767) (error)", errorState: true, fn: func() error { return enc.EncodeInt32(32767) }, expect: expect{error: encerr}},
		{spec: "EncodeInt32(2147483647) (error)", errorState: true, fn: func() error { return enc.EncodeInt32(2147483647) }, expect: expect{error: encerr}},
		// int64
		{spec: "EncodeInt64(-9223372036854775808)", fn: func() error { return enc.EncodeInt64(-9223372036854775808) }, expect: expect{result: []byte{typeInt64, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "EncodeInt64(-2147483649)", fn: func() error { return enc.EncodeInt64(-2147483649) }, expect: expect{result: []byte{typeInt64, 0xff, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff}}},
		{spec: "EncodeInt64(-2147483648)", fn: func() error { return enc.EncodeInt64(-2147483648) }, expect: expect{result: []byte{typeInt32, 0x80, 0x00, 0x00, 0x00}}},
		{spec: "EncodeInt64(-32769)", fn: func() error { return enc.EncodeInt64(-32769) }, expect: expect{result: []byte{typeInt32, 0xff, 0xff, 0x7f, 0xff}}},
		{spec: "EncodeInt64(-32768)", fn: func() error { return enc.EncodeInt64(-32768) }, expect: expect{result: []byte{typeInt16, 0x80, 0x00}}},
		{spec: "EncodeInt64(-129)", fn: func() error { return enc.EncodeInt64(-129) }, expect: expect{result: []byte{typeInt16, 0xff, 0x7f}}},
		{spec: "EncodeInt64(-128)", fn: func() error { return enc.EncodeInt64(-128) }, expect: expect{result: []byte{typeInt8, 0x80}}},
		{spec: "EncodeInt64(-33)", fn: func() error { return enc.EncodeInt64(-33) }, expect: expect{result: []byte{typeInt8, 0xdf}}},
		{spec: "EncodeInt64(-32)", fn: func() error { return enc.EncodeInt64(-32) }, expect: expect{result: []byte{0xe0}}},
		{spec: "EncodeInt64(0)", fn: func() error { return enc.EncodeInt64(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeInt64(127)", fn: func() error { return enc.EncodeInt64(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeInt64(128)", fn: func() error { return enc.EncodeInt64(128) }, expect: expect{result: []byte{typeUint8, 0x80}}},
		{spec: "EncodeInt64(255)", fn: func() error { return enc.EncodeInt64(255) }, expect: expect{result: []byte{typeUint8, 0xff}}},
		{spec: "EncodeInt64(256)", fn: func() error { return enc.EncodeInt64(256) }, expect: expect{result: []byte{typeUint16, 0x01, 0x00}}},
		{spec: "EncodeInt64(65535)", fn: func() error { return enc.EncodeInt64(65535) }, expect: expect{result: []byte{typeUint16, 0xff, 0xff}}},
		{spec: "EncodeInt64(65536)", fn: func() error { return enc.EncodeInt64(65536) }, expect: expect{result: []byte{typeUint32, 0x00, 0x01, 0x00, 0x00}}},
		{spec: "EncodeInt64(4294967295)", fn: func() error { return enc.EncodeInt64(4294967295) }, expect: expect{result: []byte{typeUint32, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "EncodeInt64(4294967296)", fn: func() error { return enc.EncodeInt64(4294967296) }, expect: expect{result: []byte{typeUint64, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "EncodeInt64(9223372036854775807)", fn: func() error { return enc.EncodeInt64(9223372036854775807) }, expect: expect{result: []byte{typeUint64, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "EncodeInt64(-9223372036854775808) (error)", errorState: true, fn: func() error { return enc.EncodeInt64(-9223372036854775808) }, expect: expect{error: encerr}},
		{spec: "EncodeInt64(-2147483648) (error)", errorState: true, fn: func() error { return enc.EncodeInt64(-2147483648) }, expect: expect{error: encerr}},
		{spec: "EncodeInt64(-32768) (error)", errorState: true, fn: func() error { return enc.EncodeInt64(-32768) }, expect: expect{error: encerr}},
		{spec: "EncodeInt64(-128) (error)", errorState: true, fn: func() error { return enc.EncodeInt64(-128) }, expect: expect{error: encerr}},
		{spec: "EncodeInt64(-32) (error)", errorState: true, fn: func() error { return enc.EncodeInt64(-32) }, expect: expect{error: encerr}},
		{spec: "EncodeInt64(127) (error)", errorState: true, fn: func() error { return enc.EncodeInt64(127) }, expect: expect{error: encerr}},
		{spec: "EncodeInt64(32767) (error)", errorState: true, fn: func() error { return enc.EncodeInt64(32767) }, expect: expect{error: encerr}},
		{spec: "EncodeInt64(2147483647) (error)", errorState: true, fn: func() error { return enc.EncodeInt64(2147483647) }, expect: expect{error: encerr}},
		{spec: "EncodeInt64(9223372036854775807) (error)", errorState: true, fn: func() error { return enc.EncodeInt64(9223372036854775807) }, expect: expect{error: encerr}},
		// int
		{spec: "EncodeInt(-9223372036854775808)", fn: func() error { return enc.EncodeInt(-9223372036854775808) }, expect: expect{result: []byte{typeInt64, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "EncodeInt(-2147483649)", fn: func() error { return enc.EncodeInt(-2147483649) }, expect: expect{result: []byte{typeInt64, 0xff, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff}}},
		{spec: "EncodeInt(-2147483648)", fn: func() error { return enc.EncodeInt(-2147483648) }, expect: expect{result: []byte{typeInt32, 0x80, 0x00, 0x00, 0x00}}},
		{spec: "EncodeInt(-32769)", fn: func() error { return enc.EncodeInt(-32769) }, expect: expect{result: []byte{typeInt32, 0xff, 0xff, 0x7f, 0xff}}},
		{spec: "EncodeInt(-32768)", fn: func() error { return enc.EncodeInt(-32768) }, expect: expect{result: []byte{typeInt16, 0x80, 0x00}}},
		{spec: "EncodeInt(-129)", fn: func() error { return enc.EncodeInt(-129) }, expect: expect{result: []byte{typeInt16, 0xff, 0x7f}}},
		{spec: "EncodeInt(-128)", fn: func() error { return enc.EncodeInt(-128) }, expect: expect{result: []byte{typeInt8, 0x80}}},
		{spec: "EncodeInt(-33)", fn: func() error { return enc.EncodeInt(-33) }, expect: expect{result: []byte{typeInt8, 0xdf}}},
		{spec: "EncodeInt(-32)", fn: func() error { return enc.EncodeInt(-32) }, expect: expect{result: []byte{0xe0}}},
		{spec: "EncodeInt(0)", fn: func() error { return enc.EncodeInt(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeInt(127)", fn: func() error { return enc.EncodeInt(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeInt(128)", fn: func() error { return enc.EncodeInt(128) }, expect: expect{result: []byte{typeUint8, 0x80}}},
		{spec: "EncodeInt(32767)", fn: func() error { return enc.EncodeInt(32767) }, expect: expect{result: []byte{typeUint16, 0x7f, 0xff}}},
		{spec: "EncodeInt(32768)", fn: func() error { return enc.EncodeInt(32768) }, expect: expect{result: []byte{typeUint16, 0x80, 0x00}}},
		{spec: "EncodeInt(2147483647)", fn: func() error { return enc.EncodeInt(2147483647) }, expect: expect{result: []byte{typeUint32, 0x7f, 0xff, 0xff, 0xff}}},
		{spec: "EncodeInt(2147483648)", fn: func() error { return enc.EncodeInt(2147483648) }, expect: expect{result: []byte{typeUint32, 0x80, 0x00, 0x00, 0x00}}},
		{spec: "EncodeInt(4294967295)", fn: func() error { return enc.EncodeInt(4294967295) }, expect: expect{result: []byte{typeUint32, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "EncodeInt(4294967296)", fn: func() error { return enc.EncodeInt(4294967296) }, expect: expect{result: []byte{typeUint64, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "EncodeInt(9223372036854775807)", fn: func() error { return enc.EncodeInt(9223372036854775807) }, expect: expect{result: []byte{typeUint64, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "EncodeInt(-9223372036854775808) (error)", errorState: true, fn: func() error { return enc.EncodeInt(-9223372036854775808) }, expect: expect{error: encerr}},
		{spec: "EncodeInt(-2147483648) (error)", errorState: true, fn: func() error { return enc.EncodeInt(-2147483648) }, expect: expect{error: encerr}},
		{spec: "EncodeInt(-32768) (error)", errorState: true, fn: func() error { return enc.EncodeInt(-32768) }, expect: expect{error: encerr}},
		{spec: "EncodeInt(-128) (error)", errorState: true, fn: func() error { return enc.EncodeInt(-128) }, expect: expect{error: encerr}},
		{spec: "EncodeInt(-32) (error)", errorState: true, fn: func() error { return enc.EncodeInt(-32) }, expect: expect{error: encerr}},
		{spec: "EncodeInt(127) (error)", errorState: true, fn: func() error { return enc.EncodeInt(127) }, expect: expect{error: encerr}},
		{spec: "EncodeInt(32767) (error)", errorState: true, fn: func() error { return enc.EncodeInt(32767) }, expect: expect{error: encerr}},
		{spec: "EncodeInt(2147483647) (error)", errorState: true, fn: func() error { return enc.EncodeInt(2147483647) }, expect: expect{error: encerr}},
		{spec: "EncodeInt(9223372036854775807) (error)", errorState: true, fn: func() error { return enc.EncodeInt(9223372036854775807) }, expect: expect{error: encerr}},
		// uint8
		{spec: "EncodeUint8(0)", fn: func() error { return enc.EncodeUint8(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeUint8(127)", fn: func() error { return enc.EncodeUint8(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeUint8(128)", fn: func() error { return enc.EncodeUint8(128) }, expect: expect{result: []byte{typeUint8, 0x80}}},
		{spec: "EncodeUint8(255)", fn: func() error { return enc.EncodeUint8(255) }, expect: expect{result: []byte{typeUint8, 0xff}}},
		{spec: "EncodeUint8(0) (error)", errorState: true, fn: func() error { return enc.EncodeUint8(0) }, expect: expect{error: encerr}},
		{spec: "EncodeUint8(255) (error)", errorState: true, fn: func() error { return enc.EncodeUint8(255) }, expect: expect{error: encerr}},
		// uint16
		{spec: "EncodeUint16(0)", fn: func() error { return enc.EncodeUint16(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeUint16(127)", fn: func() error { return enc.EncodeUint16(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeUint16(128)", fn: func() error { return enc.EncodeUint16(128) }, expect: expect{result: []byte{typeUint8, 0x80}}},
		{spec: "EncodeUint16(255)", fn: func() error { return enc.EncodeUint16(255) }, expect: expect{result: []byte{typeUint8, 0xff}}},
		{spec: "EncodeUint16(256)", fn: func() error { return enc.EncodeUint16(256) }, expect: expect{result: []byte{typeUint16, 0x01, 0x00}}},
		{spec: "EncodeUint16(65535)", fn: func() error { return enc.EncodeUint16(65535) }, expect: expect{result: []byte{typeUint16, 0xff, 0xff}}},
		{spec: "EncodeUint16(0) (error)", errorState: true, fn: func() error { return enc.EncodeUint16(0) }, expect: expect{error: encerr}},
		{spec: "EncodeUint16(255) (error)", errorState: true, fn: func() error { return enc.EncodeUint16(255) }, expect: expect{error: encerr}},
		{spec: "EncodeUint16(65535) (error)", errorState: true, fn: func() error { return enc.EncodeUint16(65535) }, expect: expect{error: encerr}},
		// uint32
		{spec: "EncodeUint32(0)", fn: func() error { return enc.EncodeUint32(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeUint32(127)", fn: func() error { return enc.EncodeUint32(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeUint32(128)", fn: func() error { return enc.EncodeUint32(128) }, expect: expect{result: []byte{typeUint8, 0x80}}},
		{spec: "EncodeUint32(255)", fn: func() error { return enc.EncodeUint32(255) }, expect: expect{result: []byte{typeUint8, 0xff}}},
		{spec: "EncodeUint32(256)", fn: func() error { return enc.EncodeUint32(256) }, expect: expect{result: []byte{typeUint16, 0x01, 0x00}}},
		{spec: "EncodeUint32(65535)", fn: func() error { return enc.EncodeUint32(65535) }, expect: expect{result: []byte{typeUint16, 0xff, 0xff}}},
		{spec: "EncodeUint32(65536)", fn: func() error { return enc.EncodeUint32(65536) }, expect: expect{result: []byte{typeUint32, 0x00, 0x01, 0x00, 0x00}}},
		{spec: "EncodeUint32(4294967295)", fn: func() error { return enc.EncodeUint32(4294967295) }, expect: expect{result: []byte{typeUint32, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "EncodeUint32(0) (error)", errorState: true, fn: func() error { return enc.EncodeUint32(0) }, expect: expect{error: encerr}},
		{spec: "EncodeUint32(255) (error)", errorState: true, fn: func() error { return enc.EncodeUint32(255) }, expect: expect{error: encerr}},
		{spec: "EncodeUint32(65535) (error)", errorState: true, fn: func() error { return enc.EncodeUint32(65535) }, expect: expect{error: encerr}},
		{spec: "EncodeUint32(4294967295) (error)", errorState: true, fn: func() error { return enc.EncodeUint32(4294967295) }, expect: expect{error: encerr}},
		// uint64
		{spec: "EncodeUint64(0)", fn: func() error { return enc.EncodeUint64(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeUint64(127)", fn: func() error { return enc.EncodeUint64(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeUint64(128)", fn: func() error { return enc.EncodeUint64(128) }, expect: expect{result: []byte{typeUint8, 0x80}}},
		{spec: "EncodeUint64(255)", fn: func() error { return enc.EncodeUint64(255) }, expect: expect{result: []byte{typeUint8, 0xff}}},
		{spec: "EncodeUint64(256)", fn: func() error { return enc.EncodeUint64(256) }, expect: expect{result: []byte{typeUint16, 0x01, 0x00}}},
		{spec: "EncodeUint64(65535)", fn: func() error { return enc.EncodeUint64(65535) }, expect: expect{result: []byte{typeUint16, 0xff, 0xff}}},
		{spec: "EncodeUint64(65536)", fn: func() error { return enc.EncodeUint64(65536) }, expect: expect{result: []byte{typeUint32, 0x00, 0x01, 0x00, 0x00}}},
		{spec: "EncodeUint64(4294967295)", fn: func() error { return enc.EncodeUint64(4294967295) }, expect: expect{result: []byte{typeUint32, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "EncodeUint64(4294967296)", fn: func() error { return enc.EncodeUint64(4294967296) }, expect: expect{result: []byte{typeUint64, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "EncodeUint64(18446744073709551615)", fn: func() error { return enc.EncodeUint64(18446744073709551615) }, expect: expect{result: []byte{typeUint64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "EncodeUint64(0) (error)", errorState: true, fn: func() error { return enc.EncodeUint64(0) }, expect: expect{error: encerr}},
		{spec: "EncodeUint64(255) (error)", errorState: true, fn: func() error { return enc.EncodeUint64(255) }, expect: expect{error: encerr}},
		{spec: "EncodeUint64(65535) (error)", errorState: true, fn: func() error { return enc.EncodeUint64(65535) }, expect: expect{error: encerr}},
		{spec: "EncodeUint64(4294967295) (error)", errorState: true, fn: func() error { return enc.EncodeUint64(4294967295) }, expect: expect{error: encerr}},
		{spec: "EncodeUint64(18446744073709551615) (error)", errorState: true, fn: func() error { return enc.EncodeUint64(18446744073709551615) }, expect: expect{error: encerr}},
		// uint
		{spec: "EncodeUint(0)", fn: func() error { return enc.EncodeUint(0) }, expect: expect{result: []byte{0x00}}},
		{spec: "EncodeUint(127)", fn: func() error { return enc.EncodeUint(127) }, expect: expect{result: []byte{0x7f}}},
		{spec: "EncodeUint(128)", fn: func() error { return enc.EncodeUint(128) }, expect: expect{result: []byte{typeUint8, 0x80}}},
		{spec: "EncodeUint(255)", fn: func() error { return enc.EncodeUint(255) }, expect: expect{result: []byte{typeUint8, 0xff}}},
		{spec: "EncodeUint(256)", fn: func() error { return enc.EncodeUint(256) }, expect: expect{result: []byte{typeUint16, 0x01, 0x00}}},
		{spec: "EncodeUint(65535)", fn: func() error { return enc.EncodeUint(65535) }, expect: expect{result: []byte{typeUint16, 0xff, 0xff}}},
		{spec: "EncodeUint(65536)", fn: func() error { return enc.EncodeUint(65536) }, expect: expect{result: []byte{typeUint32, 0x00, 0x01, 0x00, 0x00}}},
		{spec: "EncodeUint(4294967295)", fn: func() error { return enc.EncodeUint(4294967295) }, expect: expect{result: []byte{typeUint32, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "EncodeUint(4294967296)", fn: func() error { return enc.EncodeUint(4294967296) }, expect: expect{result: []byte{typeUint64, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "EncodeUint(18446744073709551615)", fn: func() error { return enc.EncodeUint(18446744073709551615) }, expect: expect{result: []byte{typeUint64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "EncodeUint(0) (error)", errorState: true, fn: func() error { return enc.EncodeUint(0) }, expect: expect{error: encerr}},
		{spec: "EncodeUint(255) (error)", errorState: true, fn: func() error { return enc.EncodeUint(255) }, expect: expect{error: encerr}},
		{spec: "EncodeUint(65535) (error)", errorState: true, fn: func() error { return enc.EncodeUint(65535) }, expect: expect{error: encerr}},
		{spec: "EncodeUint(4294967295) (error)", errorState: true, fn: func() error { return enc.EncodeUint(4294967295) }, expect: expect{error: encerr}},
		{spec: "EncodeUint(18446744073709551615) (error)", errorState: true, fn: func() error { return enc.EncodeUint(18446744073709551615) }, expect: expect{error: encerr}},

		// float family
		// float32
		{spec: "EncodeFloat32(0)", fn: func() error { return enc.EncodeFloat32(0) }, expect: expect{result: []byte{typeFloat32, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "EncodeFloat32(1.5)", fn: func() error { return enc.EncodeFloat32(1.5) }, expect: expect{result: []byte{typeFloat32, 0x3f, 0xc0, 0x00, 0x00}}},
		{spec: "EncodeFloat32(3.141592653589793)", fn: func() error { return enc.EncodeFloat32(3.141592653589793) }, expect: expect{result: []byte{typeFloat32, 0x40, 0x49, 0x0f, 0xdb}}},
		{spec: "EncodeFloat32(0) (error)", errorState: true, fn: func() error { return enc.EncodeFloat32(0) }, expect: expect{error: encerr}},
		{spec: "EncodeFloat32(1.5) (error)", errorState: true, fn: func() error { return enc.EncodeFloat32(1.5) }, expect: expect{error: encerr}},
		{spec: "EncodeFloat32(3.141592653589793) (error)", errorState: true, fn: func() error { return enc.EncodeFloat32(3.141592653589793) }, expect: expect{error: encerr}},
		// float64
		{spec: "EncodeFloat64(0)", fn: func() error { return enc.EncodeFloat64(0) }, expect: expect{result: []byte{typeFloat64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "EncodeFloat64(1.5)", fn: func() error { return enc.EncodeFloat64(1.5) }, expect: expect{result: []byte{typeFloat64, 0x3f, 0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "EncodeFloat64(3.141592653589793)", fn: func() error { return enc.EncodeFloat64(3.141592653589793) }, expect: expect{result: []byte{typeFloat64, 0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18}}},
		{spec: "EncodeFloat64(0) (error)", errorState: true, fn: func() error { return enc.EncodeFloat64(0) }, expect: expect{error: encerr}},
		{spec: "EncodeFloat64(1.5) (error)", errorState: true, fn: func() error { return enc.EncodeFloat64(1.5) }, expect: expect{error: encerr}},
		{spec: "EncodeFloat64(3.141592653589793) (error)", errorState: true, fn: func() error { return enc.EncodeFloat64(3.141592653589793) }, expect: expect{error: encerr}},

		// binary data
		// []byte
		{spec: "EncodeBytes(nil)", fn: func() error { return enc.EncodeBytes(nil) }, expect: expect{result: []byte{atomNil}}},
		{spec: "EncodeBytes([]byte{})", fn: func() error { return enc.EncodeBytes([]byte{}) }, expect: expect{result: []byte{typeBin8, 0x00}}},
		{spec: "EncodeBytes([]byte{}) (error)", errorState: true, fn: func() error { return enc.EncodeBytes([]byte{}) }, expect: expect{error: encerr}},

		// compound types (arrays, maps, strings)
		// beginX tests here / encodeX tested separately
		// begin array
		{spec: "WriteArrayHeader(0)", fn: func() error { return enc.WriteArrayHeader(0) }, expect: expect{result: []byte{0x90}}},
		{spec: "WriteArrayHeader(1)", fn: func() error { return enc.WriteArrayHeader(1) }, expect: expect{result: []byte{0x91}}},
		{spec: "WriteArrayHeader(15)", fn: func() error { return enc.WriteArrayHeader(15) }, expect: expect{result: []byte{0x9f}}},
		{spec: "WriteArrayHeader(16)", fn: func() error { return enc.WriteArrayHeader(16) }, expect: expect{result: []byte{0xdc, 0x00, 0x10}}},
		{spec: "WriteArrayHeader(65535)", fn: func() error { return enc.WriteArrayHeader(65535) }, expect: expect{result: []byte{0xdc, 0xff, 0xff}}},
		{spec: "WriteArrayHeader(65536)", fn: func() error { return enc.WriteArrayHeader(65536) }, expect: expect{result: []byte{0xdd, 0x00, 0x01, 0x00, 0x00}}},
		{spec: "WriteArrayHeader(4294967295)", fn: func() error { return enc.WriteArrayHeader(4294967295) }, expect: expect{result: []byte{0xdd, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "WriteArrayHeader(0) (error)", errorState: true, fn: func() error { return enc.WriteArrayHeader(0) }, expect: expect{error: encerr}},
		{spec: "WriteArrayHeader(1) (error)", errorState: true, fn: func() error { return enc.WriteArrayHeader(1) }, expect: expect{error: encerr}},
		{spec: "WriteArrayHeader(15) (error)", errorState: true, fn: func() error { return enc.WriteArrayHeader(15) }, expect: expect{error: encerr}},
		{spec: "WriteArrayHeader(16) (error)", errorState: true, fn: func() error { return enc.WriteArrayHeader(16) }, expect: expect{error: encerr}},
		{spec: "WriteArrayHeader(65535) (error)", errorState: true, fn: func() error { return enc.WriteArrayHeader(65535) }, expect: expect{error: encerr}},
		{spec: "WriteArrayHeader(65536) (error)", errorState: true, fn: func() error { return enc.WriteArrayHeader(65536) }, expect: expect{error: encerr}},
		{spec: "WriteArrayHeader(4294967295) (error)", errorState: true, fn: func() error { return enc.WriteArrayHeader(4294967295) }, expect: expect{error: encerr}},
		// begin map
		{spec: "WriteMapHeader(0)", fn: func() error { return enc.WriteMapHeader(0) }, expect: expect{result: []byte{0x80}}},
		{spec: "WriteMapHeader(1)", fn: func() error { return enc.WriteMapHeader(1) }, expect: expect{result: []byte{0x81}}},
		{spec: "WriteMapHeader(15)", fn: func() error { return enc.WriteMapHeader(15) }, expect: expect{result: []byte{0x8f}}},
		{spec: "WriteMapHeader(16)", fn: func() error { return enc.WriteMapHeader(16) }, expect: expect{result: []byte{0xde, 0x00, 0x10}}},
		{spec: "WriteMapHeader(65535)", fn: func() error { return enc.WriteMapHeader(65535) }, expect: expect{result: []byte{0xde, 0xff, 0xff}}},
		{spec: "WriteMapHeader(65536)", fn: func() error { return enc.WriteMapHeader(65536) }, expect: expect{result: []byte{0xdf, 0x00, 0x01, 0x00, 0x00}}},
		{spec: "WriteMapHeader(4294967295)", fn: func() error { return enc.WriteMapHeader(4294967295) }, expect: expect{result: []byte{0xdf, 0xff, 0xff, 0xff, 0xff}}},
		{spec: "WriteMapHeader(0) (error)", errorState: true, fn: func() error { return enc.WriteMapHeader(0) }, expect: expect{error: encerr}},
		{spec: "WriteMapHeader(1) (error)", errorState: true, fn: func() error { return enc.WriteMapHeader(1) }, expect: expect{error: encerr}},
		{spec: "WriteMapHeader(15) (error)", errorState: true, fn: func() error { return enc.WriteMapHeader(15) }, expect: expect{error: encerr}},
		{spec: "WriteMapHeader(16) (error)", errorState: true, fn: func() error { return enc.WriteMapHeader(16) }, expect: expect{error: encerr}},
		{spec: "WriteMapHeader(65535) (error)", errorState: true, fn: func() error { return enc.WriteMapHeader(65535) }, expect: expect{error: encerr}},
		{spec: "WriteMapHeader(65536) (error)", errorState: true, fn: func() error { return enc.WriteMapHeader(65536) }, expect: expect{error: encerr}},
		{spec: "WriteMapHeader(4294967295) (error)", errorState: true, fn: func() error { return enc.WriteMapHeader(4294967295) }, expect: expect{error: encerr}},
		// begin string
		{spec: "WriteStringHeader(0)", fn: func() error { return enc.WriteStringHeader(0) }, expect: expect{result: []byte{0b10100000}}},
		{spec: "WriteStringHeader(1)", fn: func() error { return enc.WriteStringHeader(1) }, expect: expect{result: []byte{0b10100001}}},
		{spec: "WriteStringHeader(31)", fn: func() error { return enc.WriteStringHeader(31) }, expect: expect{result: []byte{0b10111111}}},
		{spec: "WriteStringHeader(32)", fn: func() error { return enc.WriteStringHeader(32) }, expect: expect{result: []byte{0xd9, 0b00100000}}},
		{spec: "WriteStringHeader(255)", fn: func() error { return enc.WriteStringHeader(255) }, expect: expect{result: []byte{0xd9, 0b11111111}}},
		{spec: "WriteStringHeader(256)", fn: func() error { return enc.WriteStringHeader(256) }, expect: expect{result: []byte{0xda, 0b00000001, 0b00000000}}},
		{spec: "WriteStringHeader(65535)", fn: func() error { return enc.WriteStringHeader(65535) }, expect: expect{result: []byte{0xda, 0b11111111, 0b11111111}}},
		{spec: "WriteStringHeader(65536)", fn: func() error { return enc.WriteStringHeader(65536) }, expect: expect{result: []byte{0xdb, 0b00000000, 0b00000001, 0b00000000, 0b00000000}}},
		{spec: "WriteStringHeader(16777216)", fn: func() error { return enc.WriteStringHeader(16777216) }, expect: expect{result: []byte{0xdb, 0b00000001, 0b00000000, 0b00000000, 0b00000000}}},
		{spec: "WriteStringHeader(4294967295)", fn: func() error { return enc.WriteStringHeader(4294967295) }, expect: expect{result: []byte{0xdb, 0b11111111, 0b11111111, 0b11111111, 0b11111111}}},
		{spec: "WriteStringHeader(0) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(0) }, expect: expect{error: encerr}},
		{spec: "WriteStringHeader(1) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(1) }, expect: expect{error: encerr}},
		{spec: "WriteStringHeader(31) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(31) }, expect: expect{error: encerr}},
		{spec: "WriteStringHeader(32) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(32) }, expect: expect{error: encerr}},
		{spec: "WriteStringHeader(255) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(255) }, expect: expect{error: encerr}},
		{spec: "WriteStringHeader(256) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(256) }, expect: expect{error: encerr}},
		{spec: "WriteStringHeader(65535) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(65535) }, expect: expect{error: encerr}},
		{spec: "WriteStringHeader(65536) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(65536) }, expect: expect{error: encerr}},
		{spec: "WriteStringHeader(16777216) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(16777216) }, expect: expect{error: encerr}},
		{spec: "WriteStringHeader(4294967295) (error)", errorState: true, fn: func() error { return enc.WriteStringHeader(4294967295) }, expect: expect{error: encerr}},

		// low level writer
		// write (byte)
		{spec: "Write(byte(0x01))", fn: func() error { return enc.Write(byte(0x01)) }, expect: expect{result: []byte{0x01}}},
		{spec: "Write(byte(0x01)) (error)", errorState: true, fn: func() error { return enc.Write(byte(0x01)) }, expect: expect{error: encerr}},
		// write ([]byte)
		{spec: "Write([]byte{0x01, 0x02, 0x03})", fn: func() error { return enc.Write([]byte{0x01, 0x02, 0x03}) }, expect: expect{result: []byte{0x01, 0x02, 0x03}}},
		{spec: "Write([]byte{0x01, 0x02, 0x03}) (error)", errorState: true, fn: func() error { return enc.Write([]byte{0x01, 0x02, 0x03}) }, expect: expect{error: encerr}},
		// write (int family)
		{spec: "Write(int(0))", fn: func() error { return enc.Write(int(0)) }, expect: expect{panic: ErrUnsupportedType}},
		{spec: "Write(int8(0))", fn: func() error { return enc.Write(int8(0)) }, expect: expect{result: []byte{0x00}}},
		{spec: "Write(int16(256))", fn: func() error { return enc.Write(int16(256)) }, expect: expect{result: []byte{0x01, 0x00}}},
		{spec: "Write(int32(65536))", fn: func() error { return enc.Write(int32(65536)) }, expect: expect{result: []byte{0x00, 0x01, 0x00, 0x00}}},
		{spec: "Write(int64(4294967296))", fn: func() error { return enc.Write(int64(4294967296)) }, expect: expect{result: []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}},
		{spec: "Write(uint(0))", fn: func() error { return enc.Write(uint(0)) }, expect: expect{panic: ErrUnsupportedType}},
		{spec: "Write(uint8(0))", fn: func() error { return enc.Write(uint8(0)) }, expect: expect{result: []byte{0x00}}},
		{spec: "Write(uint16(65280))", fn: func() error { return enc.Write(uint16(65280)) }, expect: expect{result: []byte{0xff, 0x00}}},
		{spec: "Write(uint32(4294901760))", fn: func() error { return enc.Write(uint32(4294901760)) }, expect: expect{result: []byte{0xff, 0xff, 0x00, 0x00}}},
		{spec: "Write(uint64(18446744069414584320))", fn: func() error { return enc.Write(uint64(18446744069414584320)) }, expect: expect{result: []byte{0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00}}},
		// write (bool)
		{spec: "Write(true)", fn: func() error { return enc.Write(true) }, expect: expect{panic: ErrUnsupportedType}},
	}
	for _, tc := range testcases {
		t.Run(tc.spec, func(t *testing.T) {
			defer buf.Reset()
			defer func() { _ = enc.ResetError() }()

			// ARRANGE
			if tc.errorState {
				enc.err = encerr
			}
			defer testPanic(t, tc.expect.panic)

			// ACT
			err := tc.fn()

			// ASSERT
			testError(t, tc.expect.error, err)

			t.Run("result", func(t *testing.T) {
				wanted := tc.result
				got := buf.Bytes()
				if !bytes.Equal(wanted, got) {
					t.Errorf("\nwanted: %x\ngot:    %x", wanted, got)
				}
			})
		})
	}

	t.Run("EncodeBytes", func(t *testing.T) {
		// ARRANGE
		type expect struct {
			lead []byte
			error
		}
		testcases := []struct {
			spec       string
			errorState bool
			len        int
			expect
			skip bool
		}{
			// binary data
			// []byte
			{spec: "EncodeBytes([]byte{})", len: 0, expect: expect{lead: []byte{typeBin8, 0x00}}},
			{spec: "EncodeBytes([]byte{..x255})", len: 255, expect: expect{lead: []byte{typeBin8, 0xff}}},
			{spec: "EncodeBytes([]byte{..x256})", len: 256, expect: expect{lead: []byte{typeBin16, 0x01, 0x00}}},
			{spec: "EncodeBytes([]byte{..x65535})", len: 65535, expect: expect{lead: []byte{typeBin16, 0xff, 0xff}}},
			{spec: "EncodeBytes([]byte{..x65536})", len: 65536, expect: expect{lead: []byte{typeBin32, 0x00, 0x01, 0x00, 0x00}}},
			{spec: "EncodeBytes([]byte{..x16777216})", len: 16777216, expect: expect{lead: []byte{typeBin32, 0x01, 0x00, 0x00, 0x00}}, skip: !*allTests},
			{spec: "EncodeBytes([]byte{..x4294967295})", len: 4294967295, expect: expect{lead: []byte{typeBin32, 0xff, 0xff, 0xff, 0xff}}, skip: true}, // NOTE: this test cannot be run by passing -all; it must be explicitly set to skip: false
			{spec: "EncodeBytes([]byte{}) (error)", errorState: true, len: 0, expect: expect{error: encerr}},
			{spec: "EncodeBytes([]byte{..x255}) (error)", errorState: true, len: 255, expect: expect{error: encerr}},
			{spec: "EncodeBytes([]byte{..x256}) (error)", errorState: true, len: 256, expect: expect{error: encerr}},
			{spec: "EncodeBytes([]byte{..x65535}) (error)", errorState: true, len: 65535, expect: expect{error: encerr}},
			{spec: "EncodeBytes([]byte{..x65536}) (error)", errorState: true, len: 65536, expect: expect{error: encerr}},
			{spec: "EncodeBytes([]byte{..x16777216}) (error)", errorState: true, len: 16777216, expect: expect{error: encerr}, skip: !*allTests},
			{spec: "EncodeBytes([]byte{..x4294967295}) (error)", errorState: true, len: 4294967295, expect: expect{error: encerr}, skip: true}, // NOTE: this test cannot be run by passing -all; it must be explicitly set to skip: false
		}
		for _, tc := range testcases {
			t.Run(tc.spec, func(t *testing.T) {
				if tc.skip {
					return
				}
				defer buf.Reset()
				defer func() { _ = enc.ResetError() }()

				// ARRANGE
				if tc.errorState {
					enc.err = encerr
				}

				b := bytes.Repeat([]byte{0x01}, tc.len)

				// ACT
				err := enc.EncodeBytes(b)

				// ASSERT
				testError(t, tc.expect.error, err)

				t.Run("encodes as", func(t *testing.T) {
					wanted := tc.lead
					got := buf.Bytes()[:len(wanted)]
					if !bytes.Equal(wanted, got) {
						t.Errorf("\nwanted %v\ngot    %v", wanted, got)
					}
				})
			})
		}
	})

	t.Run("EncodeString", func(t *testing.T) {
		// ARRANGE
		type expect struct {
			lead []byte
			error
		}
		testcases := []struct {
			spec       string // information only, not part of the test
			errorState bool
			len        int64
			expect
			skip bool
		}{
			{spec: "EncodeString(0)", len: 0, expect: expect{lead: []byte{0b10100000}}},
			{spec: "EncodeString(1)", len: 1, expect: expect{lead: []byte{0b10100001}}},
			{spec: "EncodeString(31)", len: 31, expect: expect{lead: []byte{0b10111111}}},
			{spec: "EncodeString(32)", len: 32, expect: expect{lead: []byte{0xd9, 0b00100000}}},
			{spec: "EncodeString(255)", len: 255, expect: expect{lead: []byte{0xd9, 0b11111111}}},
			{spec: "EncodeString(256)", len: 256, expect: expect{lead: []byte{0xda, 0b00000001, 0b00000000}}},
			{spec: "EncodeString(65535)", len: 65535, expect: expect{lead: []byte{0xda, 0b11111111, 0b11111111}}},
			{spec: "EncodeString(65536)", len: 65536, expect: expect{lead: []byte{0xdb, 0b00000000, 0b00000001, 0b00000000, 0b00000000}}},
			{spec: "EncodeString(16777216)", len: 16777216, expect: expect{lead: []byte{0xdb, 0b00000001, 0b00000000, 0b00000000, 0b00000000}}, skip: !*allTests},
			{spec: "EncodeString(4294967295)", len: 4294967295, expect: expect{lead: []byte{0xdb, 0b11111111, 0b11111111, 0b11111111, 0b11111111}}, skip: true}, // NOTE: this test cannot be run by passing -all; it must be explicitly set to skip: false
			{spec: "EncodeString(0)", errorState: true, len: 0, expect: expect{error: encerr}},
			{spec: "EncodeString(1)", errorState: true, len: 1, expect: expect{error: encerr}},
			{spec: "EncodeString(31)", errorState: true, len: 31, expect: expect{error: encerr}},
			{spec: "EncodeString(32)", errorState: true, len: 32, expect: expect{error: encerr}},
			{spec: "EncodeString(255)", errorState: true, len: 255, expect: expect{error: encerr}},
			{spec: "EncodeString(256)", errorState: true, len: 256, expect: expect{error: encerr}},
			{spec: "EncodeString(65535)", errorState: true, len: 65535, expect: expect{error: encerr}},
			{spec: "EncodeString(65536)", errorState: true, len: 65536, expect: expect{error: encerr}},
			{spec: "EncodeString(16777216)", errorState: true, len: 16777216, skip: !*allTests, expect: expect{error: encerr}},
			{spec: "EncodeString(4294967295)", errorState: true, len: 4294967295, expect: expect{error: encerr}, skip: true}, // NOTE: this test cannot be run by passing -all; it must be explicitly set to skip: false
		}
		for _, tc := range testcases {
			t.Run(fmt.Sprintf("%s, error %v", tc.spec, tc.errorState), func(t *testing.T) {
				if tc.skip {
					t.Skip("skipping slow test")
				}
				defer buf.Reset()
				defer func() { _ = enc.ResetError() }()

				// ARRANGE
				if tc.errorState {
					enc.err = encerr
				}

				s := strings.Repeat("a", int(tc.len))

				// ACT
				err := enc.EncodeString(s)

				// ASSERT
				testError(t, tc.expect.error, err)

				wanted := tc.lead
				got := buf.Bytes()[:len(wanted)]
				if !bytes.Equal(wanted, got) {
					t.Errorf("\nwanted: %v + %d x 'a'\ngot   : %v + %d x 'a'", wanted, tc.len, got[:len(wanted)], len(buf.Bytes())-len(wanted))
				}
			})
		}
	})

	t.Run("ResetError", func(t *testing.T) {
		// ARRANGE
		enc.err = encerr

		// ACT
		err := enc.ResetError()

		// ASSERT
		t.Run("returns", func(t *testing.T) {
			wanted := encerr
			got := err
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})

		t.Run("clears the error", func(t *testing.T) {
			wanted := error(nil)
			got := enc.err
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("SetWriter", func(t *testing.T) {
		// ARRANGE
		enc.err = encerr
		enc.out = buf
		defer func() { enc.out = buf }()

		// ACT
		enc.SetWriter(io.Discard)

		// ASSERT
		t.Run("sets output", func(t *testing.T) {
			wanted := io.Discard
			got := enc.out
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("Using", func(t *testing.T) {
		// ARRANGE
		enc.err = nil
		enc.out = buf
		buf.Reset()
		other := &bytes.Buffer{}
		defer func() { enc.out = buf }()

		// ACT
		err := enc.Using(other, func() error {
			_ = enc.Encode(1492)
			return encerr
		})

		// ASSERT
		t.Run("returns error", func(t *testing.T) {
			wanted := encerr
			got := err
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})

		t.Run("sets encoder error", func(t *testing.T) {
			wanted := encerr
			got := enc.err
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})

		t.Run("encoded to original writer", func(t *testing.T) {
			wanted := []byte{}
			got := buf.Bytes()
			if !bytes.Equal(wanted, got) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})

		t.Run("encoded to specified writer", func(t *testing.T) {
			wanted := []byte{typeUint8, 0x05, 0xd4}
			got := other.Bytes()
			if !bytes.Equal(wanted, got) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})
}
