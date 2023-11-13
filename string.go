package msgpack

import (
	"bytes"
	"sync"
)

// sw provides a pool of Encoders used by the String() function when
// writing a string in msgpack format.
var sw = &sync.Pool{New: func() any { return &Encoder{out: &bytes.Buffer{}} }}

// String returns a []byte containing a msgpack encoded string.
func String(s string) []byte {
	enc := sw.Get().(*Encoder)
	defer sw.Put(enc)

	buf := enc.out.(*bytes.Buffer)
	buf.Reset()

	_ = enc.EncodeString(s)

	return append([]byte{}, buf.Bytes()...)
}
