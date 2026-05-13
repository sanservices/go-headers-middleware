package propagation

import (
	"context"
	"net/http"
)

// HTTPMiddleware captures configured headers from inbound requests and stores
// them in the request context.
func HTTPMiddleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled() {
				next.ServeHTTP(w, r)
				return
			}

			captured := make(http.Header)

			for _, headerName := range cfg.Headers {
				values := r.Header.Values(headerName)
				for _, v := range values {
					captured.Add(headerName, v)
				}
			}
			if len(captured) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			ctx := WithHeaders(r.Context(), captured)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// InjectHTTPHeaders copies stored headers from ctx into the outgoing request,
// replacing any existing values for those header names.
func InjectHTTPHeaders(ctx context.Context, req *http.Request) {
	headers := HeadersFromContext(ctx)
	if headers == nil {
		return
	}

	for key, values := range headers {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

// Transport wraps an http.RoundTripper so every outgoing request automatically
// gets the headers stored in its context (via WithHeaders or HTTPMiddleware).
// If Base is nil, http.DefaultTransport is used.
type Transport struct {
	Base http.RoundTripper
}

// NewTransport returns a Transport that delegates to base (or DefaultTransport
// if base is nil) after injecting stored headers from the request context.
func NewTransport(base http.RoundTripper) *Transport {
	return &Transport{Base: base}
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	InjectHTTPHeaders(req.Context(), req)
	return base.RoundTrip(req)
}
