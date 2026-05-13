package main

import (
	"context"
	"log"
	"net"

	"go-headers-middleware/propagation"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/metadata"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	log.Println("=== gRPC call received ===")
	log.Println("Metadata:", md)

	return &pb.HelloReply{
		Message: "Hello " + req.Name,
	}, nil
}

func main() {
	cfg := propagation.LoadConfig()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(propagation.UnaryServerInterceptor(cfg)),
		grpc.StreamInterceptor(propagation.StreamServerInterceptor(cfg)),
	)

	pb.RegisterGreeterServer(s, &server{})

	log.Println("gRPC server running on :50051")
	log.Fatal(s.Serve(lis))
}
