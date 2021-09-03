package cmrcsession

import (
	"net/http"
	"strings"
	"time"

	cmrcjwt "github.com/bestbytes/commerce/pkg/jwt"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

type Session struct {
	jwt           *cmrcjwt.JWT
	secure        bool
	domains       []string
	domainMapping map[string]string
}

func NewSession(jwt *cmrcjwt.JWT, domains []string, domainMapping map[string]string, secure bool) *Session {
	return &Session{
		jwt:           jwt,
		secure:        secure,
		domains:       domains,
		domainMapping: domainMapping,
	}
}

func (s *Session) HasDomain(domain string) bool {
	for _, value := range s.domains {
		if domain == value || (strings.HasPrefix(value, "*.") && strings.HasSuffix(domain, value[2:])) {
			return true
		}
	}
	return false
}

func (s *Session) GetCookieClaims(r *http.Request, name string, claims jwt.Claims) (jwt.Claims, error) {
	if cookie, err := r.Cookie(name); err != nil {
		return nil, err
	} else if token, err := s.jwt.ParseWithClaims(cookie.Value, claims); err != nil {
		return nil, err
	} else if !token.Valid {
		return nil, errors.New("invalid token")
	} else {
		return token.Claims, nil
	}
}

func (s *Session) SetCookieClaims(r *http.Request, w http.ResponseWriter, name string, claims jwt.Claims, lifetime time.Duration) error {
	domain, err := s.getDomain(r)
	if err != nil {
		return err
	}

	token, err := s.jwt.GetSignedToken(claims)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		MaxAge:   int(lifetime.Seconds()),
		HttpOnly: true,
		Secure:   s.secure,
		Domain:   domain,
	})

	return nil
}

func (s *Session) SetSessionCookie(r *http.Request, w http.ResponseWriter, name string, value string) error {
	domain, err := s.getDomain(r)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secure,
		Domain:   domain,
	})

	return nil
}

func (s *Session) DeleteCookie(r *http.Request, w http.ResponseWriter, name string) error {
	domain, err := s.getDomain(r)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secure,
		Domain:   domain,
		MaxAge:   -1,
		Expires:  time.Now().AddDate(0, 0, -1),
	})

	return nil
}

func (s *Session) getDomain(r *http.Request) (string, error) {
	var domain string
	if r.Header.Get("X-Forwarded-Host") != "" {
		domain = r.Header.Get("X-Forwarded-Host")
	} else if !r.URL.IsAbs() {
		domain = r.Host
	} else {
		domain = r.URL.Host
	}

	// right trim port
	portIndex := strings.Index(domain, ":")
	if portIndex != -1 {
		domain = domain[:portIndex]
	}

	if s.domainMapping != nil {
		if value, ok := s.domainMapping[domain]; ok {
			domain = value
		}
	}

	if !s.HasDomain(domain) {
		return "", errors.New("invalid domain: " + domain)
	}

	return domain, nil
}
