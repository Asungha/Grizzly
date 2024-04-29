package main

import (
	"log"
	"net"

	controller "github.com/Asungha/Grizzly/example/basic/unary/controller"
	"google.golang.org/grpc"

	pb "github.com/Asungha/Grizzly/example/basic/unary/stub"
)

func main() {
	lis, err := net.Listen("tcp", ":6001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	c := controller.NewUnaryController()

	pb.RegisterUnaryServiceServer(s, c)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
