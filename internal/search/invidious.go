package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"fana-karaoke/internal/domain"
)

type Client struct {
	instances []string
	region    string
	http      *http.Client
	queryTTL  time.Duration
	staleTTL  time.Duration
	videoTTL  time.Duration
	oembedURL string
	mu        sync.Mutex
	queries   map[string]cachedQuery
	videos    map[string]cachedVideo
	blocked   map[string]time.Time
	unhealthy map[string]time.Time
	now       func() time.Time
}

type cachedQuery struct {
	results []domain.SearchResult
	at      time.Time
}

type cachedVideo struct {
	meta domain.VideoMeta
	at   time.Time
}

type Options struct {
	Instances []string
	Region    string
	QueryTTL  time.Duration
	StaleTTL  time.Duration
	VideoTTL  time.Duration
}

func New(opts Options) *Client {
	if opts.Region == "" {
		opts.Region = "FR"
	}
	if opts.QueryTTL <= 0 {
		opts.QueryTTL = 24 * time.Hour
	}
	if opts.StaleTTL <= 0 {
		opts.StaleTTL = 7 * 24 * time.Hour
	}
	if opts.VideoTTL <= 0 {
		opts.VideoTTL = 7 * 24 * time.Hour
	}
	instances := make([]string, 0, len(opts.Instances))
	for _, in := range opts.Instances {
		in = strings.TrimRight(strings.TrimSpace(in), "/")
		if in != "" {
			instances = append(instances, in)
		}
	}
	return &Client{
		instances: instances,
		region:    opts.Region,
		http:      &http.Client{Timeout: 5 * time.Second},
		queryTTL:  opts.QueryTTL,
		staleTTL:  opts.StaleTTL,
		videoTTL:  opts.VideoTTL,
		oembedURL: "https://www.youtube.com/oembed",
		queries:   make(map[string]cachedQuery),
		videos:    make(map[string]cachedVideo),
		blocked:   make(map[string]time.Time),
		unhealthy: make(map[string]time.Time),
		now:       time.Now,
	}
}

func (c *Client) Search(ctx context.Context, raw string) ([]domain.SearchResult, bool, error) {
	q := normalizeQuery(raw)
	if q == "" {
		return []domain.SearchResult{}, false, nil
	}
	key := strings.ToLower(q) + "|" + c.region
	if results, ok := c.freshQuery(key); ok {
		return results, false, nil
	}

	results, err := c.fetchSearch(ctx, q)
	if err == nil {
		results = c.keepEmbeddable(ctx, results)
		c.storeQuery(key, results)
		return results, false, nil
	}
	if stale, ok := c.staleQuery(key); ok {
		return stale, true, nil
	}
	return nil, false, domain.ErrSearchFailed
}

func (c *Client) Lookup(ctx context.Context, videoID string) (domain.VideoMeta, error) {
	if !validID(videoID) {
		return domain.VideoMeta{}, domain.ErrInvalidURL
	}
	if meta, ok := c.freshVideo(videoID); ok {
		return meta, nil
	}
	if c.isBlocked(videoID) {
		return domain.VideoMeta{}, domain.ErrEmbedBlocked
	}

	meta, err := c.fetchVideo(ctx, videoID)
	if err == nil {
		// Invidious can describe videos whose owners disabled embedding. Probe
		// oEmbed before queueing so the guest gets a clear reason instead of a
		// player error later. Network failures fail open.
		if !c.embedAllowed(ctx, videoID) {
			return domain.VideoMeta{}, domain.ErrEmbedBlocked
		}
		c.storeVideo(videoID, meta)
		return meta, nil
	}
	if stale, ok := c.staleVideo(videoID); ok {
		return stale, nil
	}
	// Metadata often has to fall back to oEmbed after slow public instances.
	// Give it its own small budget instead of inheriting an almost-expired
	// request context.
	embedCtx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	embed, oerr := c.oEmbed(embedCtx, videoID)
	cancel()
	if oerr == nil {
		c.storeVideo(videoID, embed)
		return embed, nil
	}
	if errors.Is(oerr, domain.ErrEmbedBlocked) {
		c.markBlocked(videoID)
		return domain.VideoMeta{}, domain.ErrEmbedBlocked
	}
	return domain.VideoMeta{}, domain.ErrVideoMissing
}

type embedState int

const (
	embedUnknown embedState = iota
	embedYes
	embedNo
)

