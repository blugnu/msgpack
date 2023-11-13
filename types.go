package msgpack

type MapEncoder[K comparable, V any] func(Encoder, K, V) error

const (
	minFixedInt  int8  = -32
	maxFixedInt  int8  = 127
	minFixedUint uint8 = 0
	maxFixedUint uint8 = 127

	// atoms are single-byte values that encode both type and value in a single
	// byte, with no following data bytes
	atomNil         byte = atomNull
	atomNull        byte = 0xc0
	atomFalse       byte = 0xc2
	atomTrue        byte = 0xc3
	atomEmptyArray  byte = 0x90 // alias for a PackFixArray with no entries
	atomEmptyMap    byte = 0x80 // alias for a PackFixMap with no entries
	atomEmptyString byte = 0xa0 // alias for a PackFixString with zero length
	atomZero        byte = 0x00 // alias for a maskFixInt with value 0

	// masks are single-byte masks that encode type and size information in a single
	// byte, to be followed by the number of data bytes indicated by the size value
	maskFixArray  byte = 0x90 // 0x90-0x9f: array with 0-15 entries
	maskFixInt    byte = 0x00 // 0x00-0x7f: positive fixint (0-127)
	maskFixMap    byte = 0x80 // 0x80-0x8f: map with 0-15 entries
	maskFixString byte = 0xa0 // 0xa0-0xbf: string with 0-31 bytes
	maskNegFixInt byte = 0xe0 // 0xe0-0xff: negative fixint (-1 to -32)

	// types are single-byte type indicators that encode a type with size
	// in following bytes

	// arrays
	typeArray16 byte = 0xdc
	typeArray32 byte = 0xdd

	// binary data
	typeBin8  byte = 0xc4
	typeBin16 byte = 0xc5
	typeBin32 byte = 0xc6

	// floats
	typeFloat32 byte = 0xca
	typeFloat64 byte = 0xcb

	// maps
	typeMap16 byte = 0xde
	typeMap32 byte = 0xdf

	// ints
	typeInt8  byte = 0xd0
	typeInt16 byte = 0xd1
	typeInt32 byte = 0xd2
	typeInt64 byte = 0xd3

	// unsigned ints
	typeUint8  byte = 0xcc
	typeUint16 byte = 0xcc
	typeUint32 byte = 0xcd
	typeUint64 byte = 0xce

	// strings
	typeString8  byte = 0xd9
	typeString16 byte = 0xda
	typeString32 byte = 0xdb
)
