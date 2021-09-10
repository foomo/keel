package main

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"

	jwt2 "github.com/golang-jwt/jwt"

	"github.com/foomo/keel"
	"github.com/foomo/keel/jwt"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
	httputils "github.com/foomo/keel/utils/net/http"
)

func main() {
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
