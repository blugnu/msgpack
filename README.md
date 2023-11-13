<div align="center" style="margin-bottom:20px">
  <!-- <img src=".assets/banner.png" alt="logger" /> -->
  <div align="center">
    <a href="https://github.com/blugnu/msgpack/actions/workflows/qa.yml"><img alt="build-status" src="https://github.com/blugnu/msgpack/actions/workflows/qa.yml/badge.svg?branch=master&style=flat-square"/></a>
    <a href="https://goreportcard.com/report/github.com/blugnu/msgpack" ><img alt="go report" src="https://goreportcard.com/badge/github.com/blugnu/msgpack"/></a>
    <a><img alt="go version >= 1.18" src="https://img.shields.io/github/go-mod/go-version/blugnu/msgpack?style=flat-square"/></a>
    <a href="https://github.com/blugnu/msgpack/blob/master/LICENSE"><img alt="MIT License" src="https://img.shields.io/github/license/blugnu/msgpack?color=%234275f5&style=flat-square"/></a>
    <a href="https://coveralls.io/github/blugnu/magpack?branch=master"><img alt="coverage" src="https://img.shields.io/coveralls/github/blugnu/msgpack?style=flat-square"/></a>
    <a href="https://pkg.go.dev/github.com/blugnu/msgpack"><img alt="docs" src="https://pkg.go.dev/badge/github.com/blugnu/msgpack"/></a>
    <hr/>
  </div>
</div>

<br>

# blugnu/msgpack

Provides an efficient implementation of an encoder that may be used to stream structured data to an `io.Writer` in [`msgpack`](https://msgpack.org) format.

## Using the Encoder

A new `Encoder` is obtained using `NewEncoder()`,  supplying an initial `io.Writer` to which the encoder output is sent.  To avoid allocations of encoders when encoding to various outputs, an existing `Encoder` may be retargeted to a different `io.Writer` using the `SetWriter()` method.  To temporarily redirect output to a different `io.Writer`, the `Using()` method may be used.

`Encoder` offers high and low-level encoding functions to cater for a wide range of encoding scenarios.

The `Encode(any)` method will encode an `any` value in the most efficient manner possible according to the underlying type.  There is a small overhead using this method, due to the need to type-switch on the supplied value to determine the appropriate encoding method.

For more efficient encoding, when streaming values of known types, type-specific encoder methods may be used directly (_`EncodeBool()`, `EncodeString()` etc_) to avoid this type-switch.

Whichever encoder method is used, _all_ determine the most efficient encoding possible for the values they are given.

e.g. `Encode()`, `EncodeInt()`, `EncodeInt16()`, `EncodeInt32()` will encode the following values as described:

| value  | encoded size | msgpack encoding | description |
| -- | --| -- | -- |
| `0`    | 1 byte       | fixed int | a single byte encoding the value and type |
| `127`  | 1 bytes      | fixed int | a single byte encoding the value and type |
| `128`  | 2 bytes      | uint8 | 1 type byte + 1 byte of value encoding |
| `1024` | 3 bytes      | uint16 | 1 type byte + 2 bytes of value encoding |

## Error Handling

If an error is returned from the `io.Writer` when encoding a value the error is returned but is also captured on the `Encoder`.

Any further encoder calls will return this captured error without attempting to encode any further information to the `io.Writer`.  The `Encoder` remains in this error state until the `ResetError()` method is called.

`ResetError()` returns and clears (sets `nil`) any currently captured error.

This enables error handling to be simplified by deferring a single error check to the end of compound encoding statements.

i.e. instead of:

```go
  if err := enc.EncodeString("id"); err != nil {
    return err
  } 
  if err := enc.Encode(id); err != nil {
    return err
  } 
  if err := enc.EncodeString("name"); err != nil {
    return err
  } 
  if err := enc.Encode(name); err != nil {
    return err
  }
```

You may instead use:

```go
  _ = enc.EncodeString("id")
  _ = enc.Encode(id)
  _ = enc.EncodeString("name")
  _ = enc.Encode(name)
  return enc.ResetError()
```

> The returned error must be explicitly ignored to avoid lint problems when using this approach.

Although convenient this approach is less efficient when an error occurs; when there is no error the difference is negligible.

## `EncodeArray[T]()` / `EncodeMap[K, V]()`
These generic functions are provided to encode slices and maps.

These are not `Encoder` _methods_ but are first order functions accepting an `Encoder` parameter; this is dictated by the generics implementation in Go which (at time of writing at least) does not support generic functions on non-generic types.   

> _The `EncodeArray()` function is named to reflect the `msgpack` terminology for the encoded value; strictly speaking it is used to encode slices, not arrays._

In addition to an `Encoder`, both functions accept a `slice` or `map` to be encoded and an optional encoder function.  The encoder function is called for each item in the slice or map to encode that item.  If `nil` is specified for this function then a default encoder function is assumed, encoding items using the high-level `Encode()` method of the supplied `Encoder`.

More efficient encoding may be achieved by supplying a function which uses encoder methods appropriate to the types/values involved (to avoid type-switching in the `Encode()` method).

### Slices, Maps and Errors
If an `io.Writer` error occurs while writing the items in an slice or map, the encoder will stop processing any further items and immediately returns from the `EncodeArray()` or `EncodeMap()` function.

_**NOTE:** the `msgpack` format encodes the number of items in an array or map ahead of the items in the output stream; therefore, if an error occurs while writing the items, the `msgpack` output will be invalid._

## Using()

If you need to temporarily redirect output of an encoder to a different `io.Writer`, the `Using()` method may be used.

The `Using()` method accepts an `io.Writer` and a function to be called with the encoder.

This may be useful if you need to encode a value which may be cached for later re-use by the same encoding function, to avoid re-encoding the same values, for example:

```go
  if len(buf) == 0 {
     enc.Using(bytes.NewBuffer(buf), func(enc *Encoder) error {
        _ = enc.EncodeString("id")
        _ = enc.Encode(id)
        _ = enc.EncodeString("name")
        return enc.Encode(name)
     })
  }
  return enc.Write(buf)
```

The encoder is retargeted to the supplied `io.Writer` for the duration of the function call, after which it is retargeted to the original `io.Writer`.

If the supplied function returns an error, the encoder is retargeted to the original `io.Writer` before the error is returned.

# Decoder / Marshal / Unmarshal

_**Not currently implemented.**_

These may be implemented in the future, but at this time this module provides only an Encoder.