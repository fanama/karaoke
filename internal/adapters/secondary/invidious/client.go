package invidious

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"fana-karaoke/internal/domain"
)

// Options configures the upstream client.
type Options struct {
	Instances []string
	Region    string
}

// Client queries public Invidious instances for search and metadata.
type Client struct {
	instances []string
	region    string
	http      *http.Client
	mu        sync.Mutex
	unhealthy map[string]time.Time
	now       func() time.Time
}

// NewClient builds the client and normalises the instance list.
func NewClient(opts Options) *Client {
	if opts.Region == "" {
		opts.Region = "FR"
	}
	instances := make([]string, 0, len(opts.Instances))
	for _, in := range opts.Instances {
		if in = strings.TrimRight(strings.TrimSpace(in), "/"); in != "" {
			instances = append(instances, in)
		}
	}
	return &Client{
		instances: instances,
		region:    opts.Region,
		http:      &http.Client{Timeout: 5 * time.Second},
		unhealthy: map[string]time.Time{},
		now:       time.Now,
	}
}

// SetClock lets tests control the circuit-breaker window.
func (c *Client) SetClock(now func() time.Time) { c.now = now }

// Search returns at most twenty karaoke hits from the first healthy instance.
func (c *Client) Search(ctx context.Context, query string) ([]domain.SearchResult, error) {
	for _, base := range c.pickInstances(4) {
		target := fmt.Sprintf("%s/api/v1/search?q=%s&type=video&region=%s&page=1",
			base, url.QueryEscape(query), url.QueryEscape(c.region))
		var payload []searchItem
		if err := c.getJSON(ctx, target, &payload); err != nil {
			c.markUnhealthy(base)
			continue
		}
		return parseSearchItems(payload), nil
	}
	return nil, errors.New("all invidious instances failed")
}

// Video returns metadata for one video id.
func (c *Client) Video(ctx context.Context, videoID string) (domain.VideoMeta, error) {
	for _, base := range c.pickInstances(2) {
		target := fmt.Sprintf("%s/api/v1/videos/%s", base, url.PathEscape(videoID))
		var payload videoItem
		if err := c.getJSON(ctx, target, &payload); err != nil {
			c.markUnhealthy(base)
			continue
		}
		if meta, ok := parseVideoItem(videoID, payload); ok {
			return meta, nil
		}
	}
	return domain.VideoMeta{}, errors.New("metadata lookup failed")
}

func (c *Client) getJSON(ctx context.Context, rawURL string, dst any) error {
	reqCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "FanaKaraoke/1.0 (+https://github.com/fana-karaoke)")
	req.Header.Set("Accept", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("upstream status %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(dst)
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
		if until, bad := c.unhealthy[in]; bad && now.Before(until) {
			unhealthy = append(unhealthy, in)
			continue
		}
		healthy = append(healthy, in)
	}
	c.mu.Unlock()
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
