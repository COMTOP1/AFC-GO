package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type (
	Accesser struct {
		conf Config
	}
	Config struct {
		AccessCookieName string
		DomainName       string
	}
)

var (
	ErrNoToken      = errors.New("token not found")
	ErrInvalidToken = errors.New("invalid token")
)

// NewAccesser allows the validation of JWT tokens both as
// headers and as cookies
func NewAccesser(conf Config) *Accesser {
	return &Accesser{
		conf: conf,
	}
}

// GetAFCToken will return the claims from an AFC access token JWT
//
// First will check the Authorization header, if unset will
// check the access cookie
func (a *Accesser) GetAFCToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")

	if len(token) == 0 {
		cookie, err := r.Cookie(a.conf.AccessCookieName)
		if err != nil {
			if errors.As(http.ErrNoCookie, &err) {
				return "", ErrNoToken
			}
			return "", fmt.Errorf("failed to get cookie: %w", err)
		}
		token = cookie.Value
	} else {
		splitToken := strings.Split(token, "Bearer ")
		if len(splitToken) != 2 {
			return "", ErrInvalidToken
		}
		token = splitToken[1]
	}

	if token == "" {
		return "", ErrNoToken
	}
	return token, nil
}
