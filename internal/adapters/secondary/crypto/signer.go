package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"fana-karaoke/internal/ports"
)

// Signer signs cookie payloads with HMAC-SHA256.
type Signer struct {
	secret []byte
}

// NewSigner builds a signer over the given secret.
func NewSigner(secret []byte) *Signer { return &Signer{secret: secret} }

// Sign returns the URL-safe MAC for a value.
func (s *Signer) Sign(value string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Seal encodes a value and appends its MAC, separated by a dot.
func (s *Signer) Seal(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value)) + "." + s.Sign(value)
}

// Verify decodes a sealed value and checks its MAC in constant time.
func (s *Signer) Verify(signed string) (string, bool) {
	parts := strings.Split(signed, ".")
	if len(parts) != 2 {
		return "", false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	value := string(decoded)
	if subtle.ConstantTimeCompare([]byte(s.Sign(value)), []byte(parts[1])) != 1 {
		return "", false
	}
	return value, true
}

var _ ports.SessionSigner = (*Signer)(nil)
