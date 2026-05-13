package main

import (
	"fmt"
	"log"
	"net/http"

	"go-headers-middleware/propagation"
)

func main() {
	cfg := propagation.LoadConfig()

	mux := http.NewServeMux()

	// inbound endpoint
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Incoming headers:", r.Header)

		// simulate outbound call
		req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://localhost:8080/downstream", nil)

		propagation.InjectHTTPHeaders(r.Context(), req)

		fmt.Println("Outgoing headers:", req.Header)

		w.Write([]byte("ok"))
	})

	// downstream endpoint (simulates another service)
	mux.HandleFunc("/downstream", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Downstream received headers:", r.Header)
		w.Write([]byte("downstream ok"))
	})

	handler := propagation.HTTPMiddleware(cfg)(mux)

	fmt.Println("HTTP server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
