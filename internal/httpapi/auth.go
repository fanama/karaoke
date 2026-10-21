package httpapi

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	deviceCookie = "fana_device"
	hostTTL      = 12 * time.Hour
)

func (s *Server) sign(value string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Server) signed(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value)) + "." + s.sign(value)
}

func (s *Server) verifySigned(raw string) (string, bool) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return "", false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	value := string(decoded)
	expected := s.sign(value)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[1])) != 1 {
		return "", false
	}
	return value, true
}

func (s *Server) secureCookies() bool {
	return strings.HasPrefix(s.baseURL, "https://")
}

func newDeviceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "dev_" + hex.EncodeToString(b)
}

func (s *Server) deviceFrom(r *http.Request) (string, bool) {
	c, err := r.Cookie(deviceCookie)
	if err != nil {
		return "", false
	}
	value, ok := s.verifySigned(c.Value)
	if !ok || !strings.HasPrefix(value, "dev_") {
		return "", false
	}
	return value, true
}

func (s *Server) ensureDevice(w http.ResponseWriter, r *http.Request) string {
	if id, ok := s.deviceFrom(r); ok {
		return id
	}
	id := newDeviceID()
	http.SetCookie(w, &http.Cookie{
		Name:     deviceCookie,
		Value:    s.signed(id),
		Path:     "/",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   s.secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
	return id
}

func hostCookieName(code string) string {
	return "fana_host_" + strings.ReplaceAll(code, "-", "_")
}

func (s *Server) setTVRoomCookie(w http.ResponseWriter, code string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "fana_tv_room",
		Value:    s.signed(code),
		Path:     "/",
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   s.secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) tvRoomFrom(r *http.Request) (string, bool) {
	c, err := r.Cookie("fana_tv_room")
	if err != nil {
		return "", false
	}
	value, ok := s.verifySigned(c.Value)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

func (s *Server) setHostCookie(w http.ResponseWriter, code string) {
	value := code + "|" + strconv.FormatInt(time.Now().Add(hostTTL).Unix(), 10)
	http.SetCookie(w, &http.Cookie{
		Name:     hostCookieName(code),
		Value:    s.signed(value),
		Path:     "/",
		MaxAge:   int(hostTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) isHost(r *http.Request, code string) bool {
	c, err := r.Cookie(hostCookieName(code))
	if err != nil {
		return false
	}
	value, ok := s.verifySigned(c.Value)
	if !ok {
		return false
	}
	parts := strings.Split(value, "|")
	if len(parts) != 2 || parts[0] != code {
		return false
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	return true
}

func (s *Server) roleFor(r *http.Request, code string) string {
	if s.isHost(r, code) {
		if r.URL.Query().Get("client") == "tv" {
			return "tv"
		}
		return "phone"
	}
	return "guest"
}

func (s *Server) clientIP(r *http.Request) string {
	if s.trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			first := strings.TrimSpace(strings.Split(xff, ",")[0])
			if first != "" {
				return first
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
