package msgpack

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		len  int
		lead []byte
		skip bool
	}{
		{len: 0, lead: []byte{0b10100000}},
		{len: 1, lead: []byte{0b10100001}},
		{len: 31, lead: []byte{0b10111111}},
		{len: 32, lead: []byte{0xd9, 0b00100000}},
		{len: 255, lead: []byte{0xd9, 0b11111111}},
		{len: 256, lead: []byte{0xda, 0b00000001, 0b00000000}},
		{len: 65535, lead: []byte{0xda, 0b11111111, 0b11111111}},
		{len: 65536, lead: []byte{0xdb, 0b00000000, 0b00000001, 0b00000000, 0b00000000}},
	}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("string of length %d", tc.len), func(t *testing.T) {
			// ARRANGE
			str := strings.Repeat("a", tc.len)

			// ACT
			got := String(str)

			// ASSERT
			t.Run("lead bytes", func(t *testing.T) {
				wanted := tc.lead
				if !bytes.Equal(wanted, got[:len(tc.lead)]) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("data bytes", func(t *testing.T) {
				wanted := []byte(str)
				got := got[len(tc.lead):]
				if !bytes.Equal(wanted, got) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})
	}
}
