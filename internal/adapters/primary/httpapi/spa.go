package httpapi

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

func (s *Server) handleSPA(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	switch {
	case p == "/":
		s.writeIndex(w)
	case strings.HasPrefix(p, "/r/") && singleSegment(p):
		s.ensureDevice(w, r)
		s.writeIndex(w)
	case strings.HasPrefix(p, "/tv/") && singleSegment(p):
		s.serveTV(w, r)
	default:
		http.NotFound(w, r)
	}
}

func singleSegment(p string) bool {
	return strings.Count(strings.Trim(p, "/"), "/") == 1
}

func (s *Server) serveTV(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/tv/")
	code, ok := s.svc.RoomByHostToken(token)
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.svc.HostSeen(code)
	device := s.ensureDevice(w, r)
	s.setHostCookie(w, code)
	s.setTVRoomCookie(w, code)
	s.svc.TouchDevice(device, true)
	s.writeIndex(w)
}

func (s *Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	if s.static == nil {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(s.static, "favicon.svg")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}

var contentTypes = map[string]string{
	".js":    "text/javascript; charset=utf-8",
	".css":   "text/css; charset=utf-8",
	".svg":   "image/svg+xml",
	".png":   "image/png",
	".woff2": "font/woff2",
}

func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	if s.static == nil {
		http.NotFound(w, r)
		return
	}
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	data, err := fs.ReadFile(s.static, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	contentType, ok := contentTypes[path.Ext(name)]
	if !ok {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, _ = w.Write(data)
}
