package service

import (
	"context"

	repo "github.com/Asungha/Grizzly/example/basic/client-stream/repository"
	pb "github.com/Asungha/Grizzly/example/basic/client-stream/stub"
)

type ClientStreamService struct {
	pb.UnimplementedClientStreamServiceServer
	database *repo.ClientStreamRepository
}

func NewUnaryService() *ClientStreamService {
	return &ClientStreamService{}
}

func (s *ClientStreamService) ClientStreamCall(ctx context.Context, req *pb.ClientStreamRequest) (*pb.ClientStreamResponse, error) {
	data := s.database.GetData()
	return &pb.ClientStreamResponse{Message: data}, nil
}
