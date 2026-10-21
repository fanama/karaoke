package invidious

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// Prober asks YouTube oEmbed whether a video can be embedded.
type Prober struct {
	http     *http.Client
	endpoint string
}

// NewProber builds the prober. An empty endpoint uses YouTube's public oEmbed.
func NewProber(client *http.Client, endpoint string) *Prober {
	if client == nil {
		client = &http.Client{Timeout: 6 * time.Second}
	}
	if endpoint == "" {
		endpoint = "https://www.youtube.com/oembed"
	}
	return &Prober{http: client, endpoint: endpoint}
}

// Probe returns metadata, or ErrEmbedBlocked when the owner disabled embeds.
func (p *Prober) Probe(ctx context.Context, videoID string) (domain.VideoMeta, error) {
	target := p.oembedURL(videoID)
	var payload struct {
		Title        string `json:"title"`
		AuthorName   string `json:"author_name"`
		ThumbnailURL string `json:"thumbnail_url"`
	}
	status, err := p.get(ctx, target, &payload)
	if err != nil {
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			return domain.VideoMeta{}, domain.ErrEmbedBlocked
		}
		return domain.VideoMeta{}, err
	}
	if payload.Title == "" {
		return domain.VideoMeta{}, errors.New("empty oembed")
	}
	return domain.VideoMeta{
		VideoID:      videoID,
		Title:        payload.Title,
		Channel:      payload.AuthorName,
		ThumbnailURL: domain.ThumbnailURL(videoID),
	}, nil
}

func (p *Prober) oembedURL(videoID string) string {
	watch := "https://www.youtube.com/watch?v=" + videoID
	separator := "?"
	if strings.Contains(p.endpoint, "?") {
		separator = "&"
	}
	return p.endpoint + separator + "format=json&url=" + url.QueryEscape(watch)
}

func (p *Prober) get(ctx context.Context, target string, dst any) (int, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, target, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "FanaKaraoke/1.0 (+https://github.com/fana-karaoke)")
	res, err := p.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, errors.New("oembed status " + res.Status)
	}
	return res.StatusCode, json.NewDecoder(res.Body).Decode(dst)
}

var _ ports.EmbedProber = (*Prober)(nil)
