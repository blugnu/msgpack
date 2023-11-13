package msgpack

// EncodeMap encodes a map to the current writer.
//
// A function may be provided to encode the key and value of each
// map entry. If no function is provided (nil), the default behaviour is
// to encode the key and value using the Encoder.Encode method.
//
// If an error is returned from the function, encoding will stop and
// the error will be returned to the caller.
func EncodeMap[K comparable, V any](enc Encoder, m map[K]V, fn MapEncoder[K, V]) error {
	if err := enc.WriteMapHeader(len(m)); err != nil {
		return err
	}

	if fn == nil {
		fn = func(enc Encoder, k K, v V) error {
			_ = enc.Encode(k)
			return enc.Encode(v)
		}
	}

	for k, v := range m {
		if enc.err != nil {
			return enc.err
		}
		enc.err = fn(enc, k, v)
	}

	return enc.err
}
