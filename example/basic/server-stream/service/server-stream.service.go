package service

import (
	"context"

	repo "github.com/Asungha/Grizzly/example/basic/server-stream/repository"
	pb "github.com/Asungha/Grizzly/example/basic/server-stream/stub"
)

type ServerStreamService struct {
	pb.UnimplementedServerStreamServiceServer
	database *repo.ServerStreamRepository
}

func NewServerStreamService() *ServerStreamService {
	return &ServerStreamService{}
}

func (s *ServerStreamService) ServerStreamCall(ctx context.Context, req *pb.Data) (*pb.ServerStreamResponse, error) {
	data := s.database.GetData()
	return &pb.ServerStreamResponse{Message: data}, nil
}
