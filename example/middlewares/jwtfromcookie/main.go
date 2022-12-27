package main

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"strings"

	jwt2 "github.com/golang-jwt/jwt"
	"go.uber.org/zap"

	"github.com/foomo/keel"
	"github.com/foomo/keel/jwt"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/cookie"
	"github.com/foomo/keel/net/http/middleware"
	httputils "github.com/foomo/keel/utils/net/http"
)

type CustomClaims struct {
	jwt2.StandardClaims
	Name     string `json:"name"`
	Language string `json:"language"`
}

const (
	ContextKey = "custom"
)

func main() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

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
	tokenProvider := middleware.CookieTokenProvider("demo")

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// retrieve from context
		if claims, ok := r.Context().Value(ContextKey).(*CustomClaims); ok {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(claims.Name + " - " + claims.Language))
		}
	})
	svs.HandleFunc("/fr", func(w http.ResponseWriter, r *http.Request) {
		// retrieve from context
		if claims, ok := r.Context().Value(ContextKey).(*CustomClaims); ok {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(claims.Name + " - " + claims.Language))
		}
	})
	svs.HandleFunc("/en", func(w http.ResponseWriter, r *http.Request) {
		// retrieve from context
		if claims, ok := r.Context().Value(ContextKey).(*CustomClaims); ok {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(claims.Name + " - " + claims.Language))
		}
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs,
			middleware.Skip(
				middleware.JWT(
					jwtInst,
					ContextKey,
					// use custom token provider
					middleware.JWTWithTokenProvider(tokenProvider),
					// user custom claims
					middleware.JWTWithClaimsProvider(func() jwt2.Claims {
						return &CustomClaims{}
					}),
					// handle existing claim
					middleware.JWTWithClaimsHandler(func(l *zap.Logger, w http.ResponseWriter, r *http.Request, claims jwt2.Claims) bool {
						if value, ok := claims.(*CustomClaims); ok {
							var language string
							switch {
							case strings.HasPrefix(r.URL.Path, "/fr"):
								language = "fr"
							case strings.HasPrefix(r.URL.Path, "/en"):
								language = "en"
							default:
								language = "de"
							}
							if value.Language != language {
								value.Language = language
								if token, err := jwtInst.GetSignedToken(claims); err != nil {
									httputils.InternalServerError(l, w, r, err)
									return false
								} else if c, err := jwtCookie.Set(w, r, token); err != nil {
									httputils.InternalServerError(l, w, r, err)
									return false
								} else {
									r.AddCookie(c)
									l.Info("updated cookie", zap.String("path", r.URL.Path))
									return true
								}
							} else {
								return true
							}
						} else {
							httputils.InternalServerError(l, w, r, err)
							return false
						}
					}),
					// create cookie if missing
					middleware.JWTWithMissingTokenHandler(func(l *zap.Logger, w http.ResponseWriter, r *http.Request) (jwt2.Claims, bool) {
						claims := &CustomClaims{
							StandardClaims: jwt.NewStandardClaims(),
							Name:           "JWT From Cookie Example",
							Language:       "de",
						}
						if token, err := jwtInst.GetSignedToken(claims); err != nil {
							httputils.InternalServerError(l, w, r, err)
							return nil, false
						} else if c, err := jwtCookie.Set(w, r, token); err != nil {
							httputils.InternalServerError(l, w, r, err)
							return nil, false
						} else {
							r.AddCookie(c)
							l.Info("added cookie", zap.String("path", r.URL.Path))
						}
						return claims, true
					}),
					// delete cookie if e.g. sth is wrong with it
					middleware.JWTWithErrorHandler(func(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) bool {
						if err := jwtCookie.Delete(w, r); err != nil {
							httputils.InternalServerError(l, w, r, err)
							return false
						}
						l.Info("deleted cookie")
						http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
						return false
					}),
				),
				middleware.RequestURIBlacklistSkipper("/favicon.ico"),
			),
		),
	)

	svr.Run()
}
