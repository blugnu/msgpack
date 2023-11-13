package msgpack

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestEncodeMap(t *testing.T) {
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
		{n: 0, expect: expect{header: []byte{atomEmptyMap}}},
		{n: 1, expect: expect{header: []byte{maskFixMap | byte(1)}}},
		{n: 15, expect: expect{header: []byte{maskFixMap | byte(15)}}},
		{n: 16, expect: expect{header: []byte{typeMap16, 0x00, 0x10}}},
		{n: 65535, expect: expect{header: []byte{typeMap16, 0xff, 0xff}}},
		{n: 65536, expect: expect{header: []byte{typeMap32, 0x00, 0x01, 0x00, 0x00}}},
		{n: 1 << 24, expect: expect{header: []byte{typeMap32, 0x01, 0x00, 0x00, 0x00}}, skip: !*allTests},
		{n: (1 << 32) - 1, expect: expect{header: []byte{typeMap32, 0x01, 0x00, 0x00, 0x00}}, skip: true}, // NOTE: this test cannot be run by passing -all; it must be explicitly set to skip: false
		{errorState: true, n: 0, expect: expect{error: encerr}},
		{errorState: true, n: 1, expect: expect{error: encerr}},
		{errorState: true, n: 15, expect: expect{error: encerr}},
		{errorState: true, n: 16, expect: expect{error: encerr}},
		{errorState: true, n: 65535, expect: expect{error: encerr}},
		{errorState: true, n: 65536, expect: expect{error: encerr}},
		{errorState: true, n: 16777216, expect: expect{error: encerr}, skip: !*allTests},
		{errorState: true, n: 4294967295, expect: expect{error: encerr}, skip: true}, // NOTE: this test cannot be run by passing -all; it must be explicitly set to skip: false
	}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%d entries, error = %v", tc.n, tc.errorState), func(t *testing.T) {
			if tc.skip {
				t.Skip("skipping slow test")
			}
			defer buf.Reset()
			defer func() { _ = enc.ResetError() }()

			// ARRANGE
			if tc.errorState {
				enc.err = encerr
			}
			m := make(map[string]int, tc.n)
			for i := 0; i < tc.n; i++ {
				m[fmt.Sprintf("%.12d", i)] = 0
			}

			// ACT
			err := EncodeMap(enc, m, nil)

			// ASSERT
			testError(t, tc.expect.error, err)

			t.Run("map header", func(t *testing.T) {
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
				got := (buf.Len() - len(tc.header)) / 14 // 14 bytes per entry: 13 for each string key (1 byte header, 12 bytes character data) and 1 byte for a zero-value int value
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

		// map ranging order is not guaranteed so in this test we record the first key encoded
		// by the map encoder function and immediately return an error; no more map entries
		// should be encoded so we should encode a map with a length of 3 followed by only 1
		// item consisting of a key:value pair => encoded:encoded

		encoded := byte(0) // values in the test are in the fixed int range, i.e. a single byte

		// ACT
		err := EncodeMap(enc, map[int]int{1: 1, 2: 2, 3: 3}, func(enc Encoder, k int, v int) error {
			_ = enc.Encode(k)
			_ = enc.Encode(v)
			encoded = byte(k)
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

		t.Run("writes expected items", func(t *testing.T) {
			wanted := []byte{maskFixMap | byte(3), encoded, encoded} // length =3 , but only 1 k:v pair of fixed ints of value encoded
			got := buf.Bytes()
			if !bytes.Equal(wanted, got) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

}
