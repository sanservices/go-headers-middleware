.PHONY: tidy test fmt vet build run-http run-grpc-server run-grpc-client

FORWARD_HEADERS ?= X-Deployment-Color,X-Canary

tidy:
	go mod tidy

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	go build ./...

run-http:
	FORWARD_HEADERS=$(FORWARD_HEADERS) go run ./examples/http-server

run-grpc-server:
	FORWARD_HEADERS=$(FORWARD_HEADERS) go run ./examples/grpc-server

run-grpc-client:
	FORWARD_HEADERS=$(FORWARD_HEADERS) go run ./examples/grpc-client
