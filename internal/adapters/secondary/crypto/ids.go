package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"fana-karaoke/internal/ports"
)

// IDs implements ports.IDGenerator with crypto/rand.
type IDs struct{}

// NewIDs builds the generator.
func NewIDs() *IDs { return &IDs{} }

// NewID returns a prefixed random identifier.
func (IDs) NewID(prefix string) string {
	b := make([]byte, 9)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

// NewRoomCode returns an adjective-noun pair, e.g. "blue-mic".
func (IDs) NewRoomCode() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	adjective := adjectives[int(b[0])%len(adjectives)]
	noun := nouns[int(b[1])%len(nouns)]
	return adjective + "-" + noun
}

// NewHostToken returns a 32-byte URL-safe secret.
func (IDs) NewHostToken() string {
	return base64.RawURLEncoding.EncodeToString(randomBytes(32))
}

// NewPIN returns a zero-padded 4-digit PIN.
func (IDs) NewPIN() string {
	b := randomBytes(2)
	return leftPad4((int(b[0])<<8 | int(b[1])) % 10000)
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func leftPad4(n int) string {
	out := make([]byte, 4)
	for i := 3; i >= 0; i-- {
		out[i] = byte('0' + n%10)
		n /= 10
	}
	return string(out)
}

var _ ports.IDGenerator = IDs{}
