package domain

import "crypto/subtle"

// ConstantTimeEqual compares two strings without leaking their contents
// through timing. Both lengths must match, so callers should not branch on it
// before checking lengths for secrets of unknown size.
func ConstantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
