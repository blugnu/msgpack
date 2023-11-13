// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package benchmarks

import (
	"errors"
	"io"
	"testing"

	"github.com/blugnu/msgpack"
)

func Benchmark(b *testing.B) {
	b.Run("encode(256)", func(b *testing.B) {
		enc := msgpack.NewEncoder(io.Discard)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = enc.Encode(256)
			}
		})
	})
	b.Run("encodeint(256)", func(b *testing.B) {
		enc := msgpack.NewEncoder(io.Discard)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = enc.EncodeInt(256)
			}
		})
	})
	b.Run("encodeint16(256)", func(b *testing.B) {
		enc := msgpack.NewEncoder(io.Discard)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = enc.EncodeInt16(256)
			}
		})
	})
	b.Run("encodestring", func(b *testing.B) {
		enc := msgpack.NewEncoder(io.Discard)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = enc.EncodeString("tiny string, < 32 chars")
				_ = enc.EncodeString("this is a short string < 256 chars")
			}
		})
	})
	b.Run("encodemap(.., nil)", func(b *testing.B) {
		enc := msgpack.NewEncoder(io.Discard)
		data := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = msgpack.EncodeMap(enc, data, nil)
			}
		})
	})
	b.Run("encodemap(.., fn)", func(b *testing.B) {
		enc := msgpack.NewEncoder(io.Discard)
		data := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = msgpack.EncodeMap(enc, data, func(enc msgpack.Encoder, k string, v int) error {
					_ = enc.EncodeString(k)
					return enc.EncodeInt(v)
				})
			}
		})
	})
	b.Run("encode x4 + x4 error checks", func(b *testing.B) {
		enc := msgpack.NewEncoder(io.Discard)
		id := 1
		name := "foo"
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := enc.EncodeString("id"); err != nil {
					return
				}
				if err := enc.EncodeInt(id); err != nil {
					return
				}
				if err := enc.EncodeString("name"); err != nil {
					return
				}
				if err := enc.EncodeString(name); err != nil {
					return
				}
				_ = enc.ResetError()
			}
		})
	})
	b.Run("encode x4 + x1 error check", func(b *testing.B) {
		enc := msgpack.NewEncoder(io.Discard)
		id := 1
		name := "foo"
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = enc.EncodeString("id")
				_ = enc.EncodeInt(id)
				_ = enc.EncodeString("name")
				_ = enc.EncodeString(name)
				_ = enc.ResetError()
			}
		})
	})

	b.Run("logfmt", func(b *testing.B) {
		enc := msgpack.NewEncoder(io.Discard)
		_ = enc.Using(io.Discard, func() error { return errors.New("encoder error") })

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = enc.BeginMap(3)
				_ = enc.EncodeString("timestamp")
				_ = enc.EncodeString("2010-09-08:07:06:05.432100Z")
				_ = enc.EncodeString("level")
				_ = enc.EncodeString("info")
				_ = enc.EncodeString("message")
				_ = enc.EncodeString("this is a representative log message, it is quite long and contains a lot of information")
			}
		})
	})
}
