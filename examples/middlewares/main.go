package main

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"time"

	jwt2 "github.com/golang-jwt/jwt"
	"go.uber.org/zap"

	"github.com/foomo/keel"
	"github.com/foomo/keel/jwt"
	"github.com/foomo/keel/log"
	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/cookie"
	"github.com/foomo/keel/net/http/middleware"
	httputils "github.com/foomo/keel/utils/net/http"
)

func ExampleCORS() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Header.Get(keelhttp.HeaderXRequestID)))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.CORS(
				middleware.CORSWithAllowOrigins("example.com"),
				middleware.CORSWithAllowMethods(http.MethodGet, http.MethodPost),
			),
		),
	)

	svr.Run()
}

func ExampleBasicAuth() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	username := "demo"

	// create password hash
	passwordHash, err := httputils.HashBasicAuthPassword([]byte("demo"))
	log.Must(l, err, "failed to hash password")

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.BasicAuth(
				username,
				passwordHash,
				middleware.BasicAuthWithRealm("demo"),
			),
		),
	)

	svr.Run()
}

func ExampleJWTFromToken() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	contextKey := "custom"

	type CustomClaims struct {
		jwt2.StandardClaims
		Name string `json:"name"`
	}

	// generate rsa key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	log.Must(l, err, "")

	// create jwt key
	jwtKey := jwt.NewKey("demo", &rsaKey.PublicKey, rsaKey)
	log.Must(l, err, "failed to create jwt key")

	// init jwt with key files
	jwtInst := jwt.New(jwtKey)

	// custom token provider
	tokenProvider := middleware.HeaderTokenProvider(
		middleware.HeaderTokenProviderWithPrefix("Custom "),
		middleware.HeaderTokenProviderWithHeader("X-Authorization"),
	)

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// retrieve from context
		claims := r.Context().Value(contextKey).(*CustomClaims)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(claims.Name))
	})
	svs.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if token, err := jwtInst.GetSignedToken(CustomClaims{
			StandardClaims: jwt.NewStandardClaims(),
			Name:           "JWT From Token Example",
		}); err != nil {
			httputils.InternalServerError(l, w, r, err)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(token))
		}
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.Skip(
				middleware.JWT(
					jwtInst,
					contextKey,
					// require token
					middleware.JWTWithMissingTokenHandler(middleware.RequiredJWTMissingTokenHandler),
					// use custom token provider
					middleware.JWTWithTokenProvider(tokenProvider),
					// user custom claims
					middleware.JWTWithClaimsProvider(func() jwt2.Claims {
						return &CustomClaims{}
					}),
				),
				middleware.RequestURIWhitelistSkipper("/token"),
			),
		),
	)

	svr.Run()
}

func ExampleJWTFromCookie() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	contextKey := "custom"

	type CustomClaims struct {
		jwt2.StandardClaims
		Name string `json:"name"`
	}

	// define jwt cookie
	jwtCookie := cookie.New("demo")

	// generate rsa key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	log.Must(l, err, "")

	// create jwt key
	jwtKey := jwt.NewKey("demo", &rsaKey.PublicKey, rsaKey)
	log.Must(l, err, "failed to create jwt key")

	// init jwt with key files
	jwtInst := jwt.New(jwtKey, jwt.WithDeprecatedKeys())

	// custom token provider
	tokenProvider := middleware.CookieTokenProvider("keel-jwt")

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// retrieve from context
		claims := r.Context().Value(contextKey).(*CustomClaims)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(claims.Name))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.JWT(
				jwtInst,
				contextKey,
				// use custom token provider
				middleware.JWTWithTokenProvider(tokenProvider),
				// user custom claims
				middleware.JWTWithClaimsProvider(func() jwt2.Claims {
					return &CustomClaims{}
				}),
				// create cookie if missing
				middleware.JWTWithMissingTokenHandler(func(l *zap.Logger, w http.ResponseWriter, r *http.Request) (jwt2.Claims, bool) {
					claims := &CustomClaims{
						StandardClaims: jwt.NewStandardClaims(),
						Name:           "JWT From Cookie Example",
					}
					if token, err := jwtInst.GetSignedToken(claims); err != nil {
						httputils.InternalServerError(l, w, r, err)
						return nil, false
					} else if err := jwtCookie.Set(w, r, token); err != nil {
						httputils.InternalServerError(l, w, r, err)
						return nil, false
					}
					return claims, true
				}),
				// delete cookie if e.g. sth is wrong with it
				middleware.JWTWithErrorHandler(func(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) bool {
					if err := jwtCookie.Delete(w, r); err != nil {
						httputils.InternalServerError(l, w, r, err)
						return false
					}
					http.Redirect(w, r, r.URL.String(), http.StatusFound)
					return false
				}),
			),
		),
	)

	svr.Run()
}

