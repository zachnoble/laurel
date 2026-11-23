package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
)

const CookieName = "session_token"

var errInvalidCookie = errors.New("invalid session cookie")

func sign(secret, token string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func sessionCookie(secret, token string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    token + "." + sign(secret, token),
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func expiredCookie() *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func readToken(header http.Header, secret string) (string, error) {
	cookie, err := (&http.Request{Header: header}).Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		return "", errInvalidCookie
	}

	token, signature, ok := strings.Cut(cookie.Value, ".")
	if !ok || token == "" || signature == "" {
		return "", errInvalidCookie
	}

	expected := sign(secret, token)
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) != 1 {
		return "", errInvalidCookie
	}

	return token, nil
}
