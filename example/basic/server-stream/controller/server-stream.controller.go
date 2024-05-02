package controller

import (
	"context"
	"log"

	eventbroker "github.com/Asungha/Grizzly/eventbroker"
	serverStreamService "github.com/Asungha/Grizzly/example/basic/server-stream/service"
	pb "github.com/Asungha/Grizzly/example/basic/server-stream/stub"

	controller "github.com/Asungha/Grizzly/controller"
)

type ServerStreamController struct {
	pb.UnimplementedServerStreamServiceServer
	service *serverStreamService.ServerStreamService

	// You might have multiple event brokers in your application for different event sources.
	EventBroker eventbroker.IEventBroker[*pb.Data]
}

func NewServerStreamController(eventBroker eventbroker.IEventBroker[*pb.Data]) *ServerStreamController {
	return &ServerStreamController{service: serverStreamService.NewServerStreamService(), EventBroker: eventBroker}
}

// Argument of this function is the same as normal gRPC function
// Read more about gRPC server stream here: https://grpc.io/docs/languages/go/basics/#server-streaming-rpc
func (c *ServerStreamController) ServerStreamCall(reqData *pb.ServerStreamRequest, stream pb.ServerStreamService_ServerStreamCallServer) error {
	// Create a new handler for this stream
	// This will defind how the stream will be handled

	// It take input as a data it listening in eventbroker and output as a response to the client
	// In this example, we are listening to *pb.Data and sending *pb.ServerStreamResponse to the client
	handler := controller.NewServerStreamHandler[*pb.Data, *pb.ServerStreamResponse]()
	handler.SetConnectHandler(func() error {
		log.Println("Client connected")
		return nil
	})
	handler.SetDataHandler(func(d *pb.Data) (*pb.ServerStreamResponse, error) {
		log.Println("Data received: ", d)
		return c.service.ServerStreamCall(context.Background(), d)
	})

	// Create task for this stream
	// This will tell where data is coming from and how to handle it
	// If you have multiple event brokers, you can create multiple handlers and tasks for each broker.
	task := &controller.ServerStreamTask[*pb.Data, *pb.ServerStreamResponse]{}
	task.SetHandler(handler)
	task.Seteventbroker(c.EventBroker)

	// Create a binder for this stream
	binder := controller.NewServerStreamBinder(&controller.ServerStreamWrapper[*pb.ServerStreamResponse]{Stream: stream})

	// Bind the binder with the task
	// You can bind multiple tasks with a single binder. And it's will be handled concurrently in the same stream.
	controller.Bind[*pb.Data, *pb.ServerStreamResponse](binder, task)

	// This will start the stream until the client disconnects or error occurs
	// This function is blocking. Only 1 binder can be used in a single stream.
	return binder.Serve()
}
