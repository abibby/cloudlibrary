package cloudlibrary

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type audiobookTransport func(*http.Request) (*http.Response, error)

func (f audiobookTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func audiobookResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

const listenFixture = `{"fulfillmentId":"123","session_key":"test-session","licenseId":"test-license","title":"Brigands & Breadknives"}`
const playlistFixture = `{"playlist":[{"url":"https://chapters.example/one"},{"url":"https://chapters.example/two"}]}`

func TestDownloadAudiobook(t *testing.T) {
	c := New()
	originalTimeout := c.http.Timeout
	for _, host := range []string{"api.findawayworld.com", "chapters.example"} {
		u, err := url.Parse("https://" + host)
		if err != nil {
			t.Fatal(err)
		}
		c.http.Jar.SetCookies(u, []*http.Cookie{{Name: "private", Value: "cookie"}})
	}
	var requests []string
	c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.Path)
		if req.URL.Host == "audio.yourcloudlibrary.com" {
			if req.Method != http.MethodGet || req.URL.Path != "/listen/book-id" || req.URL.Query().Get("_data") != "routes/listen.$id" {
				t.Fatalf("unexpected listen request: %s %s", req.Method, req.URL)
			}
			return audiobookResponse(http.StatusOK, listenFixture), nil
		}
		if req.Header.Get("Cookie") != "" {
			t.Error("cookies sent to download host")
		}
		if req.Referer() != audioOrigin {
			t.Errorf("Referer = %q", req.Referer())
		}
		if req.URL.Host == "api.findawayworld.com" {
			if req.Method != http.MethodPost || req.URL.Path != "/v4/audiobooks/123/playlists" {
				t.Fatalf("unexpected playlist request: %s %s", req.Method, req.URL)
			}
			if req.Header.Get("session-key") != "test-session" || req.Header.Get("Content-Type") != "application/json" {
				t.Error("missing playlist headers")
			}
			var body map[string]string
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["license_id"] != "test-license" {
				t.Errorf("playlist body = %v", body)
			}
			return audiobookResponse(http.StatusOK, playlistFixture), nil
		}
		if req.URL.Host != "chapters.example" || req.Method != http.MethodGet || req.Header.Get("session-key") != "" {
			t.Fatalf("unexpected chapter request: %s %s", req.Method, req.URL)
		}
		return audiobookResponse(http.StatusOK, "audio:"+req.URL.Path), nil
	})
	dir := filepath.Join(t.TempDir(), "book")
	if err := c.DownloadAudiobook("book-id", dir); err != nil {
		t.Fatal(err)
	}
	for i, chapter := range []string{"one", "two"} {
		data, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("Brigands_Breadknives_%03d.mp3", i+1)))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "audio:/"+chapter {
			t.Errorf("chapter %d = %q", i+1, data)
		}
	}
	if got := strings.Join(requests, ","); got != "/listen/book-id,/v4/audiobooks/123/playlists,/one,/two" {
		t.Errorf("request order = %s", got)
	}
	if c.http.Jar == nil || c.http.Timeout != originalTimeout {
		t.Error("download changed the original HTTP client")
	}
}

func TestDownloadAudiobookErrors(t *testing.T) {
	for _, tt := range []struct {
		name     string
		listen   string
		playlist string
		failAt   int
		wantErr  string
	}{
		{name: "listen status", listen: listenFixture, failAt: 1, wantErr: "load audiobook"},
		{name: "missing metadata", listen: `{}`, wantErr: "missing audiobook metadata"},
		{name: "playlist status", listen: listenFixture, failAt: 2, wantErr: "playlist status"},
		{name: "invalid playlist", listen: listenFixture, playlist: `{`, wantErr: "decode audiobook playlist"},
		{name: "empty playlist", listen: listenFixture, playlist: `{"playlist":[]}`, wantErr: "playlist is empty"},
		{name: "chapter status", listen: listenFixture, playlist: playlistFixture, failAt: 3, wantErr: "download chapter 1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			calls := 0
			c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
				calls++
				if calls == tt.failAt {
					return audiobookResponse(http.StatusForbidden, "denied"), nil
				}
				if calls == 1 {
					return audiobookResponse(http.StatusOK, tt.listen), nil
				}
				return audiobookResponse(http.StatusOK, tt.playlist), nil
			})
			dir := t.TempDir()
			err := c.DownloadAudiobook("book-id", dir)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 0 {
				t.Fatalf("failed download left files: %v, %v", entries, err)
			}
		})
	}
}

type interruptedChapter struct{}

func (interruptedChapter) Read(p []byte) (int, error) {
	return copy(p, "partial audio"), io.ErrUnexpectedEOF
}

func (interruptedChapter) Close() error { return nil }

func TestDownloadAudiobookPreservesExistingChapterOnFailure(t *testing.T) {
	c := New()
	c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Host {
		case "audio.yourcloudlibrary.com":
			return audiobookResponse(http.StatusOK, listenFixture), nil
		case "api.findawayworld.com":
			return audiobookResponse(http.StatusOK, playlistFixture), nil
		default:
			resp := audiobookResponse(http.StatusOK, "")
			resp.Body = interruptedChapter{}
			return resp, nil
		}
	})
	dir := t.TempDir()
	filename := filepath.Join(dir, "Brigands_Breadknives_001.mp3")
	if err := os.WriteFile(filename, []byte("existing audio"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := c.DownloadAudiobook("book-id", dir); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("error = %v, want unexpected EOF", err)
	}
	data, err := os.ReadFile(filename)
	if err != nil || string(data) != "existing audio" {
		t.Fatalf("existing chapter = %q, %v", data, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary file was not removed: %v, %v", entries, err)
	}
}
