package jev

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

const validResponse = `{
	"model": "jev-2",
	"answers": {
		"urgent": {"type": "noul", "noul": 0.9},
		"tone": {"type": "choice", "choice": "angry", "probabilities": {"angry": 0.7, "calm": 0.3}, "confidence": 0.6},
		"quality": {"type": "score", "score": 4, "legend": {"4": "good"}, "probabilities": {"4": 1}, "confidence": 0.95}
	},
	"usage": {"input_tokens": 12, "output_tokens": 3}
}`

func TestSystemOneRequestMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		req  SystemOneRequest
		want string
	}{
		{
			name: "defaults model",
			req: SystemOneRequest{
				State:     "hello",
				Questions: map[string]Question{"q": NoulQuestion{Instructions: "i"}},
			},
			want: `{"model":"jev-latest","state":"hello","questions":{"q":{"type":"noul","instructions":"i"}}}`,
		},
		{
			name: "explicit model kept",
			req:  SystemOneRequest{Model: "jev-2", State: "hello"},
			want: `{"model":"jev-2","state":"hello","questions":null}`,
		},
		{
			name: "structured state",
			req: SystemOneRequest{
				Model:     "m",
				State:     map[string]any{"messages": []string{"a", "b"}},
				Questions: map[string]Question{},
			},
			want: `{"model":"m","state":{"messages":["a","b"]},"questions":{}}`,
		},
		{
			name: "nil state",
			req:  SystemOneRequest{Model: "m"},
			want: `{"model":"m","state":null,"questions":null}`,
		},
		{
			name: "questions sorted by key",
			req: SystemOneRequest{
				Model: "m",
				Questions: map[string]Question{
					"b": ScoreQuestion{Instructions: "s", Criteria: []any{"x"}},
					"a": ChoiceQuestion{Instructions: "c", Criteria: map[string]any{"k": 1}},
				},
			},
			want: `{"model":"m","state":null,"questions":{"a":{"type":"choice","instructions":"c","criteria":{"k":1}},"b":{"type":"score","instructions":"s","criteria":["x"]}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestSystemOneRequestMarshalJSON_DoesNotMutate(t *testing.T) {
	req := SystemOneRequest{State: "x"}
	if _, err := json.Marshal(req); err != nil {
		t.Fatal(err)
	}
	if req.Model != "" {
		t.Errorf("Model = %q, marshal mutated the request", req.Model)
	}
}

func TestSystemOneRequestMarshalJSON_Error(t *testing.T) {
	_, err := json.Marshal(SystemOneRequest{State: make(chan int)})
	if err == nil {
		t.Fatal("expected error for unmarshalable state")
	}
}

func TestSystemOneResponseUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    SystemOneResponse
		wantErr string
	}{
		{
			name: "all answer types",
			data: validResponse,
			want: SystemOneResponse{
				Model: "jev-2",
				Answers: map[string]Answer{
					"urgent": &NoulAnswer{Noul: 0.9},
					"tone": &ChoiceAnswer{
						Choice:        "angry",
						Probabilities: map[string]float64{"angry": 0.7, "calm": 0.3},
						Confidence:    0.6,
					},
					"quality": &ScoreAnswer{
						Score:         4,
						Legend:        map[string]string{"4": "good"},
						Probabilities: map[string]float64{"4": 1},
						Confidence:    0.95,
					},
				},
				Usage: Usage{InputTokens: 12, OutputTokens: 3},
			},
		},
		{
			name: "empty answers",
			data: `{"model":"m","answers":{},"usage":{"input_tokens":0,"output_tokens":0}}`,
			want: SystemOneResponse{Model: "m", Answers: map[string]Answer{}},
		},
		{
			name: "unknown top level fields ignored",
			data: `{"model":"m","answers":{},"usage":{"input_tokens":1,"output_tokens":2},"id":"abc"}`,
			want: SystemOneResponse{Model: "m", Answers: map[string]Answer{}, Usage: Usage{1, 2}},
		},
		{
			name: "large token counts",
			data: `{"model":"m","answers":{},"usage":{"input_tokens":9007199254740993,"output_tokens":0}}`,
			want: SystemOneResponse{Model: "m", Answers: map[string]Answer{}, Usage: Usage{InputTokens: 9007199254740993}},
		},
		{
			name:    "missing model",
			data:    `{"answers":{},"usage":{"input_tokens":1,"output_tokens":2}}`,
			wantErr: `jev: decode response: missing or null field "model"`,
		},
		{
			name:    "null answers",
			data:    `{"model":"m","answers":null,"usage":{"input_tokens":1,"output_tokens":2}}`,
			wantErr: `jev: decode response: missing or null field "answers"`,
		},
		{
			name:    "missing usage",
			data:    `{"model":"m","answers":{}}`,
			wantErr: `jev: decode response: missing or null field "usage"`,
		},
		{
			name:    "usage missing input tokens",
			data:    `{"model":"m","answers":{},"usage":{"output_tokens":2}}`,
			wantErr: `jev: decode usage: missing or null field "input_tokens"`,
		},
		{
			name:    "usage null output tokens",
			data:    `{"model":"m","answers":{},"usage":{"input_tokens":1,"output_tokens":null}}`,
			wantErr: `jev: decode usage: missing or null field "output_tokens"`,
		},
		{
			name:    "usage not an object",
			data:    `{"model":"m","answers":{},"usage":"lots"}`,
			wantErr: "jev: decode usage:",
		},
		{
			name:    "usage fractional tokens",
			data:    `{"model":"m","answers":{},"usage":{"input_tokens":1.5,"output_tokens":2}}`,
			wantErr: "jev: decode usage:",
		},
		{
			name:    "answers not an object",
			data:    `{"model":"m","answers":[],"usage":{"input_tokens":1,"output_tokens":2}}`,
			wantErr: "jev: decode response:",
		},
		{
			name:    "model wrong type",
			data:    `{"model":1,"answers":{},"usage":{"input_tokens":1,"output_tokens":2}}`,
			wantErr: "jev: decode response:",
		},
		{
			name:    "bad answer names id",
			data:    `{"model":"m","answers":{"ok":{"type":"noul","noul":1},"bad":{"type":"noul"}},"usage":{"input_tokens":1,"output_tokens":2}}`,
			wantErr: `jev: decode answer "bad": decode noul answer: missing or null field "noul"`,
		},
		{
			name:    "unsupported answer type",
			data:    `{"model":"m","answers":{"x":{"type":"nope"}},"usage":{"input_tokens":1,"output_tokens":2}}`,
			wantErr: `jev: decode answer "x": unsupported answer type "nope"`,
		},
		{
			name:    "answer not an object",
			data:    `{"model":"m","answers":{"x":5},"usage":{"input_tokens":1,"output_tokens":2}}`,
			wantErr: `jev: decode answer "x": decode answer type:`,
		},
		{
			name:    "invalid json",
			data:    `{"model":`,
			wantErr: "unexpected end of JSON input",
		},
		{
			name:    "top level array",
			data:    `[]`,
			wantErr: "jev: decode response:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got SystemOneResponse
			err := json.Unmarshal([]byte(tt.data), &got)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got  %#v\nwant %#v", got, tt.want)
			}
		})
	}
}

func TestSystemOneResponseUnmarshalJSON_ReplacesExisting(t *testing.T) {
	got := SystemOneResponse{
		Model:   "old",
		Answers: map[string]Answer{"stale": &NoulAnswer{Noul: 1}},
		Usage:   Usage{99, 99},
	}
	data := `{"model":"new","answers":{"fresh":{"type":"noul","noul":0.5}},"usage":{"input_tokens":1,"output_tokens":2}}`
	if err := json.Unmarshal([]byte(data), &got); err != nil {
		t.Fatal(err)
	}
	want := SystemOneResponse{
		Model:   "new",
		Answers: map[string]Answer{"fresh": &NoulAnswer{Noul: 0.5}},
		Usage:   Usage{1, 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %#v\nwant %#v", got, want)
	}
}

func TestSystemOneResponseUnmarshalJSON_ErrorLeavesTargetUntouched(t *testing.T) {
	got := SystemOneResponse{Model: "old"}
	if err := json.Unmarshal([]byte(`{"model":"new","answers":{}}`), &got); err == nil {
		t.Fatal("expected error")
	}
	if got.Model != "old" {
		t.Errorf("Model = %q, want %q", got.Model, "old")
	}
}

type recordedRequest struct {
	method string
	path   string
	header http.Header
	body   []byte
}

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *recordedRequest) {
	t.Helper()

	rec := &recordedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		*rec = recordedRequest{
			method: r.Method,
			path:   r.URL.Path,
			header: r.Header.Clone(),
			body:   body,
		}
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(
		WithAPIKey("test-key"),
		WithBaseURL(srv.URL+"/v1"),
		WithHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, rec
}

func TestSystemOne_Success(t *testing.T) {
	c, rec := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, validResponse)
	})

	req := SystemOneRequest{
		State: "I want a refund",
		Questions: map[string]Question{
			"urgent": NoulQuestion{Instructions: "Urgent?"},
		},
	}
	resp, err := c.SystemOne(context.Background(), req)
	if err != nil {
		t.Fatalf("SystemOne: %v", err)
	}

	if rec.method != http.MethodPost {
		t.Errorf("method = %q, want POST", rec.method)
	}
	if rec.path != "/v1/systemone" {
		t.Errorf("path = %q, want /v1/systemone", rec.path)
	}
	wantHeaders := map[string]string{
		"Authorization": "Bearer test-key",
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}
	for k, want := range wantHeaders {
		if got := rec.header.Get(k); got != want {
			t.Errorf("header %s = %q, want %q", k, got, want)
		}
	}
	wantBody := `{"model":"jev-latest","state":"I want a refund","questions":{"urgent":{"type":"noul","instructions":"Urgent?"}}}`
	if string(rec.body) != wantBody {
		t.Errorf("body = %s\nwant %s", rec.body, wantBody)
	}

	if resp.Model != "jev-2" {
		t.Errorf("Model = %q, want jev-2", resp.Model)
	}
	if len(resp.Answers) != 3 {
		t.Errorf("len(Answers) = %d, want 3", len(resp.Answers))
	}
	if n, ok := resp.Answers["urgent"].(*NoulAnswer); !ok || n.Noul != 0.9 {
		t.Errorf("Answers[urgent] = %#v", resp.Answers["urgent"])
	}
	if resp.Usage != (Usage{12, 3}) {
		t.Errorf("Usage = %+v", resp.Usage)
	}
}

func TestSystemOne_BaseURLPaths(t *testing.T) {
	tests := []struct {
		name     string
		suffix   string
		wantPath string
	}{
		{"no path", "", "/systemone"},
		{"trailing slash", "/", "/systemone"},
		{"prefix", "/v1", "/v1/systemone"},
		{"prefix trailing slash", "/v1/", "/v1/systemone"},
		{"nested prefix", "/api/v2", "/api/v2/systemone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				io.WriteString(w, validResponse)
			}))
			defer srv.Close()

			c, err := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL+tt.suffix))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := c.SystemOne(context.Background(), SystemOneRequest{}); err != nil {
				t.Fatal(err)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestSystemOne_APIError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"bad request", http.StatusBadRequest, `{"error":"bad"}`},
		{"unauthorized", http.StatusUnauthorized, ``},
		{"rate limited", http.StatusTooManyRequests, `slow down`},
		{"server error", http.StatusInternalServerError, `oops`},
		{"redirect", http.StatusMovedPermanently, ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-Id", "req-1")
				w.WriteHeader(tt.status)
				io.WriteString(w, tt.body)
			})

			resp, err := c.SystemOne(context.Background(), SystemOneRequest{})
			if resp != nil {
				t.Errorf("resp = %+v, want nil", resp)
			}

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("err = %T %v, want *APIError", err, err)
			}
			if apiErr.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.status)
			}
			if string(apiErr.Body) != tt.body {
				t.Errorf("Body = %q, want %q", apiErr.Body, tt.body)
			}
			if apiErr.Header.Get("X-Request-Id") != "req-1" {
				t.Errorf("Header missing X-Request-Id: %v", apiErr.Header)
			}
			if apiErr.Unwrap() != nil {
				t.Errorf("Unwrap() = %v, want nil", apiErr.Unwrap())
			}
		})
	}
}

func TestSystemOne_EmptySuccessBody(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := c.SystemOne(context.Background(), SystemOneRequest{})
	if err == nil || !strings.Contains(err.Error(), "jev: decode evaluation response:") {
		t.Fatalf("err = %v, want decode error for empty 2xx body", err)
	}
}

func TestSystemOne_MalformedSuccessBody(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"invalid json", `{not json`, "jev: decode evaluation response:"},
		{"missing usage", `{"model":"m","answers":{}}`, `missing or null field "usage"`},
		{"bad answer", `{"model":"m","answers":{"a":{"type":"x"}},"usage":{"input_tokens":1,"output_tokens":1}}`, `decode answer "a"`},
		{"html", `<html>oops</html>`, "jev: decode evaluation response:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, tt.body)
			})

			resp, err := c.SystemOne(context.Background(), SystemOneRequest{})
			if resp != nil {
				t.Errorf("resp = %+v, want nil", resp)
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want containing %q", err, tt.want)
			}
			var apiErr *APIError
			if errors.As(err, &apiErr) {
				t.Error("decode failure should not be an *APIError")
			}
		})
	}
}

func TestSystemOne_ContextErrors(t *testing.T) {
	var nilCtx context.Context

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	expired, cancelExpired := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancelExpired()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
		wantMsg string
	}{
		{"nil context", nilCtx, nil, "jev: context must not be nil"},
		{"cancelled before call", cancelled, context.Canceled, "jev: evaluate:"},
		{"expired before call", expired, context.DeadlineExceeded, "jev: evaluate:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				called = true
			})

			resp, err := c.SystemOne(tt.ctx, SystemOneRequest{})
			if resp != nil {
				t.Errorf("resp = %+v, want nil", resp)
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantMsg) {
				t.Fatalf("err = %v, want containing %q", err, tt.wantMsg)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("errors.Is(err, %v) = false", tt.wantErr)
			}
			if called {
				t.Error("server was called despite bad context")
			}
		})
	}
}

func TestSystemOne_ContextCancelledDuringRequest(t *testing.T) {
	release := make(chan struct{})
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-release
	})
	t.Cleanup(func() { close(release) })

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.SystemOne(ctx, SystemOneRequest{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
	if !strings.HasPrefix(err.Error(), "jev: send request:") {
		t.Errorf("err = %q, want send request prefix", err)
	}
}

func TestSystemOne_EncodeRequestError(t *testing.T) {
	called := false
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	_, err := c.SystemOne(context.Background(), SystemOneRequest{State: make(chan int)})
	if err == nil || !strings.HasPrefix(err.Error(), "jev: encode request:") {
		t.Fatalf("err = %v, want encode request error", err)
	}
	if called {
		t.Error("server was called despite encode failure")
	}
}

func TestSystemOne_TransportError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	c, err := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.SystemOne(context.Background(), SystemOneRequest{})
	if err == nil || !strings.HasPrefix(err.Error(), "jev: send request:") {
		t.Fatalf("err = %v, want send request error", err)
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		t.Error("transport failure should not be an *APIError")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }
func (failingReader) Close() error               { return nil }

func TestSystemOne_BodyReadFailure(t *testing.T) {
	readErr := errors.New("connection reset")

	newClient := func(t *testing.T, status int) *Client {
		t.Helper()
		hc := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: status,
				Header:     http.Header{"X-Test": []string{"1"}},
				Body:       failingReader{err: readErr},
				Request:    r,
			}, nil
		})}
		c, err := NewClient(WithAPIKey("k"), WithHTTPClient(hc))
		if err != nil {
			t.Fatal(err)
		}
		return c
	}

	t.Run("success status", func(t *testing.T) {
		c := newClient(t, http.StatusOK)
		_, err := c.SystemOne(context.Background(), SystemOneRequest{})
		if !errors.Is(err, readErr) {
			t.Fatalf("err = %v, want wrapped read error", err)
		}
		if !strings.HasPrefix(err.Error(), "jev: read response:") {
			t.Errorf("err = %q", err)
		}
	})

	t.Run("error status", func(t *testing.T) {
		c := newClient(t, http.StatusBadGateway)
		_, err := c.SystemOne(context.Background(), SystemOneRequest{})

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("err = %T %v, want *APIError", err, err)
		}
		if apiErr.StatusCode != http.StatusBadGateway {
			t.Errorf("StatusCode = %d", apiErr.StatusCode)
		}
		if !errors.Is(err, readErr) {
			t.Error("APIError should unwrap to the read error")
		}
		if apiErr.Header.Get("X-Test") != "1" {
			t.Errorf("Header = %v", apiErr.Header)
		}
		if want := "jev: API returned HTTP 502; failed to read response body"; err.Error() != want {
			t.Errorf("err = %q, want %q", err, want)
		}
	})
}

func TestSystemOne_DoesNotSendAPIKeyInURL(t *testing.T) {
	c, rec := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("unexpected query: %q", r.URL.RawQuery)
		}
		io.WriteString(w, validResponse)
	})

	if _, err := c.SystemOne(context.Background(), SystemOneRequest{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(rec.path, "test-key") {
		t.Errorf("api key leaked into path %q", rec.path)
	}
}
