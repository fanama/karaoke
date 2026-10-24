package d1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
)

const maxResponseBytes = 8 << 20

type client struct {
	baseURL string
	token   string
	http    *http.Client
	log     *slog.Logger
}

type apiError struct {
	code    string
	status  int
	message string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("d1: %s (%d): %s", e.code, e.status, e.message)
}

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type songPage struct {
	Songs      []apiSong `json:"songs"`
	NextOffset *int      `json:"nextOffset"`
}

type apiSong struct {
	VideoID         string `json:"videoId"`
	Title           string `json:"title"`
	Channel         string `json:"channel"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	DurationSeconds int    `json:"durationSeconds"`
	SearchText      string `json:"searchText"`
	FirstSeenAt     string `json:"firstSeenAt"`
	LastSeenAt      string `json:"lastSeenAt"`
}

func newClient(baseURL, token string, httpClient *http.Client, log *slog.Logger) *client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    httpClient,
		log:     log,
	}
}

func (c *client) do(ctx context.Context, method, path string, body any, out any) error {
	req, err := c.request(ctx, method, path, body)
	if err != nil {
		return &apiError{code: codeUnreachable, status: 0, message: err.Error()}
	}
	res, err := c.http.Do(req)
	if err != nil {
		return &apiError{code: codeUnreachable, status: 0, message: err.Error()}
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return c.failure(res)
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, maxResponseBytes))
		return nil
	}
	dec := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes))
	if err := dec.Decode(out); err != nil {
		return &apiError{code: codeInvalidResponse, status: res.StatusCode, message: err.Error()}
	}
	return nil
}

func (c *client) request(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *client) failure(res *http.Response) error {
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	var envelope errorEnvelope
	_ = json.Unmarshal(raw, &envelope)
	code := envelope.Error.Code
	if code == "" {
		code = statusCode(res.StatusCode)
	}
	message := envelope.Error.Message
	if message == "" {
		message = http.StatusText(res.StatusCode)
	}
	return &apiError{code: code, status: res.StatusCode, message: message}
}

func statusCode(status int) string {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return codeAuth
	case http.StatusTooManyRequests:
		return codeLimit
	default:
		return codeUnreachable
	}
}

func (c *client) health(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, "/v1/health", nil, nil)
}

func (c *client) songs(ctx context.Context, offset, limit int) (songPage, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))
	var page songPage
	err := c.do(ctx, http.MethodGet, "/v1/songs?"+query.Encode(), nil, &page)
	return page, err
}

func (c *client) upsert(ctx context.Context, songs []domain.LibrarySong) error {
	return c.sendSongs(ctx, "/v1/songs/upsert", songs)
}

func (c *client) importSongs(ctx context.Context, songs []domain.LibrarySong) error {
	return c.sendSongs(ctx, "/v1/songs/import", songs)
}

func (c *client) sendSongs(ctx context.Context, path string, songs []domain.LibrarySong) error {
	payload := struct {
		Songs []apiSong `json:"songs"`
	}{Songs: payloadSongs(songs)}
	return c.do(ctx, http.MethodPost, path, payload, nil)
}

func payloadSongs(songs []domain.LibrarySong) []apiSong {
	out := make([]apiSong, 0, len(songs))
	for _, song := range songs {
		out = append(out, apiSong{
			VideoID: song.VideoID, Title: song.Title, Channel: song.Channel,
			ThumbnailURL: song.ThumbnailURL, DurationSeconds: song.DurationSeconds,
			SearchText: song.SearchText, FirstSeenAt: stamp(song.FirstSeenAt),
			LastSeenAt: stamp(song.LastSeenAt),
		})
	}
	return out
}

func stamp(t time.Time) string {
	if t.IsZero() {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return t.UTC().Format(time.RFC3339)
}

func isContextError(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}
