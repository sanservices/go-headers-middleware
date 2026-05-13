package propagation

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryServerInterceptorCaptures(t *testing.T) {
	cfg := Config{Headers: []string{"X-Deployment-Color"}}

	md := metadata.Pairs(
		"x-deployment-color", "green",
		"x-not-configured", "leak",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var captured http.Header
	handler := func(ctx context.Context, req any) (any, error) {
		captured = HeadersFromContext(ctx)
		return nil, nil
	}

	_, err := UnaryServerInterceptor(cfg)(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatal(err)
	}

	if captured.Get("X-Deployment-Color") != "green" {
		t.Errorf("expected captured X-Deployment-Color=green, got %v", captured)
	}
	if captured.Get("X-Not-Configured") != "" {
		t.Errorf("non-configured header leaked: %v", captured)
	}
}

func TestUnaryServerInterceptorDisabled(t *testing.T) {
	cfg := Config{}
	md := metadata.Pairs("x-deployment-color", "green")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req any) (any, error) {
		if got := HeadersFromContext(ctx); got != nil {
			t.Errorf("expected nil headers when cfg disabled, got %v", got)
		}
		return nil, nil
	}
	_, _ = UnaryServerInterceptor(cfg)(ctx, nil, &grpc.UnaryServerInfo{}, handler)
}

func TestUnaryServerInterceptorNoMetadata(t *testing.T) {
	cfg := Config{Headers: []string{"X-Deployment-Color"}}
	handler := func(ctx context.Context, req any) (any, error) {
		if got := HeadersFromContext(ctx); got != nil {
			t.Errorf("expected nil headers with no incoming metadata, got %v", got)
		}
		return nil, nil
	}
	_, _ = UnaryServerInterceptor(cfg)(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
}

func TestUnaryClientInterceptorInjects(t *testing.T) {
	ctx := WithHeaders(context.Background(), http.Header{
		"X-Deployment-Color": {"green"},
	})

	var seen metadata.MD
	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, _ := metadata.FromOutgoingContext(ctx)
		seen = md
		return nil
	}

	err := UnaryClientInterceptor()(ctx, "/svc/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatal(err)
	}

	if got := seen.Get("x-deployment-color"); len(got) != 1 || got[0] != "green" {
		t.Errorf("outgoing metadata = %v, want x-deployment-color:[green]", seen)
	}
}

func TestUnaryClientInterceptorEmptyContext(t *testing.T) {
	invokerCalled := false
	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		invokerCalled = true
		md, _ := metadata.FromOutgoingContext(ctx)
		if len(md) != 0 {
			t.Errorf("expected no outgoing metadata, got %v", md)
		}
		return nil
	}
	_ = UnaryClientInterceptor()(context.Background(), "/svc/M", nil, nil, nil, invoker)
	if !invokerCalled {
		t.Error("invoker was not called")
	}
}

// fakeServerStream satisfies grpc.ServerStream with a controllable context.
type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (f *fakeServerStream) Context() context.Context { return f.ctx }

func TestStreamServerInterceptorCaptures(t *testing.T) {
	cfg := Config{Headers: []string{"X-Deployment-Color"}}

	md := metadata.Pairs("x-deployment-color", "green")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ss := &fakeServerStream{ctx: ctx}

	var captured http.Header
	handler := func(srv any, stream grpc.ServerStream) error {
		captured = HeadersFromContext(stream.Context())
		return nil
	}

	err := StreamServerInterceptor(cfg)(nil, ss, &grpc.StreamServerInfo{}, handler)
	if err != nil {
		t.Fatal(err)
	}

	if captured.Get("X-Deployment-Color") != "green" {
		t.Errorf("StreamServerInterceptor did not capture: %v", captured)
	}
}

func TestStreamClientInterceptorInjects(t *testing.T) {
	ctx := WithHeaders(context.Background(), http.Header{
		"X-Deployment-Color": {"green"},
	})

	var seen metadata.MD
	streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		md, _ := metadata.FromOutgoingContext(ctx)
		seen = md
		return nil, nil
	}

	_, err := StreamClientInterceptor()(ctx, &grpc.StreamDesc{}, nil, "/svc/M", streamer)
	if err != nil {
		t.Fatal(err)
	}

	if got := seen.Get("x-deployment-color"); len(got) != 1 || got[0] != "green" {
		t.Errorf("outgoing metadata = %v, want x-deployment-color:[green]", seen)
	}
}
