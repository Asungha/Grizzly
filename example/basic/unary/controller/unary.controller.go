package controller

import (
	"context"

	unaryService "github.com/Asungha/Grizzly/example/basic/unary/service"
	pb "github.com/Asungha/Grizzly/example/basic/unary/stub"
)

type UnaryController struct {
	pb.UnimplementedUnaryServiceServer
	service *unaryService.UnaryService
}

func NewUnaryController() *UnaryController {
	return &UnaryController{service: unaryService.NewUnaryService()}
}

func (c *UnaryController) UnaryCall(ctx context.Context, req *pb.UnaryRequest) (*pb.UnaryResponse, error) {
	return c.service.Unary(ctx, req)
}
