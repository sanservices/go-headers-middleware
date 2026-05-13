package propagation

import (
	"context"
	"net/http"
	"testing"
)

func TestHeadersFromContextEmpty(t *testing.T) {
	if got := HeadersFromContext(context.Background()); got != nil {
		t.Errorf("HeadersFromContext(empty) = %#v, want nil", got)
	}
}

func TestWithHeadersRoundTrip(t *testing.T) {
	in := http.Header{
		"X-Deployment-Color": {"green"},
		"X-Canary":           {"true"},
	}

	ctx := WithHeaders(context.Background(), in)
	got := HeadersFromContext(ctx)

	if got.Get("X-Deployment-Color") != "green" {
		t.Errorf("X-Deployment-Color = %q, want green", got.Get("X-Deployment-Color"))
	}
	if got.Get("X-Canary") != "true" {
		t.Errorf("X-Canary = %q, want true", got.Get("X-Canary"))
	}
}

func TestWithHeadersDeepCopyOnStore(t *testing.T) {
	in := http.Header{"X-A": {"one"}}
	ctx := WithHeaders(context.Background(), in)

	// Mutate the original after storing — stored value must be unaffected.
	in["X-A"][0] = "mutated"
	in["X-B"] = []string{"injected"}

	got := HeadersFromContext(ctx)
	if got.Get("X-A") != "one" {
		t.Errorf("stored header changed after mutating input: got %q", got.Get("X-A"))
	}
	if got.Get("X-B") != "" {
		t.Errorf("stored map changed after adding to input: got %q", got.Get("X-B"))
	}
}

func TestHeadersFromContextDeepCopyOnRead(t *testing.T) {
	in := http.Header{"X-A": {"one"}}
	ctx := WithHeaders(context.Background(), in)

	first := HeadersFromContext(ctx)
	first["X-A"][0] = "mutated"
	first["X-B"] = []string{"injected"}

	second := HeadersFromContext(ctx)
	if second.Get("X-A") != "one" {
		t.Errorf("stored header changed after mutating reader copy: got %q", second.Get("X-A"))
	}
	if second.Get("X-B") != "" {
		t.Errorf("stored map changed after adding to reader copy: got %q", second.Get("X-B"))
	}
}
