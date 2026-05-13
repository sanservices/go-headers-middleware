# poc-header-middleware

[![Go Reference](https://pkg.go.dev/badge/github.com/sanservices/poc-header-middleware.svg)](https://pkg.go.dev/github.com/sanservices/poc-header-middleware)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A small Go library for propagating a configured list of headers (for example
`X-Deployment-Color`, `X-Tenant-Id`, `X-Canary`) across HTTP and gRPC request
boundaries.

It captures matching headers from inbound requests into the request context,
and injects them back onto outbound HTTP requests and gRPC metadata so the
same routing or tenancy information flows end-to-end.

## Install

```bash
go get github.com/sanservices/poc-header-middleware
```

Requires Go 1.24 or newer.

## Configuration

Set the `FORWARD_HEADERS` environment variable with a comma-separated list of
header names to propagate. Matching is case-insensitive.

```bash
export FORWARD_HEADERS=X-Deployment-Color,X-Canary
```

If the variable is empty or unset the middleware/interceptors become no-ops on
the server side.

## HTTP usage

### Server (capture inbound headers)

```go
import (
    "net/http"

    "github.com/sanservices/poc-header-middleware/propagation"
)

func main() {
    cfg := propagation.LoadConfig()

    mux := http.NewServeMux()
    mux.HandleFunc("/hello", helloHandler)

    handler := propagation.HTTPMiddleware(cfg)(mux)
    _ = http.ListenAndServe(":8080", handler)
}
```

### Client (inject into outbound requests)

The easiest path is to wrap your `http.Client`'s transport once:

```go
client := &http.Client{
    Transport: propagation.NewTransport(nil), // nil → http.DefaultTransport
}

req, _ := http.NewRequestWithContext(ctx, http.MethodGet, downstreamURL, nil)
resp, err := client.Do(req)
```

For one-off requests you can call the helper directly:

```go
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, downstreamURL, nil)
propagation.InjectHTTPHeaders(ctx, req)
```

`InjectHTTPHeaders` *replaces* any existing values for the headers it manages,
so calling it twice is safe.

## gRPC usage

### Server

```go
srv := grpc.NewServer(
    grpc.UnaryInterceptor(propagation.UnaryServerInterceptor(cfg)),
    grpc.StreamInterceptor(propagation.StreamServerInterceptor(cfg)),
)
```

### Client

```go
conn, _ := grpc.NewClient(addr,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithUnaryInterceptor(propagation.UnaryClientInterceptor()),
    grpc.WithStreamInterceptor(propagation.StreamClientInterceptor()),
)
```

Note: gRPC metadata keys are always lowercased on the wire, regardless of how
they were written in `FORWARD_HEADERS`.

## Seeding the context manually

If a request enters your process from somewhere other than an HTTP/gRPC server
(e.g. a Kafka consumer or a CLI), seed the context yourself:

```go
ctx = propagation.WithHeaders(ctx, http.Header{
    "X-Deployment-Color": {"green"},
})
```

From that point on, both `propagation.Transport` and the gRPC client
interceptors will forward the headers automatically.

## How it works

```
                  ┌────────────────────────────┐
inbound request → │ HTTPMiddleware / Unary…    │ → ctx with captured headers
                  │ ServerInterceptor          │
                  └────────────────────────────┘

ctx with headers  ┌────────────────────────────┐
       → outbound │ Transport / Unary…         │ → request with injected
                  │ ClientInterceptor          │   headers / metadata
                  └────────────────────────────┘
```

The captured headers are stored in the context under a private key, so they
travel through any normal `context.WithValue` / `r.WithContext` boundary
without leaking into other libraries.

## Examples

See [examples/](examples/) for runnable HTTP and gRPC demos. With the Makefile:

```bash
make run-http
make run-grpc-server      # in one terminal
make run-grpc-client      # in another
```

## License

[MIT](LICENSE) — see the LICENSE file for details.
