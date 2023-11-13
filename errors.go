package msgpack

import "errors"

var (
	ErrValueOutOfRange = errors.New("value out of range")
	ErrUnsupportedType = errors.New("unsupported type")
)
