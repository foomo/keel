package keelgotsrpc

import "github.com/mitchellh/mapstructure"

// Decode decodes the custom data into the given pointer to a map or struct.
// Deprecated: use mapstructure.Decode instead.
func Decode(data interface{}, v interface{}) error {
	return mapstructure.Decode(data, v)
}
