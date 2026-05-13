// Package propagation forwards a configured set of headers (e.g.
// X-Deployment-Color, X-Tenant-Id) across HTTP and gRPC request boundaries.
//
// The list of headers to propagate is read from the FORWARD_HEADERS
// environment variable as a comma-separated list. Inbound middleware and
// interceptors capture matching headers into the request context; outbound
// helpers and interceptors read from that context and inject them onto
// downstream requests.
//
// Typical wiring on the server side:
//
//	cfg := propagation.LoadConfig()
//	handler = propagation.HTTPMiddleware(cfg)(handler)
//
//	srv := grpc.NewServer(
//	    grpc.UnaryInterceptor(propagation.UnaryServerInterceptor(cfg)),
//	    grpc.StreamInterceptor(propagation.StreamServerInterceptor(cfg)),
//	)
//
// Typical wiring on the client side:
//
//	httpClient := &http.Client{Transport: propagation.NewTransport(nil)}
//
//	conn, _ := grpc.NewClient(addr,
//	    grpc.WithUnaryInterceptor(propagation.UnaryClientInterceptor()),
//	    grpc.WithStreamInterceptor(propagation.StreamClientInterceptor()),
//	)
package propagation
