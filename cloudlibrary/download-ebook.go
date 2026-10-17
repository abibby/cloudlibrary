package cloudlibrary

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// DownloadEBook downloads a borrowed ebook's ACSM fulfillment file to filename.
// Call MyBooks on this client first to load the download credentials.
func (c *Client) DownloadEBook(book PatronItem, filename string) error {
	if book.MediaCategory != "eBook" {
		return fmt.Errorf("cloudlibrary: item %q is not an ebook", book.ItemID)
	}
	if book.LoanID == "" || c.reaktor == "" || c.rpcDomainPublic == "" {
		return fmt.Errorf("cloudlibrary: missing ebook download metadata")
	}
	q := url.Values{
		"token":     {c.reaktor},
		"exporter":  {"com.bookpac.exporter.fulfillmenttoken"},
		"udid":      {book.LoanID},
		"tokenType": {"deviceVendorID"},
	}
	downloadURL := strings.TrimRight(c.rpcDomainPublic, "/") + "/delivery/metadata?" + q.Encode()
	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("cloudlibrary: create ebook download request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("cloudlibrary: download ebook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cloudlibrary: ebook download status: %s", resp.Status)
	}
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cloudlibrary: create ebook directory: %w", err)
	}
	// Publish only a complete download, preserving an existing file on failure.
	file, err := os.CreateTemp(dir, ".ebook-*.acsm")
	if err != nil {
		return fmt.Errorf("cloudlibrary: create ebook file: %w", err)
	}
	defer os.Remove(file.Name())
	_, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()
	if copyErr != nil {
		return fmt.Errorf("cloudlibrary: save ebook: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("cloudlibrary: close ebook file: %w", closeErr)
	}
	if err := os.Rename(file.Name(), filename); err != nil {
		return fmt.Errorf("cloudlibrary: publish ebook file: %w", err)
	}
	return nil
}
