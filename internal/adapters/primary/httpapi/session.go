package httpapi

import (
	"crypto/rand"
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
	deviceTTL    = 7 * 24 * time.Hour
)

func (s *Server) signed(value string) string { return s.signer.Seal(value) }

func (s *Server) secureCookies() bool { return strings.HasPrefix(s.baseURL, "https://") }

func (s *Server) newDeviceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "dev_" + hex.EncodeToString(b)
}

func (s *Server) deviceFrom(r *http.Request) (string, bool) {
	c, err := r.Cookie(deviceCookie)
	if err != nil {
		return "", false
	}
	value, ok := s.signer.Verify(c.Value)
	if !ok || !strings.HasPrefix(value, "dev_") {
		return "", false
	}
	return value, true
}

func (s *Server) ensureDevice(w http.ResponseWriter, r *http.Request) string {
	if id, ok := s.deviceFrom(r); ok {
		return id
	}
	id := s.newDeviceID()
	s.setCookie(w, deviceCookie, s.signed(id), deviceTTL)
	return id
}

func hostCookieName(code string) string {
	return "fana_host_" + strings.ReplaceAll(code, "-", "_")
}

func (s *Server) setHostCookie(w http.ResponseWriter, code string) {
	value := code + "|" + strconv.FormatInt(time.Now().Add(hostTTL).Unix(), 10)
	s.setCookie(w, hostCookieName(code), s.signed(value), hostTTL)
}

func (s *Server) setTVRoomCookie(w http.ResponseWriter, code string) {
	s.setCookie(w, "fana_tv_room", s.signed(code), 24*time.Hour)
}

func (s *Server) setCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
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
	value, ok := s.signer.Verify(c.Value)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

func (s *Server) isHost(r *http.Request, code string) bool {
	c, err := r.Cookie(hostCookieName(code))
	if err != nil {
		return false
	}
	value, ok := s.signer.Verify(c.Value)
	if !ok {
		return false
	}
	parts := strings.Split(value, "|")
	if len(parts) != 2 || parts[0] != code {
		return false
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	return err == nil && time.Now().Unix() <= exp
}

func (s *Server) roleFor(r *http.Request, code string) string {
	if !s.isHost(r, code) {
		return "guest"
	}
	if r.URL.Query().Get("client") == "tv" {
		return "tv"
	}
	return "phone"
}

func (s *Server) clientIP(r *http.Request) string {
	if s.trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if first := strings.TrimSpace(strings.Split(xff, ",")[0]); first != "" {
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
