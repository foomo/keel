// Package config provides typed, lazily evaluated accessors for
// configuration values backed by [github.com/spf13/viper].
//
// A global [github.com/spf13/viper.Viper] instance, returned by [Config], reads environment
// variables automatically (with "." in keys replaced by "_") and treats
// config sources as YAML. Accessors take the *viper.Viper to read from;
// [Get], [MustGet] and [GetStruct] accept nil to select the global instance.
//
// # Getters
//
// Get* functions register a key with a default and return a getter that
// reads the current value on each call. MustGet* functions register a
// required key and panic if it is not set:
//
//	addr := config.Get(nil, "service.addr", ":8080")
//	token := config.MustGet[string](nil, "service.token")
//	fmt.Println(addr(), token())
//
// Registered keys, defaults and types can be inspected with
// [RequiredKeys], [Defaults], [Types] and [TypeOf].
//
// # Watching
//
// [Watch] and [WatchChan] (and their typed variants) poll a getter every
// second and report changes, which is useful together with remote providers
// added through [WithRemoteConfig].
package config
