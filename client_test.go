package jev

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	custom := &http.Client{Timeout: time.Second}

	tests := []struct {
		name    string
		env     string
		opts    []Option
		wantErr error
		check   func(t *testing.T, c *Client)
	}{
		{
			name: "api key from env",
			env:  "env-key",
			check: func(t *testing.T, c *Client) {
				if c.apiKey != "env-key" {
					t.Errorf("apiKey = %q, want %q", c.apiKey, "env-key")
				}
			},
		},
		{
			name: "option overrides env",
			env:  "env-key",
			opts: []Option{WithAPIKey("opt-key")},
			check: func(t *testing.T, c *Client) {
				if c.apiKey != "opt-key" {
					t.Errorf("apiKey = %q, want %q", c.apiKey, "opt-key")
				}
			},
		},
		{
			name:    "missing api key",
			wantErr: ErrAPIKeyMissing,
		},
		{
			name:    "whitespace api key",
			env:     "  \t\n",
			wantErr: ErrAPIKeyMissing,
		},
		{
			name:    "empty option key does not fall back to env",
			env:     "env-key",
			opts:    []Option{WithAPIKey("")},
			wantErr: ErrAPIKeyMissing,
		},
		{
			name:    "nil option",
			env:     "env-key",
			opts:    []Option{WithAPIKey("k"), nil},
			wantErr: ErrNilClientOpts,
		},
		{
			name:    "nil http client",
			env:     "env-key",
			opts:    []Option{WithHTTPClient(nil)},
			wantErr: ErrNilHTTPClient,
		},
		{
			name:    "invalid base url",
			env:     "env-key",
			opts:    []Option{WithBaseURL("ftp://example.com")},
			wantErr: ErrInvalidBaseURL,
		},
		{
			name: "defaults",
			env:  "env-key",
			check: func(t *testing.T, c *Client) {
				if got := c.baseURL.String(); got != defaultBaseURL {
					t.Errorf("baseURL = %q, want %q", got, defaultBaseURL)
				}
				if c.httpClient == nil || c.httpClient.Timeout != defaultTimeout {
					t.Errorf("httpClient = %+v, want timeout %v", c.httpClient, defaultTimeout)
				}
			},
		},
		{
			name: "custom http client and base url",
			env:  "env-key",
			opts: []Option{WithHTTPClient(custom), WithBaseURL("http://localhost:8080/api/")},
			check: func(t *testing.T, c *Client) {
				if c.httpClient != custom {
					t.Error("httpClient was not the provided client")
				}
				if got := c.baseURL.String(); got != "http://localhost:8080/api/" {
					t.Errorf("baseURL = %q", got)
				}
			},
		},
		{
			name: "last option wins",
			env:  "env-key",
			opts: []Option{WithAPIKey("first"), WithAPIKey("second")},
			check: func(t *testing.T, c *Client) {
				if c.apiKey != "second" {
					t.Errorf("apiKey = %q, want %q", c.apiKey, "second")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TYPESAFE_API_KEY", tt.env)

			c, err := NewClient(tt.opts...)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if c != nil {
					t.Errorf("client = %+v, want nil on error", c)
				}
				return
			}
			if c == nil {
				t.Fatal("client is nil")
			}
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
}

func TestParseBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "https", raw: "https://api.typesafe.ai/v1"},
		{name: "http", raw: "http://localhost:8080"},
		{name: "trailing slash", raw: "https://example.com/v1/"},
		{name: "ipv6 host", raw: "http://[::1]:9000/v1"},
		{name: "query and fragment", raw: "https://example.com/v1?x=1#frag"},
		{name: "empty", raw: "", wantErr: true},
		{name: "malformed", raw: "http://exa mple.com", wantErr: true},
		{name: "control character", raw: "https://example.com/\x7f", wantErr: true},
		{name: "no scheme", raw: "api.typesafe.ai/v1", wantErr: true},
		{name: "scheme only", raw: "https://", wantErr: true},
		{name: "unsupported scheme", raw: "ftp://example.com", wantErr: true},
		{name: "relative path", raw: "/v1", wantErr: true},
		{name: "credentials", raw: "https://user:pass@example.com", wantErr: true},
		{name: "username only", raw: "https://user@example.com", wantErr: true},
		{name: "empty userinfo", raw: "https://@example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := parseBaseURL(tt.raw)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidBaseURL) {
					t.Fatalf("err = %v, want ErrInvalidBaseURL", err)
				}
				if u != nil {
					t.Errorf("url = %v, want nil", u)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.String() != tt.raw {
				t.Errorf("url = %q, want %q", u, tt.raw)
			}
		})
	}
}
