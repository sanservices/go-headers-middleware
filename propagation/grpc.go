package propagation

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor captures configured metadata from incoming gRPC
// requests and stores it in the context.
func UnaryServerInterceptor(cfg Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx = captureFromIncomingMetadata(ctx, cfg)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor is the streaming counterpart of
// UnaryServerInterceptor. It captures configured metadata from the incoming
// stream and exposes it via the wrapped ServerStream context.
func StreamServerInterceptor(cfg Config) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := captureFromIncomingMetadata(ss.Context(), cfg)
		return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})
	}
}

// UnaryClientInterceptor injects stored headers into outgoing gRPC metadata.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = appendToOutgoingMetadata(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor is the streaming counterpart of
// UnaryClientInterceptor. It injects stored headers into the outgoing
// stream's metadata at stream creation time.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = appendToOutgoingMetadata(ctx)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func captureFromIncomingMetadata(ctx context.Context, cfg Config) context.Context {
	if !cfg.Enabled() {
		return ctx
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	captured := make(http.Header)
	for _, headerName := range cfg.Headers {
		values := md.Get(strings.ToLower(headerName))
		for _, v := range values {
			captured.Add(headerName, v)
		}
	}

	if len(captured) == 0 {
		return ctx
	}
	return WithHeaders(ctx, captured)
}

func appendToOutgoingMetadata(ctx context.Context) context.Context {
	headers := HeadersFromContext(ctx)
	if len(headers) == 0 {
		return ctx
	}

	pairs := make([]string, 0, len(headers)*2)
	for key, values := range headers {
		key = strings.ToLower(key)
		for _, value := range values {
			pairs = append(pairs, key, value)
		}
	}

	if len(pairs) == 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context { return w.ctx }
