package msgpack

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestEncodeArray(t *testing.T) {
	// ARRANGE
	enc, buf := NewTestEncoder()
	encerr := errors.New("encoder error")

	type expect struct {
		header []byte
		error
	}
	testcases := []struct {
		errorState bool
		n          int
		expect
		skip bool
	}{
		{n: 0, expect: expect{header: []byte{atomEmptyArray}}},
		{n: 1, expect: expect{header: []byte{maskFixArray | byte(1)}}},
		{n: 15, expect: expect{header: []byte{maskFixArray | byte(15)}}},
		{n: 16, expect: expect{header: []byte{typeArray16, 0x00, 0x10}}},
		{n: 65535, expect: expect{header: []byte{typeArray16, 0xff, 0xff}}},
		{n: 65536, expect: expect{header: []byte{typeArray32, 0x00, 0x01, 0x00, 0x00}}},
		{n: 1 << 24, expect: expect{header: []byte{typeArray32, 0x01, 0x00, 0x00, 0x00}}, skip: !*allTests},
		{n: (1 << 32) - 1, expect: expect{header: []byte{typeArray32, 0x01, 0x00, 0x00, 0x00}}, skip: true}, // NOTE: this test cannot be run by passing -all; it must be explicitly set to skip: false
		{errorState: true, n: 0, expect: expect{error: encerr}},
		{errorState: true, n: 1, expect: expect{error: encerr}},
		{errorState: true, n: 15, expect: expect{error: encerr}},
		{errorState: true, n: 16, expect: expect{error: encerr}},
		{errorState: true, n: 65535, expect: expect{error: encerr}},
		{errorState: true, n: 65536, expect: expect{error: encerr}},
		{errorState: true, n: 1 << 24, expect: expect{error: encerr}, skip: !*allTests},
		{errorState: true, n: (1 << 32) - 1, expect: expect{error: encerr}, skip: true}, // NOTE: this test cannot be run by passing -all; it must be explicitly set to skip: false},
	}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%d elements, error: %v", tc.n, tc.errorState), func(t *testing.T) {
			if tc.skip {
				t.Skip("skipping slow test")
			}
			defer buf.Reset()
			defer func() { _ = enc.ResetError() }()

			// ARRANGE
			if tc.errorState {
				enc.err = encerr
			}
			// we test using a slice of zero-value int's which will pack as single
			// bytes (fixed positive integer 0-127) enabling the written values to
			// be tested by simply comparing the overall buffer length that is written
			s := make([]int, tc.n)

			// ACT
			err := EncodeArray(enc, s, nil)

			// ASSERT
			testError(t, tc.expect.error, err)

			t.Run("array header", func(t *testing.T) {
				wanted := tc.header
				got := buf.Bytes()[:len(wanted)]
				if !bytes.Equal(wanted, got) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("value bytes", func(t *testing.T) {
				wanted := tc.n
				if tc.errorState {
					wanted = 0
				}
				got := buf.Len() - len(tc.header)
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})
	}

	t.Run("when error occurs writing items", func(t *testing.T) {
		// ARRANGE
		enc.err = nil
		buf.Reset()

		// ACT
		err := EncodeArray(enc, []int{1, 2, 3}, func(enc Encoder, v int) error {
			if v > 1 {
				return encerr
			}
			return enc.Encode(v)
		})

		// ASSERT
		t.Run("returns error", func(t *testing.T) {
			wanted := encerr
			got := err
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})

		t.Run("writes expected items", func(t *testing.T) {
			wanted := []byte{maskFixArray | byte(3), 0x01}
			got := buf.Bytes()
			if !bytes.Equal(wanted, got) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})
}
