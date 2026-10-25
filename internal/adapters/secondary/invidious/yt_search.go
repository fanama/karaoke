package invidious

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"fana-karaoke/internal/domain"
)

// ytConsentCookie skips YouTube's cookie-consent wall for EU servers.
const ytConsentCookie = "SOCS=CAISHAgBEhJnd3NfMjAyMzA4MTUtMF9SQzIaAmVuIAEaBgiAo_CmBg; " +
	"CONSENT=YES+cb.20210328-17-p0.en+FX+417"

const ytUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

// ytMaxBody caps how much of the search page is read into memory.
const ytMaxBody = 8 << 20

// searchYouTube scrapes the public search page instead of calling an API.
func (c *Client) searchYouTube(ctx context.Context, query string) ([]domain.SearchResult, error) {
	page, err := c.fetchYouTube(ctx, query)
	if err != nil {
		return nil, err
	}
	return parseInitialData(page)
}

func (c *Client) fetchYouTube(ctx context.Context, query string) (string, error) {
	target := fmt.Sprintf("%s?search_query=%s&hl=en&gl=%s",
		c.youtubeURL, url.QueryEscape(query), url.QueryEscape(c.region))
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ytUserAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cookie", ytConsentCookie)
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("youtube status %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, ytMaxBody))
	if err != nil {
		return "", err
	}
	return string(body), nil
}
