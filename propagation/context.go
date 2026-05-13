package propagation

import (
	"context"
	"net/http"
)

type contextKey struct{}

var headersContextKey = contextKey{}

// WithHeaders stores a copy of the provided headers in the context.
func WithHeaders(ctx context.Context, headers http.Header) context.Context {
	copyHeaders := make(http.Header)

	for k, v := range headers {
		copyHeaders[k] = append([]string(nil), v...)
	}

	return context.WithValue(ctx, headersContextKey, copyHeaders)
}

// HeadersFromContext returns a copy of stored headers.
func HeadersFromContext(ctx context.Context) http.Header {
	stored, ok := ctx.Value(headersContextKey).(http.Header)
	if !ok || stored == nil {
		return nil
	}

	copyHeaders := make(http.Header)
	for k, values := range stored {
		copied := append([]string(nil), values...)
		copyHeaders[k] = copied
	}

	return copyHeaders
}