func ExampleTokenAuthFromHeader() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	token := "some-random-token"

	// custom token provider
	tokenProvider := middleware.HeaderTokenProvider(
		middleware.HeaderTokenProviderWithPrefix("Custom "),
		middleware.HeaderTokenProviderWithHeader("X-Authorization"),
	)

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.TokenAuth(
				token,
				middleware.TokenAuthWithTokenProvider(tokenProvider),
			),
		),
	)

	svr.Run()
}

func ExampleTokenAuthFromCookie() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	token := "some-random-token"

	// custom token provider
	tokenProvider := middleware.CookieTokenProvider("keel-token")

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.TokenAuth(
				token,
				middleware.TokenAuthWithTokenProvider(tokenProvider),
			),
		),
	)

	svr.Run()
}

func ExampleTelemetry() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddServices(keel.NewDefaultServiceHTTPPrometheus())

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.Telemetry("demo"),
		),
	)

	svr.Run()
}

func ExampleLogger() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Header.Get(keelhttp.HeaderXRequestID)))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.Logger(),
		),
	)

	svr.Run()
}

func ExampleRecover() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		panic("handled")
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.Recover(
				middleware.RecoverWithDisablePrintStack(true),
			),
		),
	)

	svr.Run()
}

func ExampleRequestID() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// custom request id generator
	requestIDGenerator := func() string {
		return "my-custom-request-id"
	}

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Header.Get(keelhttp.HeaderXRequestID)))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.RequestID(
				middleware.RequestIDWithSetResponseHeader(true),
				middleware.RequestIDWithGenerator(requestIDGenerator),
			),
		),
	)

	svr.Run()
}

func ExampleSessionID() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// custom cookie time provider
	timeProvider := cookie.NewTimeProvider(
		cookie.TimeProviderWithOffset(-time.Millisecond * 500),
	)

	// custom cookie domain provider
	domainProvider := cookie.NewDomainProvider(
		cookie.DomainProviderWithDomains("foo.com", "*.bar.com"),
		cookie.DomainProviderWithMappings(map[string]string{"source.com": "target.com"}),
	)

	// custom cookie provider
	sessionCookie := cookie.New(
		"session",
		cookie.WithExpires(time.Hour*60),
		cookie.WithTimeProvider(timeProvider),
		cookie.WithDomainProvider(domainProvider),
	)

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Header.Get(keelhttp.HeaderXSessionID)))
		_, _ = w.Write([]byte(middleware.SessionIDFromContext(r.Context())))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.SessionID(
				// automatically set cookie if not exists
				middleware.SessionIDWithSetCookie(true),
				// define a custom cookie provider
				middleware.SessionIDWithCookie(sessionCookie),
			),
		),
	)

	svr.Run()
}

func ExampleSkip() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		panic("handled")
	})
	svs.HandleFunc("/skip", func(w http.ResponseWriter, r *http.Request) {
		panic("unhandled")
	})

	svr.AddServices(
		// with URI blacklist
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.Skip(
				middleware.Recover(),
				middleware.RequestURIBlacklistSkipper("/skip"),
			),
		),

		// with URI whitelist
		keel.NewServiceHTTP(l, "demo", ":8081", svs,
			middleware.Skip(
				middleware.Recover(),
				middleware.RequestURIWhitelistSkipper("/"),
			),
		),
	)

	svr.Run()
}
