package httpapi

import "net/http"

func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/config", s.handleConfig)
	mux.HandleFunc("POST /api/rooms", s.handleCreateRoom)
	mux.HandleFunc("GET /api/tv/session", s.handleTVSession)
	mux.HandleFunc("GET /api/rooms/resolve", s.handleResolveRoom)
	mux.HandleFunc("GET /api/rooms/{code}", s.handleSnapshot)
	mux.HandleFunc("GET /api/rooms/{code}/qr.svg", s.handleQR)
	mux.HandleFunc("GET /api/search", s.handleSearch)
	mux.HandleFunc("POST /api/rooms/{code}/songs", s.handleAddSong)
	mux.HandleFunc("POST /api/rooms/{code}/songs/{entryID}/vote", s.handleVote)
	mux.HandleFunc("POST /api/rooms/{code}/skip-vote", s.handleStartSkipVote)
	mux.HandleFunc("POST /api/rooms/{code}/skip-vote/cast", s.handleCastSkipVote)
	mux.HandleFunc("POST /api/rooms/{code}/host/unlock", s.handleHostUnlock)
	mux.HandleFunc("POST /api/rooms/{code}/playback", s.handlePlayback)
	mux.HandleFunc("DELETE /api/rooms/{code}/songs/{entryID}", s.handleRemoveSong)
	mux.HandleFunc("PUT /api/rooms/{code}/order", s.handleReorder)
	mux.HandleFunc("POST /api/rooms/{code}/queue/clear", s.handleClearQueue)
	mux.HandleFunc("POST /api/rooms/{code}/queue/lock", s.handleLockQueue)
	mux.HandleFunc("POST /api/rooms/{code}/skip-vote/cancel", s.handleCancelSkipVote)
	mux.HandleFunc("POST /api/rooms/{code}/pin/rotate", s.handleRotatePIN)
	mux.HandleFunc("POST /api/rooms/{code}/end", s.handleEndRoom)
	mux.HandleFunc("GET /ws", s.handleWS)
	mux.HandleFunc("GET /assets/", s.handleAsset)
	mux.HandleFunc("GET /favicon.svg", s.handleFavicon)
	mux.HandleFunc("GET /", s.handleSPA)
}
