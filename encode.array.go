package msgpack

// EncodeArray encodes an array to the current writer.
//
// A function may be provided to encode each element of the array.
// If no function is provided (nil), the default behaviour is to encode
// each element using the Encoder.Encode method.
//
// If an error is returned from the function, encoding will stop and
// the error will be returned to the caller.
func EncodeArray[T any](enc Encoder, s []T, fn func(Encoder, T) error) error {
	if err := enc.WriteArrayHeader(len(s)); err != nil {
		return err
	}

	if fn == nil {
		fn = func(enc Encoder, v T) error {
			return enc.Encode(v)
		}
	}

	for _, v := range s {
		if enc.err != nil {
			break
		}
		enc.err = fn(enc, v)
	}

	return enc.err
}
