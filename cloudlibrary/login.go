package cloudlibrary

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Login submits the library card barcode and PIN for the given library URL name
// (for example, "guelphpl"). Session cookies are retained for subsequent requests.
func (c *Client) Login(library, barcode, pin string) error {
	const origin = "https://ebook.yourcloudlibrary.com"
	libraryURL := origin + "/library/" + url.PathEscape(library) + "/featured"
	// The login action requires the anonymous session established by the
	// library page. Posting credentials without that cookie returns HTTP 500.
	page, err := c.http.Get(libraryURL)
	if err != nil {
		return fmt.Errorf("cloudlibrary.Login: initialize session: %w", err)
	}
	_, readErr := io.Copy(io.Discard, page.Body)
	page.Body.Close()
	if page.StatusCode < 200 || page.StatusCode >= 300 {
		return fmt.Errorf("cloudlibrary.Login: initialize session: invalid status: %s", page.Status)
	}
	if readErr != nil {
		return fmt.Errorf("cloudlibrary.Login: initialize session: read response: %w", readErr)
	}

	body := url.Values{
		"action":  {"login"},
		"barcode": {barcode},
		"library": {library},
		"pin":     {pin},
	}

	req, err := http.NewRequest(http.MethodPost, origin+"/?_data=root", strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("cloudlibrary.Login: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", origin)
	req.Header.Set("Referer", libraryURL)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("cloudlibrary.Login: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cloudlibrary.Login: invalid status: %s", resp.Status)
	}
	// Login can return an empty Remix redirect response rather than JSON. The
	// cookie jar has already saved any Set-Cookie headers; do not repeat the POST.
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("cloudlibrary.Login: read response: %w", err)
	}
	return nil
}

/*
curl 'https://ebook.yourcloudlibrary.com/?_data=root' \
  --compressed \
  -X POST \
  -H 'User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0' \
  -H 'Accept: * /*' \
  -H 'Accept-Language: en-US,en;q=0.5' \
  -H 'Accept-Encoding: gzip, deflate, br, zstd' \
  -H 'Referer: https://ebook.yourcloudlibrary.com/library/guelphpl/featured' \
  -H 'Content-Type: application/x-www-form-urlencoded;charset=UTF-8' \
  -H 'Origin: https://ebook.yourcloudlibrary.com' \
  -H 'Connection: keep-alive' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'Priority: u=4' \
  -H 'TE: trailers' \
  --data-raw 'action=login&barcode=<username>&library=guelphpl&pin=<password>'

*/
