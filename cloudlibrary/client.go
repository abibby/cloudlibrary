package cloudlibrary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type Client struct {
	http            *http.Client
	rpcDomainPublic string
	reaktor         string
}

func New() *Client {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		panic(err)
	}

	return &Client{
		http: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}
}

type SearchResponse struct {
	Action   json.RawMessage `json:"action"`   //: null
	Advanced bool            `json:"advanced"` //: true
	// Categories map[string][]Category `json:"categories"` //: {ADULT_FICTION: [,…],…}
	Results Results `json:"results"` //: {search: {totalItems: 1, totalSegments: 1,…}}
	Segment int     `json:"segment"` //: int
}

type Results struct {
	Search SearchResults `json:"search"`
}

func (c *Client) request(method, url string, body url.Values, maxRedirects int, v any) error {
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body.Encode()))
	if err != nil {
		return fmt.Errorf("cloudlibrary.Client: create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("cloudlibrary.Client: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cloudlibrary.Client: invalid status: %s", resp.Status)
	}

	if resp.Header.Get("X-Remix-Status") == "302" {
		return c.request(method, url, body, maxRedirects-1, v)
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("cloudlibrary.Client: json decode: %w", err)
	}

	return nil
}
