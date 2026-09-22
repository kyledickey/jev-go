package jev

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

var (
	// ErrNilClientOpts means one or more of the options provided was nil
	ErrNilClientOpts = errors.New("jev: client option must not be nil")
	// ErrAPIKeyMissing means the API key was either missing or whitespace-only
	ErrAPIKeyMissing = errors.New("jev: API key is required; set TYPESAFE_API_KEY or use WithAPIKey")
	// ErrNilHTTPClient means the provided HTTP client was nil
	ErrNilHTTPClient = errors.New("jev: HTTP client must not be nil")
)

// Client holds all the configuration for Jev API requests
//
// Create clients with NewClient; the zero value is not ready for use.
// Reuse a client across requests.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// clientConfig is temporary configuration and is only used during construction.
type clientConfig struct {
	apiKey     string
	httpClient *http.Client
}

// Option configures a client during NewClient
type Option func(*clientConfig)

// WithAPIKey overrides the API key read from TYPESAFE_API_KEY
//
// An empty key also overrides the environment. NewClient will reject that
// empty key.
func WithAPIKey(key string) Option {
	return func(cfg *clientConfig) {
		cfg.apiKey = key
	}
}

// WithHTTPClient replaces the default HTTP Client
//
// The supplied client is used as-is: the SDK does not copy it or change its
// timeout, transport, or redirect policy.
//
// NewClient rejects a nil client. The caller should avoid changing the client
// while it's handling requests.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) {
		cfg.httpClient = client
	}
}

// NewClient constructs a Jev client from defaults and provided options
//
// The API key defaults to the value of TYPESAFE_API_KEY. Changing the
// environment afterward does not affect this client. The client does
// not check the validity of the API key.
func NewClient(opts ...Option) (*Client, error) {
	cfg := clientConfig{
		apiKey: os.Getenv("TYPESAFE_API_KEY"),
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

	return &Client{
		apiKey:     cfg.apiKey,
		httpClient: cfg.httpClient,
	}, nil
}