// keepEmbeddable drops search results whose owners disabled embedding. A
// result whose probe fails for network reasons is kept: metadata lookup still
// has a chance to succeed when the guest adds it.
func (c *Client) keepEmbeddable(ctx context.Context, results []domain.SearchResult) []domain.SearchResult {
	if len(results) == 0 {
		return results
	}
	keep := make([]bool, len(results))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i := range results {
		if _, ok := c.freshVideo(results[i].VideoID); ok {
			keep[i] = true
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			meta, state := c.probeEmbed(ctx, results[i].VideoID)
			switch state {
			case embedNo:
				return
			case embedYes:
				if results[i].Title != "" {
					meta.Title = results[i].Title
				}
				if results[i].Channel != "" {
					meta.Channel = results[i].Channel
				}
				if results[i].ThumbnailURL != "" {
					meta.ThumbnailURL = results[i].ThumbnailURL
				}
				meta.VideoID = results[i].VideoID
				meta.DurationSeconds = results[i].DurationSeconds
				c.storeVideo(results[i].VideoID, meta)
				keep[i] = true
			default:
				keep[i] = true
			}
		}(i)
	}
	wg.Wait()
	out := make([]domain.SearchResult, 0, len(results))
	for i, result := range results {
		if keep[i] {
			out = append(out, result)
		}
	}
	return out
}

func (c *Client) probeEmbed(ctx context.Context, videoID string) (domain.VideoMeta, embedState) {
	if c.isBlocked(videoID) {
		return domain.VideoMeta{}, embedNo
	}
	probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	meta, err := c.oEmbed(probeCtx, videoID)
	if err == nil {
		return meta, embedYes
	}
	if errors.Is(err, domain.ErrEmbedBlocked) {
		c.markBlocked(videoID)
		return domain.VideoMeta{}, embedNo
	}
	return domain.VideoMeta{}, embedUnknown
}

func (c *Client) embedAllowed(ctx context.Context, videoID string) bool {
	_, state := c.probeEmbed(ctx, videoID)
	return state != embedNo
}

func (c *Client) markBlocked(videoID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.blocked) >= 2000 {
		for id := range c.blocked {
			delete(c.blocked, id)
			break
		}
	}
	c.blocked[videoID] = c.now().Add(c.videoTTL)
}

func (c *Client) isBlocked(videoID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	until, ok := c.blocked[videoID]
	if !ok {
		return false
	}
	if !c.now().Before(until) {
		delete(c.blocked, videoID)
		return false
	}
	return true
}

func normalizeQuery(raw string) string {
	fields := strings.Fields(raw)
	cleaned := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.Map(func(r rune) rune {
			if r < 32 || r == 127 {
				return -1
			}
			return r
		}, f)
		if f != "" {
			cleaned = append(cleaned, f)
		}
	}
	q := strings.Join(cleaned, " ")
	runes := []rune(q)
	if len(runes) > 120 {
		q = string(runes[:120])
	}
	if q == "" {
		return ""
	}
	if !strings.Contains(strings.ToLower(q), "karaoke") {
		q += " karaoke"
	}
	return q
}

func (c *Client) fetchSearch(ctx context.Context, query string) ([]domain.SearchResult, error) {
	for _, base := range c.pickInstances(4) {
		u := fmt.Sprintf("%s/api/v1/search?q=%s&type=video&region=%s&page=1",
			base, url.QueryEscape(query), url.QueryEscape(c.region))
		var payload []struct {
			Type            string `json:"type"`
			Title           string `json:"title"`
			VideoID         string `json:"videoId"`
			Author          string `json:"author"`
			LengthSeconds   int    `json:"lengthSeconds"`
			LiveNow         bool   `json:"liveNow"`
			VideoThumbnails []struct {
				URL string `json:"url"`
			} `json:"videoThumbnails"`
		}
		if err := c.getJSON(ctx, u, &payload); err != nil {
			c.markUnhealthy(base)
			continue
		}
		out := make([]domain.SearchResult, 0, 20)
		seen := map[string]bool{}
		for _, r := range payload {
			if r.Type != "video" || r.LiveNow || r.LengthSeconds <= 60 || r.LengthSeconds > 600 {
				continue
			}
			if !validID(r.VideoID) || seen[r.VideoID] {
				continue
			}
			seen[r.VideoID] = true
			out = append(out, domain.SearchResult{
				VideoID:         r.VideoID,
				Title:           html.UnescapeString(r.Title),
				Channel:         html.UnescapeString(r.Author),
				ThumbnailURL:    thumb(r.VideoID),
				DurationSeconds: r.LengthSeconds,
			})
			if len(out) >= 20 {
				break
			}
		}
		return out, nil
	}
	return nil, errors.New("all invidious instances failed")
}

