package cloudlibrary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const audioOrigin = "https://audio.yourcloudlibrary.com/"

var audiobookTitlePattern = regexp.MustCompile(`[^a-zA-Z0-9]+`)

type audiobookListenData struct {
	FulfillmentID string `json:"fulfillmentId"`
	SessionKey    string `json:"session_key"`
	LicenseID     string `json:"licenseId"`
	Title         string `json:"title"`
}

// DownloadAudiobook downloads the chapters for a listen-page ID (for example,
// "pt986g9") into dir, using the client's authenticated CloudLibrary session.
// Chapters are saved in playlist order as Title_001.mp3, Title_002.mp3, etc.
func (c *Client) DownloadAudiobook(id, dir string) error {
	if id == "" {
		return fmt.Errorf("cloudlibrary: audiobook ID is required")
	}
	book, err := c.loadAudiobook(id)
	if err != nil {
		return err
	}
	if book.FulfillmentID == "" || book.SessionKey == "" || book.LicenseID == "" || book.Title == "" {
		return fmt.Errorf("cloudlibrary: listen page is missing audiobook metadata")
	}

	body, err := json.Marshal(struct {
		LicenseID string `json:"license_id"`
	}{LicenseID: book.LicenseID})
	if err != nil {
		return fmt.Errorf("cloudlibrary: encode playlist request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.findawayworld.com/v4/audiobooks/"+url.PathEscape(book.FulfillmentID)+"/playlists", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cloudlibrary: create playlist request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("session-key", book.SessionKey)
	req.Header.Set("Referer", audioOrigin)

	// Match fetch's credentials: "omit" for the playlist and chapter hosts.
	downloadClient := *c.http
	downloadClient.Jar = nil
	var playlist struct {
		Playlist []struct {
			URL string `json:"url"`
		} `json:"playlist"`
	}
	if err := loadAudiobookPlaylist(&downloadClient, req, &playlist); err != nil {
		return err
	}
	if len(playlist.Playlist) == 0 {
		return fmt.Errorf("cloudlibrary: audiobook playlist is empty")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cloudlibrary: create audiobook directory: %w", err)
	}

	// The normal API timeout is too short for large chapters on slow connections.
	downloadClient.Timeout = 30 * time.Minute
	title := audiobookTitlePattern.ReplaceAllString(book.Title, "_")
	for i, track := range playlist.Playlist {
		filename := fmt.Sprintf("%s_%03d.mp3", title, i+1)
		if err := downloadAudiobookChapter(&downloadClient, track.URL, filepath.Join(dir, filename)); err != nil {
			return fmt.Errorf("cloudlibrary: download chapter %d (%s): %w", i+1, filename, err)
		}
	}
	return nil
}

func (c *Client) loadAudiobook(id string) (*audiobookListenData, error) {
	q := url.Values{"_data": {"routes/listen.$id"}}
	var book audiobookListenData
	if err := c.request(http.MethodGet, audioOrigin+"listen/"+url.PathEscape(id)+"?"+q.Encode(), nil, 1, &book); err != nil {
		return nil, fmt.Errorf("cloudlibrary: load audiobook: %w", err)
	}
	return &book, nil
}

func loadAudiobookPlaylist(client *http.Client, req *http.Request, playlist any) error {
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cloudlibrary: request audiobook playlist: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cloudlibrary: audiobook playlist status: %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(playlist); err != nil {
		return fmt.Errorf("cloudlibrary: decode audiobook playlist: %w", err)
	}
	return nil
}

func downloadAudiobookChapter(client *http.Client, trackURL, filename string) error {
	req, err := http.NewRequest(http.MethodGet, trackURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", audioOrigin)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Publish only complete chapters, preserving an existing file on failure.
	file, err := os.CreateTemp(filepath.Dir(filename), ".chapter-*.mp3")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(file.Name(), filename); err != nil {
		return err
	}
	return nil
}
