package main

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"

	jwt2 "github.com/golang-jwt/jwt"
	"go.uber.org/zap"

	"github.com/foomo/keel"
	"github.com/foomo/keel/jwt"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/cookie"
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
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs,
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
					} else if c, err := jwtCookie.Set(w, r, token); err != nil {
						httputils.InternalServerError(l, w, r, err)
						return nil, false
					} else {
						r.AddCookie(c)
					}
					return claims, true
				}),
				// delete cookie if e.g. sth is wrong with it
				middleware.JWTWithErrorHandler(func(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) bool {
					if err := jwtCookie.Delete(w, r); err != nil {
						httputils.InternalServerError(l, w, r, err)
						return false
					}
					http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
					return false
				}),
			),
		),
	)

	svr.Run()
}
