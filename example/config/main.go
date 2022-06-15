package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/foomo/keel"
	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
)

type (
	Nested struct {
		Int             int    `yaml:"int"`
		Bool            bool   `yaml:"bool"`
		String          string `yaml:"string"`
		CamelCaseString string `yaml:"camelCaseString"`
		SnakeCaseString string `yaml:"snake_case_string"`
	}
	Config struct {
		Int    int    `yaml:"int"`
		Bool   bool   `yaml:"bool"`
		String string `yaml:"string"`
		Nested Nested `yaml:"nested"`
	}
)

func main() {
	// set env vars to override e.g. example.string
	_ = os.Setenv("ENV_STRING", "bar")
	_ = os.Setenv("ENV_DURATION", "240h")
	_ = os.Setenv("STRUCT_STRING", "bar")
	_ = os.Setenv("STRUCT_NESTED_STRING", "bar")
	_ = os.Setenv("STRUCT_NESTED_CAMELCASESTRING", "bar")
	_ = os.Setenv("STRUCT_NESTED_SNAKE_CASE_STRING", "bar")

	svr := keel.NewServer(
		keel.WithHTTPViperService(true),
	)

	l := svr.Logger()

	c := svr.Config()

	// get config vars by key
	intFn := config.GetInt(c, "env.int", 1)
	boolFn := config.GetBool(c, "env.bool", true)
	stringFn := config.GetString(c, "env.string", "foo")
	durationFn := config.GetDuration(c, "env.duration", time.Minute)

	structFn, err := config.GetStruct(c, "struct", Config{
		Int:    66,
		Bool:   true,
		String: "foo",
		Nested: Nested{
			Int:             13,
			Bool:            false,
			String:          "foo",
			CamelCaseString: "foo",
			SnakeCaseString: "foo",
		},
	})
	log.Must(l, err, "failed to create struct config fn")

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var structCfg Config
		if err := structFn(&structCfg); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf("intCfg: %d\n", intFn())))
		_, _ = w.Write([]byte(fmt.Sprintf("boolCfg: %v\n", boolFn())))
		_, _ = w.Write([]byte(fmt.Sprintf("stringCfg: %s\n", stringFn())))
		_, _ = w.Write([]byte(fmt.Sprintf("durationCfg: %.2f\n", durationFn().Hours())))
		_, _ = w.Write([]byte(fmt.Sprintf("stuctCfg: %s\n", spew.Sdump(structCfg))))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.Run()
}

// curl localhost:8080
// localhost:9300/config
