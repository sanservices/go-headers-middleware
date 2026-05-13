package propagation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMiddlewareCapturesConfiguredHeaders(t *testing.T) {
	cfg := Config{Headers: []string{"X-Deployment-Color", "X-Canary"}}

	var captured http.Header
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = HeadersFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Deployment-Color", "green")
	req.Header.Set("X-Canary", "true")
	req.Header.Set("X-Not-Configured", "should-not-propagate")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if captured.Get("X-Deployment-Color") != "green" {
		t.Errorf("X-Deployment-Color not captured: %v", captured)
	}
	if captured.Get("X-Canary") != "true" {
		t.Errorf("X-Canary not captured: %v", captured)
	}
	if captured.Get("X-Not-Configured") != "" {
		t.Errorf("X-Not-Configured leaked through: %v", captured)
	}
}

func TestHTTPMiddlewareCaseInsensitiveMatch(t *testing.T) {
	cfg := Config{Headers: []string{"x-deployment-color"}}

	var captured http.Header
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = HeadersFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Deployment-Color", "green")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if captured.Get("X-Deployment-Color") != "green" {
		t.Errorf("case-insensitive capture failed: %v", captured)
	}
}

func TestHTTPMiddlewareDisabledPassThrough(t *testing.T) {
	cfg := Config{}

	called := false
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := HeadersFromContext(r.Context()); got != nil {
			t.Errorf("expected nil context headers when disabled, got %v", got)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Deployment-Color", "green")

	handler.ServeHTTP(httptest.NewRecorder(), req)
	if !called {
		t.Error("downstream handler was not called")
	}
}

func TestInjectHTTPHeaders(t *testing.T) {
	ctx := WithHeaders(context.Background(), http.Header{
		"X-Deployment-Color": {"green"},
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
	InjectHTTPHeaders(ctx, req)

	if req.Header.Get("X-Deployment-Color") != "green" {
		t.Errorf("InjectHTTPHeaders did not set header: %v", req.Header)
	}
}

func TestInjectHTTPHeadersReplacesExisting(t *testing.T) {
	ctx := WithHeaders(context.Background(), http.Header{
		"X-Deployment-Color": {"green"},
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
	req.Header.Set("X-Deployment-Color", "blue") // stale outbound value
	InjectHTTPHeaders(ctx, req)

	values := req.Header.Values("X-Deployment-Color")
	if len(values) != 1 || values[0] != "green" {
		t.Errorf("InjectHTTPHeaders should replace existing value, got %v", values)
	}
}

func TestInjectHTTPHeadersNoContext(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	InjectHTTPHeaders(context.Background(), req)
	if len(req.Header) != 0 {
		t.Errorf("InjectHTTPHeaders should be a no-op when context is empty: %v", req.Header)
	}
}

type captureTransport struct {
	got http.Header
}

func (c *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.got = req.Header.Clone()
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: req}, nil
}

func TestTransportInjectsHeaders(t *testing.T) {
	base := &captureTransport{}
	rt := NewTransport(base)

	ctx := WithHeaders(context.Background(), http.Header{
		"X-Deployment-Color": {"green"},
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}

	if base.got.Get("X-Deployment-Color") != "green" {
		t.Errorf("Transport did not inject header into downstream request: %v", base.got)
	}
}

func TestTransportNilBaseUsesDefault(t *testing.T) {
	rt := NewTransport(nil)
	if rt.Base != nil {
		t.Errorf("NewTransport(nil).Base = %v, want nil (uses DefaultTransport at call time)", rt.Base)
	}
}
