package service

import (
	"context"

	repo "github.com/Asungha/Grizzly/example/basic/unary/repository"
	pb "github.com/Asungha/Grizzly/example/basic/unary/stub"
)

type UnaryService struct {
	pb.UnimplementedUnaryServiceServer
	database *repo.UnaryRepository
}

func NewUnaryService() *UnaryService {
	return &UnaryService{}
}

func (s *UnaryService) Unary(ctx context.Context, req *pb.UnaryRequest) (*pb.UnaryResponse, error) {
	data := s.database.GetData()
	return &pb.UnaryResponse{Message: data}, nil
}
