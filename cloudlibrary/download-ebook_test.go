package cloudlibrary

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const ebookFixture = `{"patronItems":[{"itemId":"book-id","loanId":"loan-id","title":"Sixth of the Dusk","mediaCategory":"eBook"}],"totalSegments":1,"RPC_DOMAIN_PUBLIC":"https://download.example/","reaktor":"token+&=#?"}`

func TestDownloadEBook(t *testing.T) {
	c := New()
	loanRequests := 0
	downloadRequests := 0
	c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "ebook.yourcloudlibrary.com" {
			if req.Method != http.MethodPost || req.URL.Query().Get("_data") != "routes/library.$name.mybooks.current" {
				t.Fatalf("unexpected loans request: %s %s", req.Method, req.URL)
			}
			loanRequests++
			return audiobookResponse(http.StatusOK, ebookFixture), nil
		}
		if req.Method != http.MethodGet || req.URL.Host != "download.example" || req.URL.Path != "/delivery/metadata" {
			t.Fatalf("unexpected download request: %s %s", req.Method, req.URL)
		}
		downloadRequests++
		q := req.URL.Query()
		if q.Get("token") != "token+&=#?" || q.Get("udid") != "loan-id" || q.Get("exporter") != "com.bookpac.exporter.fulfillmenttoken" || q.Get("tokenType") != "deviceVendorID" {
			t.Fatalf("incorrect download parameters: %v", q)
		}
		return audiobookResponse(http.StatusOK, "<fulfillmentToken/>"), nil
	})
	books, err := c.MyBooks()
	if err != nil || len(books) != 1 {
		t.Fatalf("MyBooks = %v, %v", books, err)
	}
	dir := filepath.Join(t.TempDir(), "ebooks")
	filename := filepath.Join(dir, "custom-filename.acsm")
	books[0].Title = ""
	if err := c.DownloadEBook(books[0], filename); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filename)
	if err != nil || string(data) != "<fulfillmentToken/>" {
		t.Fatalf("download = %q, %v", data, err)
	}
	if loanRequests != 1 || downloadRequests != 1 {
		t.Fatalf("requests: loans = %d, downloads = %d", loanRequests, downloadRequests)
	}
}

func ebookTestClient(t *testing.T) (*Client, PatronItem) {
	t.Helper()
	var fixture MyBooksRespose
	if err := json.Unmarshal([]byte(ebookFixture), &fixture); err != nil {
		t.Fatal(err)
	}
	c := New()
	c.rpcDomainPublic = fixture.RpcDomainPublic
	c.reaktor = fixture.Reaktor
	return c, fixture.PatronItems[0]
}

func TestDownloadEBookErrors(t *testing.T) {
	for _, tt := range []struct {
		name, wantErr string
		modify        func(*Client, *PatronItem)
		status        int
	}{
		{name: "missing loan", modify: func(c *Client, b *PatronItem) { b.LoanID = "" }, wantErr: "missing ebook download metadata"},
		{name: "missing token", modify: func(c *Client, b *PatronItem) { c.reaktor = "" }, wantErr: "missing ebook download metadata"},
		{name: "missing service URL", modify: func(c *Client, b *PatronItem) { c.rpcDomainPublic = "" }, wantErr: "missing ebook download metadata"},
		{name: "audiobook", modify: func(c *Client, b *PatronItem) { b.MediaCategory = "AudioBook" }, wantErr: "not an ebook"},
		{name: "download status", status: http.StatusForbidden, wantErr: "ebook download status"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c, book := ebookTestClient(t)
			if tt.modify != nil {
				tt.modify(c, &book)
			}
			c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
				if tt.status == 0 || req.URL.Host != "download.example" {
					t.Fatalf("unexpected request: %s", req.URL)
				}
				return audiobookResponse(tt.status, "denied"), nil
			})
			dir := t.TempDir()
			if err := c.DownloadEBook(book, filepath.Join(dir, "book.acsm")); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 0 {
				t.Fatalf("failed download left files: %v, %v", entries, err)
			}
		})
	}
}

func TestDownloadEBookPreservesExistingFileOnFailure(t *testing.T) {
	c, book := ebookTestClient(t)
	c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "download.example" {
			t.Fatalf("unexpected request: %s", req.URL)
		}
		resp := audiobookResponse(http.StatusOK, "")
		resp.Body = interruptedChapter{}
		return resp, nil
	})
	dir := t.TempDir()
	filename := filepath.Join(dir, "Sixth_of_the_Dusk.acsm")
	if err := os.WriteFile(filename, []byte("existing ACSM"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := c.DownloadEBook(book, filename); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("error = %v, want unexpected EOF", err)
	}
	data, err := os.ReadFile(filename)
	if err != nil || string(data) != "existing ACSM" {
		t.Fatalf("existing file = %q, %v", data, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary file was not removed: %v, %v", entries, err)
	}
}

func TestMyBooks(t *testing.T) {
	c := New()
	c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
		return audiobookResponse(http.StatusOK, ebookFixture), nil
	})
	books, err := c.MyBooks()
	var fixture MyBooksRespose
	if err := json.Unmarshal([]byte(ebookFixture), &fixture); err != nil {
		t.Fatal(err)
	}
	if err != nil || len(books) != 1 || books[0].ItemID != fixture.PatronItems[0].ItemID {
		t.Fatalf("MyBooks = %v, %v", books, err)
	}
}
