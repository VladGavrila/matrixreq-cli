package client

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFormURLEncoded(t *testing.T) {
	bigValue := strings.Repeat("x", 16*1024)

	tests := []struct {
		name   string
		method string
		call   func(c *Client) ([]byte, error)
	}{
		{
			name:   "PostFormURLEncoded sends form body",
			method: http.MethodPost,
			call: func(c *Client) ([]byte, error) {
				form := url.Values{}
				form.Set("title", "hello")
				form.Set("fx123", bigValue)
				return c.PostFormURLEncoded("/p/item", form)
			},
		},
		{
			name:   "PutFormURLEncoded sends form body",
			method: http.MethodPut,
			call: func(c *Client) ([]byte, error) {
				form := url.Values{}
				form.Set("reason", "smoke")
				form.Set("fx456", bigValue)
				return c.PutFormURLEncoded("/p/item/REQ-1", form)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var seen struct {
				method      string
				path        string
				rawQuery    string
				contentType string
				auth        string
				fx          string
				meta        string
			}

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen.method = r.Method
				seen.path = r.URL.Path
				seen.rawQuery = r.URL.RawQuery
				seen.contentType = r.Header.Get("Content-Type")
				seen.auth = r.Header.Get("Authorization")
				if err := r.ParseForm(); err != nil {
					t.Errorf("ParseForm: %v", err)
				}
				for k, v := range r.PostForm {
					if strings.HasPrefix(k, "fx") && len(v) > 0 {
						seen.fx = v[0]
					}
					if (k == "title" || k == "reason") && len(v) > 0 {
						seen.meta = v[0]
					}
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c := New(srv.URL, "secret-token", false)
			if _, err := tc.call(c); err != nil {
				t.Fatalf("call: %v", err)
			}

			if seen.method != tc.method {
				t.Errorf("method = %q, want %q", seen.method, tc.method)
			}
			if seen.rawQuery != "" {
				t.Errorf("query string should be empty, got %q (large payload must not land in the URL)", seen.rawQuery)
			}
			if seen.contentType != "application/x-www-form-urlencoded" {
				t.Errorf("content-type = %q, want application/x-www-form-urlencoded", seen.contentType)
			}
			if seen.auth != "Token secret-token" {
				t.Errorf("authorization = %q, want %q", seen.auth, "Token secret-token")
			}
			if seen.fx != bigValue {
				t.Errorf("fx field roundtrip mismatch (len got=%d want=%d)", len(seen.fx), len(bigValue))
			}
			if seen.meta == "" {
				t.Errorf("expected metadata field (title/reason) to be present in form body")
			}
		})
	}
}
