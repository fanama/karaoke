package httpapi

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"rsc.io/qr"

	"fana-karaoke/internal/domain"
)

func (s *Server) handleQR(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	if !s.mgr.Exists(code) {
		s.writeError(w, domain.ErrRoomNotFound)
		return
	}
	joinURL := s.origin(r) + "/r/" + code
	qrCode, err := qr.Encode(joinURL, qr.M)
	if err != nil {
		s.writeError(w, err)
		return
	}
	const quiet = 4
	size := qrCode.Size + quiet*2
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" shape-rendering="crispEdges">`, size, size)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#fff"/>`, size, size)
	b.WriteString(`<path fill="#08050f" d="`)
	for y := 0; y < qrCode.Size; y++ {
		for x := 0; x < qrCode.Size; x++ {
			if qrCode.Black(x, y) {
				fmt.Fprintf(&b, "M%d %dh1v1h-1z", x+quiet, y+quiet)
			}
		}
	}
	b.WriteString(`"/></svg>`)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(b.String()))
}

func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
