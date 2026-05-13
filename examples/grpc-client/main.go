package main

import (
	"context"
	"log"
	"time"

	"github.com/sanservices/go-headers-middleware/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(propagation.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(propagation.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	ctx := context.Background()

	ctx = propagation.WithHeaders(ctx, map[string][]string{
		"X-Deployment-Color": {"green"},
	})

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	res, err := client.SayHello(ctx, &pb.HelloRequest{
		Name: "test",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Response:", res.Message)
}