func (c *Client) fetchVideo(ctx context.Context, videoID string) (domain.VideoMeta, error) {
	for _, base := range c.pickInstances(2) {
		u := fmt.Sprintf("%s/api/v1/videos/%s", base, url.PathEscape(videoID))
		var payload struct {
			Type          string `json:"type"`
			Title         string `json:"title"`
			Author        string `json:"author"`
			LengthSeconds int    `json:"lengthSeconds"`
			LiveNow       bool   `json:"liveNow"`
		}
		if err := c.getJSON(ctx, u, &payload); err != nil {
			c.markUnhealthy(base)
			continue
		}
		if payload.Title == "" {
			continue
		}
		return domain.VideoMeta{
			VideoID:         videoID,
			Title:           html.UnescapeString(payload.Title),
			Channel:         html.UnescapeString(payload.Author),
			ThumbnailURL:    thumb(videoID),
			DurationSeconds: payload.LengthSeconds,
			Live:            payload.LiveNow,
		}, nil
	}
	return domain.VideoMeta{}, errors.New("metadata lookup failed")
}

func (c *Client) oEmbed(ctx context.Context, videoID string) (domain.VideoMeta, error) {
	watch := "https://www.youtube.com/watch?v=" + videoID
	separator := "?"
	if strings.Contains(c.oembedURL, "?") {
		separator = "&"
	}
	u := c.oembedURL + separator + "format=json&url=" + url.QueryEscape(watch)
	var payload struct {
		Title        string `json:"title"`
		AuthorName   string `json:"author_name"`
		ThumbnailURL string `json:"thumbnail_url"`
	}
	status, err := c.getJSONStatus(ctx, u, &payload)
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
		ThumbnailURL: thumb(videoID),
	}, nil
}

func (c *Client) getJSON(ctx context.Context, rawURL string, dst any) error {
	_, err := c.getJSONStatus(ctx, rawURL, dst)
	return err
}

func (c *Client) getJSONStatus(ctx context.Context, rawURL string, dst any) (int, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "FanaKaraoke/1.0 (+https://github.com/fana-karaoke)")
	req.Header.Set("Accept", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, fmt.Errorf("upstream status %d", res.StatusCode)
	}
	return res.StatusCode, json.NewDecoder(res.Body).Decode(dst)
}

func (c *Client) pickInstances(max int) []string {
	if len(c.instances) == 0 {
		return nil
	}
	c.mu.Lock()
	now := c.now()
	healthy := make([]string, 0, len(c.instances))
	unhealthy := make([]string, 0, len(c.instances))
	for _, in := range c.instances {
		until, bad := c.unhealthy[in]
		if bad && now.Before(until) {
			unhealthy = append(unhealthy, in)
		} else {
			healthy = append(healthy, in)
		}
	}
	c.mu.Unlock()
	// Preserve configured priority: the first instance is the most reliable
	// one. Unhealthy instances move behind the healthy ones until their
	// circuit-breaker window expires.
	all := append(healthy, unhealthy...)
	if len(all) > max {
		all = all[:max]
	}
	return all
}

func (c *Client) markUnhealthy(instance string) {
	c.mu.Lock()
	c.unhealthy[instance] = c.now().Add(60 * time.Second)
	c.mu.Unlock()
}

func (c *Client) freshQuery(key string) ([]domain.SearchResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.queries[key]
	if !ok || c.now().Sub(entry.at) > c.queryTTL {
		return nil, false
	}
	return cloneResults(entry.results), true
}

func (c *Client) staleQuery(key string) ([]domain.SearchResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.queries[key]
	if !ok || c.now().Sub(entry.at) > c.staleTTL {
		return nil, false
	}
	return cloneResults(entry.results), true
}

func (c *Client) storeQuery(key string, results []domain.SearchResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.queries) >= 1000 {
		for k := range c.queries {
			delete(c.queries, k)
			break
		}
	}
	c.queries[key] = cachedQuery{results: cloneResults(results), at: c.now()}
}

func (c *Client) freshVideo(id string) (domain.VideoMeta, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.videos[id]
	if !ok || c.now().Sub(entry.at) > c.videoTTL {
		return domain.VideoMeta{}, false
	}
	return entry.meta, true
}

func (c *Client) staleVideo(id string) (domain.VideoMeta, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.videos[id]
	if !ok || c.now().Sub(entry.at) > c.staleTTL {
		return domain.VideoMeta{}, false
	}
	return entry.meta, true
}

func (c *Client) storeVideo(id string, meta domain.VideoMeta) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.videos) >= 2000 {
		for k := range c.videos {
			delete(c.videos, k)
			break
		}
	}
	c.videos[id] = cachedVideo{meta: meta, at: c.now()}
}

func cloneResults(in []domain.SearchResult) []domain.SearchResult {
	out := make([]domain.SearchResult, len(in))
	copy(out, in)
	return out
}

func thumb(id string) string {
	return "https://i.ytimg.com/vi/" + id + "/mqdefault.jpg"
}

func validID(id string) bool {
	if len(id) != 11 {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
