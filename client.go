package jev

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	defaultBaseURL = "https://api.typesafe.ai/v1"
)

var (
	ErrNilClientOpts  = errors.New("jev: client option must not be nil")
	ErrAPIKeyMissing  = errors.New("jev: API key is required; set TYPESAFE_API_KEY or use WithAPIKey")
	ErrNilHTTPClient  = errors.New("jev: HTTP client must not be nil")
	ErrInvalidBaseURL = errors.New("jev: invalid base URL")
)

// Client holds all the configuration for API requests
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    *url.URL
}

// clientConfig is temporary configuration and is only used during construction.
type clientConfig struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// Option configures a client during NewClient
type Option func(*clientConfig)

// WithAPIKey overrides the API key read from TYPESAFE_API_KEY.
func WithAPIKey(key string) Option {
	return func(cfg *clientConfig) {
		cfg.apiKey = key
	}
}

// WithHTTPClient replaces the default HTTP Client.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) {
		cfg.httpClient = client
	}
}

// WithBaseURL overrides the API base URL.
func WithBaseURL(baseURL string) Option {
	return func(cfg *clientConfig) {
		cfg.baseURL = baseURL
	}
}

// NewClient constructs a Jev client from defaults and provided options.
//
// Reads TYPESAFE_API_KEY unless WithAPIKey is provided.
func NewClient(opts ...Option) (*Client, error) {
	cfg := clientConfig{
		apiKey:  os.Getenv("TYPESAFE_API_KEY"),
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	for _, opt := range opts {
		if opt == nil {
			return nil, ErrNilClientOpts
		}
		opt(&cfg)
	}

	// Reject whitespace-only credentials
	if strings.TrimSpace(cfg.apiKey) == "" {
		return nil, ErrAPIKeyMissing
	}

	if cfg.httpClient == nil {
		return nil, ErrNilHTTPClient
	}

	baseURL, err := parseBaseURL(cfg.baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:     cfg.apiKey,
		httpClient: cfg.httpClient,
		baseURL:    baseURL,
	}, nil
}

// parseBaseURL validates the structural requirements of an API base URL.
func parseBaseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed URL", ErrInvalidBaseURL)
	}

	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, fmt.Errorf("%w: URL was not HTTP(S) or had an invalid hostname", ErrInvalidBaseURL)
	}

	if u.User != nil {
		return nil, fmt.Errorf("%w: credentials must not appear in the URL", ErrInvalidBaseURL)
	}

	return u, nil
}
