package invidious

import (
	"context"
	"errors"
	"sync"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// Catalog implements ports.VideoCatalog over an upstream client, an embed
// prober and a TTL cache.
type Catalog struct {
	upstream *Client
	prober   ports.EmbedProber
	cache    *Cache
}

// NewCatalog builds the catalog.
func NewCatalog(upstream *Client, prober ports.EmbedProber, cache *Cache) *Catalog {
	return &Catalog{upstream: upstream, prober: prober, cache: cache}
}

// Search returns karaoke hits, dropping videos that cannot be embedded.
func (c *Catalog) Search(ctx context.Context, raw string) ([]domain.SearchResult, bool, error) {
	query := normalizeQuery(raw)
	if query == "" {
		return []domain.SearchResult{}, false, nil
	}
	key := query + "|upstream"
	if cached, ok := c.cache.Query(key, true); ok {
		return cached, false, nil
	}
	results, err := c.upstream.Search(ctx, query)
	if err == nil {
		results = c.keepEmbeddable(ctx, results)
		c.cache.StoreQuery(key, results)
		return results, false, nil
	}
	if stale, ok := c.cache.Query(key, false); ok {
		return stale, true, nil
	}
	return nil, false, domain.ErrSearchFailed
}

// Lookup returns validated metadata, with oEmbed as the last resort.
func (c *Catalog) Lookup(ctx context.Context, videoID string) (domain.VideoMeta, error) {
	if !domain.ValidVideoID(videoID) {
		return domain.VideoMeta{}, domain.ErrInvalidURL
	}
	if meta, ok := c.cache.Video(videoID, true); ok {
		return meta, nil
	}
	if c.cache.Blocked(videoID) {
		return domain.VideoMeta{}, domain.ErrEmbedBlocked
	}
	if meta, ok := c.fromUpstream(ctx, videoID); ok {
		return meta, nil
	}
	if c.cache.Blocked(videoID) {
		return domain.VideoMeta{}, domain.ErrEmbedBlocked
	}
	if stale, ok := c.cache.Video(videoID, false); ok {
		return stale, nil
	}
	return c.fromOEmbed(videoID)
}

func (c *Catalog) fromUpstream(ctx context.Context, videoID string) (domain.VideoMeta, bool) {
	meta, err := c.upstream.Video(ctx, videoID)
	if err != nil {
		return domain.VideoMeta{}, false
	}
	if !c.embedAllowed(ctx, videoID) {
		return domain.VideoMeta{}, false
	}
	c.cache.StoreVideo(videoID, meta)
	return meta, true
}

func (c *Catalog) fromOEmbed(videoID string) (domain.VideoMeta, error) {
	embedCtx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	meta, err := c.prober.Probe(embedCtx, videoID)
	if err == nil {
		c.cache.StoreVideo(videoID, meta)
		return meta, nil
	}
	if errors.Is(err, domain.ErrEmbedBlocked) {
		c.cache.MarkBlocked(videoID)
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

func (c *Catalog) keepEmbeddable(ctx context.Context, results []domain.SearchResult) []domain.SearchResult {
	if len(results) == 0 {
		return results
	}
	keep := make([]bool, len(results))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i := range results {
		if _, ok := c.cache.Video(results[i].VideoID, false); ok {
			keep[i] = true
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			keep[i] = c.probeOne(ctx, results[i])
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

func (c *Catalog) probeOne(ctx context.Context, result domain.SearchResult) bool {
	meta, state := c.probeEmbed(ctx, result.VideoID)
	switch state {
	case embedNo:
		return false
	case embedYes:
		c.cache.StoreVideo(result.VideoID, mergeMeta(meta, result))
		return true
	}
	return true
}

func mergeMeta(meta domain.VideoMeta, result domain.SearchResult) domain.VideoMeta {
	if result.Title != "" {
		meta.Title = result.Title
	}
	if result.Channel != "" {
		meta.Channel = result.Channel
	}
	if result.ThumbnailURL != "" {
		meta.ThumbnailURL = result.ThumbnailURL
	}
	meta.VideoID = result.VideoID
	meta.DurationSeconds = result.DurationSeconds
	return meta
}

func (c *Catalog) probeEmbed(ctx context.Context, videoID string) (domain.VideoMeta, embedState) {
	if c.cache.Blocked(videoID) {
		return domain.VideoMeta{}, embedNo
	}
	probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	meta, err := c.prober.Probe(probeCtx, videoID)
	if err == nil {
		return meta, embedYes
	}
	if errors.Is(err, domain.ErrEmbedBlocked) {
		c.cache.MarkBlocked(videoID)
		return domain.VideoMeta{}, embedNo
	}
	return domain.VideoMeta{}, embedUnknown
}

func (c *Catalog) embedAllowed(ctx context.Context, videoID string) bool {
	_, state := c.probeEmbed(ctx, videoID)
	return state != embedNo
}

var _ ports.VideoCatalog = (*Catalog)(nil)
