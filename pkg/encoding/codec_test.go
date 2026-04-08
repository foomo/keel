package encoding_test

import (
	stdpem "encoding/pem"

	"github.com/foomo/keel/pkg/encoding"
	"github.com/foomo/keel/pkg/encoding/ascii85"
	"github.com/foomo/keel/pkg/encoding/asn1"
	"github.com/foomo/keel/pkg/encoding/base32"
	"github.com/foomo/keel/pkg/encoding/base64"
	"github.com/foomo/keel/pkg/encoding/csv"
	"github.com/foomo/keel/pkg/encoding/gob"
	"github.com/foomo/keel/pkg/encoding/hex"
	"github.com/foomo/keel/pkg/encoding/json"
	"github.com/foomo/keel/pkg/encoding/pem"
	"github.com/foomo/keel/pkg/encoding/xml"
)

// Codec[T] compile-time checks.
var (
	_ encoding.Codec[any] = (*json.Codec[any])(nil)
	_ encoding.Codec[any] = (*xml.Codec[any])(nil)
	_ encoding.Codec[any] = (*gob.Codec[any])(nil)
	_ encoding.Codec[any] = (*asn1.Codec[any])(nil)

	_ encoding.Codec[[]byte]        = (*base64.Codec)(nil)
	_ encoding.Codec[[]byte]        = (*base32.Codec)(nil)
	_ encoding.Codec[[]byte]        = (*hex.Codec)(nil)
	_ encoding.Codec[[]byte]        = (*ascii85.Codec)(nil)
	_ encoding.Codec[[][]string]    = (*csv.Codec)(nil)
	_ encoding.Codec[*stdpem.Block] = (*pem.Codec)(nil)
)

// StreamCodec[T] compile-time checks.
var (
	_ encoding.StreamCodec[any] = (*json.StreamCodec[any])(nil)
	_ encoding.StreamCodec[any] = (*xml.StreamCodec[any])(nil)
	_ encoding.StreamCodec[any] = (*gob.StreamCodec[any])(nil)
	_ encoding.StreamCodec[any] = (*asn1.StreamCodec[any])(nil)

	_ encoding.StreamCodec[[]byte]        = (*base64.StreamCodec)(nil)
	_ encoding.StreamCodec[[]byte]        = (*base32.StreamCodec)(nil)
	_ encoding.StreamCodec[[]byte]        = (*hex.StreamCodec)(nil)
	_ encoding.StreamCodec[[]byte]        = (*ascii85.StreamCodec)(nil)
	_ encoding.StreamCodec[[][]string]    = (*csv.StreamCodec)(nil)
	_ encoding.StreamCodec[*stdpem.Block] = (*pem.StreamCodec)(nil)
)
