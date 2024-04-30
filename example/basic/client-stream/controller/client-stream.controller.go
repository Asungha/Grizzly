package controller

import (
	"context"
	"log"
	"time"

	clientStreamService "github.com/Asungha/Grizzly/example/basic/client-stream/service"
	pb "github.com/Asungha/Grizzly/example/basic/client-stream/stub"

	controller "github.com/Asungha/Grizzly/controller"
)

type ClientStreamController struct {
	pb.UnimplementedClientStreamServiceServer
	service *clientStreamService.ClientStreamService
}

func NewClientStreamController() *ClientStreamController {
	return &ClientStreamController{service: clientStreamService.NewUnaryService()}
}

func (c *ClientStreamController) ClientStreamCall(stream pb.ClientStreamService_ClientStreamCallServer) error {
	handler := controller.NewClientStreamHandler[*pb.ClientStreamRequest, *pb.ClientStreamResponse]()
	handler.OnConnect(func() error {
		log.Println("Client connected")
		return nil
	})
	handler.OnData(func(data *pb.ClientStreamRequest) error {
		log.Println("Data received: ", data)
		return nil
	})
	handler.OnStreamEnd(func(csr *pb.ClientStreamRequest, i int) (*pb.ClientStreamResponse, error) {
		log.Println("Stream ended. Sending response data")
		ctx, cancle := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
		defer cancle()
		return c.service.ClientStreamCall(ctx, csr)
	})
	handler.OnDone(func() error {
		log.Println("Client disconnected")
		return nil
	})
	handler.OnError(func(csr *pb.ClientStreamRequest, err error) error {
		log.Println("Error: ", err)
		return nil
	})

	return controller.ClientStreamServer[*pb.ClientStreamRequest, *pb.ClientStreamResponse](
		stream,
		handler.Export(),
		nil,
		nil,
		controller.FunctionOptionsBuilder().InputValidation(false).OutputValidation(false),
	)

}
