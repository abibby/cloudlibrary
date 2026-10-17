package cloudlibrary

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestLogin(t *testing.T) {
	c := New()
	calls := 0
	c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls <= 2 {
			if req.Method != http.MethodGet || req.URL.String() != "https://ebook.yourcloudlibrary.com/library/guelphpl/featured" {
				t.Fatalf("unexpected session request: %s %s", req.Method, req.URL)
			}
			if calls == 1 {
				resp := audiobookResponse(http.StatusFound, "")
				resp.Header.Set("Location", req.URL.String())
				resp.Header.Set("Set-Cookie", "__session_PROD=anonymous-session; Path=/; Secure; HttpOnly")
				return resp, nil
			}
			if got := req.Header.Get("Cookie"); got != "__session_PROD=anonymous-session" {
				t.Errorf("session redirect Cookie = %q", got)
			}
			return audiobookResponse(http.StatusOK, "<html></html>"), nil
		}
		if calls == 4 {
			if got := req.Header.Get("Cookie"); got != "__session_PROD=test-session" {
				t.Errorf("subsequent request Cookie = %q", got)
			}
			return audiobookResponse(http.StatusOK, `{}`), nil
		}
		if req.Method != http.MethodPost || req.URL.String() != "https://ebook.yourcloudlibrary.com/?_data=root" {
			t.Fatalf("unexpected login request: %s %s", req.Method, req.URL)
		}
		for name, want := range map[string]string{
			"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
			"Accept":       "*/*",
			"Origin":       "https://ebook.yourcloudlibrary.com",
			"Referer":      "https://ebook.yourcloudlibrary.com/library/guelphpl/featured",
			"Cookie":       "__session_PROD=anonymous-session",
		} {
			if got := req.Header.Get(name); got != want {
				t.Errorf("%s = %q, want %q", name, got, want)
			}
		}
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		want := url.Values{
			"action": {"login"}, "library": {"guelphpl"},
			"barcode": {"card +&=123"}, "pin": {"p&+= é"},
		}
		if !reflect.DeepEqual(req.PostForm, want) {
			t.Errorf("form = %v, want %v", req.PostForm, want)
		}
		resp := audiobookResponse(http.StatusNoContent, "")
		resp.Header.Set("X-Remix-Status", "302")
		resp.Header.Set("Set-Cookie", "__session_PROD=test-session; Path=/; Secure; HttpOnly")
		return resp, nil
	})
	if err := c.Login("guelphpl", "card +&=123", "p&+= é"); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("login sent %d requests, want 3 (including session redirect)", calls)
	}
	var result any
	if err := c.request(http.MethodGet, "https://ebook.yourcloudlibrary.com/library/guelphpl/mybooks/current", nil, 1, &result); err != nil {
		t.Fatal(err)
	}
}

func TestLoginErrors(t *testing.T) {
	for _, stage := range []string{"session", "login"} {
		t.Run(stage, func(t *testing.T) {
			for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError} {
				t.Run(http.StatusText(status), func(t *testing.T) {
					c := New()
					calls := 0
					c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
						calls++
						if stage == "login" && req.Method == http.MethodGet {
							return audiobookResponse(http.StatusOK, ""), nil
						}
						return audiobookResponse(status, ""), nil
					})
					if err := c.Login("guelphpl", "card", "pin"); err == nil {
						t.Fatal("expected HTTP error")
					} else if stage == "session" && !strings.Contains(err.Error(), "initialize session") {
						t.Fatalf("expected session error, got %v", err)
					}
					if stage == "session" && calls != 1 {
						t.Fatalf("sent %d requests after session failure, want 1", calls)
					}
				})
			}
			t.Run("transport", func(t *testing.T) {
				c := New()
				want := errors.New("connection failed")
				c.http.Transport = audiobookTransport(func(req *http.Request) (*http.Response, error) {
					if stage == "login" && req.Method == http.MethodGet {
						return audiobookResponse(http.StatusOK, ""), nil
					}
					return nil, want
				})
				if err := c.Login("guelphpl", "card", "pin"); !errors.Is(err, want) {
					t.Fatalf("error = %v, want %v", err, want)
				}
			})
		})
	}
}
