package main

import (
	"context"
	"log"
	"net"

	pb "github.com/Asungha/Grizzly/asset_stub"
	"github.com/Asungha/Grizzly/eventbus"
	"google.golang.org/grpc"

	controller "github.com/Asungha/Grizzly/controller"

	"github.com/google/uuid"
)

func SUpload(data *pb.ImageUploadRequest) error {
	log.Printf("Upload: %v", data.ChunkId)
	return nil
}

// Service struct according to google grpc service

type AssetService struct {
	pb.UnimplementedAssetServiceServer
}

type ImageService struct {
	pb.UnimplementedImageServiceServer
	EventRepository *eventbus.IEventRepository[*pb.ImageUploadRequest, *pb.ImageUploadResponse]
}

type LogService struct {
	pb.UnimplementedLogStreamServiceServer
	EventRepository *eventbus.IEventRepository[*pb.ImageUploadRequest, *pb.ImageUploadResponse]
}

// GRPC function impplementation example

// Unary function
func (s *AssetService) GetAsset(ctx context.Context, data *pb.GetAssetRequest) (*pb.GetAssetResponse, error) {
	builder := controller.NewUnaryBuilder[*pb.GetAssetRequest, *pb.GetAssetResponse]()
	// What to do when there's a request from client
	// Request data will be passed to the function anf response will be returned immediately
	builder.OnData(func(ctx context.Context, gar *pb.GetAssetRequest) (*pb.GetAssetResponse, error) {
		return &pb.GetAssetResponse{Asset: &pb.Asset{}}, nil
	})

	// This is a non-blocking function
	return controller.UnaryServer(ctx, data, builder.Build(), controller.FunctionOptionsBuilder().InputValidation(false).OutputValidation(true))
}

// Client Stream function
func (s *ImageService) Upload(stream pb.ImageService_UploadServer) error {
	// Create topic bus for client stream
	// So other process can listen to the event via the EventBus
	// Use this only if you want to broadcast the event to other process. Otherwise, you can skip this to save resources.

	// If eventbus is connfigured to allow response pipe subscription, then the response pipe will be created and can be used to broadcast the event.
	// Same applies to request pipe subscription.
	// Server function that use said eventbus will send data respectively to the eventbus config without manual configuration.
	topicId := uuid.New().String()
	bus, _ := (*s.EventRepository).CreateTopic(
		topicId,
		eventbus.EventBusConfig{AllowRequestPipeSubscription: false, AllowResponsePipeSubscription: true},
	)

	// Create client stream builder
	builder := controller.NewClientStreamBuilder[*pb.ImageUploadRequest, *pb.ImageUploadResponse]()
	// What to do when client connect to the server
	builder.OnConnect(func() error {
		log.Printf("Create topic: %v", topicId)
		return nil
	})
	// What to do when client send a chunk of data
	// this will be called for each chunk of data. Which each request consist of multiple data chunks.
	builder.OnData(func(iur *pb.ImageUploadRequest) error {
		buff := &pb.ImageUploadRequest{
			ChunkId:  iur.ChunkId,
			Filename: iur.Filename,
			Length:   iur.Length,
		}
		buff.Image = nil
		return SUpload(buff)
	})
	// What to do when client send all data. And signal the end of the stream.
	builder.OnStreamEnd(func(iur *pb.ImageUploadRequest, i int) (*pb.ImageUploadResponse, error) {
		return &pb.ImageUploadResponse{Url: "182736127864"}, nil
	})
	// What to do after the request is done
	// Usually use for clean up session data or logging
	builder.OnDone(func() error {
		return (*s.EventRepository).DeleteTopic(topicId)
	})
	// What to do before return error
	builder.BeforeReturnError(func(iur *pb.ImageUploadRequest, err error) (*pb.ImageUploadRequest, error) {
		log.Printf("Error: %v", err)
		return iur, err
	})
	// What to do when there's an error
	// If not provided, default handler will be used instead.
	builder.OnError(func(iur *pb.ImageUploadRequest, err error) error {
		return err
	})

	// This will block the process until the stream is done
	// Beware of usage
	return controller.ClientStreamServer[*pb.ImageUploadRequest, *pb.ImageUploadResponse](
		stream,
		builder.Build(),
		bus, // Use nil if you don't want to broadcast the event
		controller.FunctionOptionsBuilder().InputValidation(false).OutputValidation(false),
	)
}

// Server Stream function
// Listen to the event bus and send the data to the client
// event is came from other process that publish the event to the event bus
func (s *LogService) Listen(reqData *pb.LogRequest, stream pb.LogStreamService_ListenServer) error {
	builder := controller.NewServerStreamBuilder[*pb.ImageUploadRequest, *pb.ImageUploadResponse, *pb.LogResponse]()
	// What to do when there're request from client within image upload service
	builder.OnReqData(func(iur *pb.ImageUploadRequest) (*pb.LogResponse, error) {
		data := &pb.ImageUploadRequest{
			ChunkId:  iur.ChunkId,
			Filename: iur.Filename,
			Length:   iur.Length,
		}
		result := &pb.LogResponse{Value: data.Filename}
		return result, nil
	})
	// What to do when there're response from server within image upload service
	builder.OnResData(func(iur *pb.ImageUploadResponse) (*pb.LogResponse, error) {
		result := &pb.LogResponse{Value: "Stream ended"}
		return result, nil
	})

	// Get topic bus that client needed
	// Autehntication is highly recommended before allow client to listen to the topic
	// Don't trust client input. Otherwise, it will be a huge security issue
	topic := reqData.Key
	bus, err := (*s.EventRepository).GetTopic(topic)
	if err != nil {
		return err
	}

	// This will block the process until the stream is done
	// Beware of usage
	return controller.ServerStreamServer[*pb.ImageUploadRequest, *pb.ImageUploadResponse, *pb.LogResponse](
		stream,
		bus,
		builder.Build(),
	)
}

func main() {
	lis, err := net.Listen("tcp", ":6001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// Depend on your need. You can create multiple event repository for different use case.
	// Highly recommended to use protobuf message as the event.
	// So it can be used for multiple services or even different services accross different remote server.
	// But, it's not mandatory. You can use any type of data as you want.

	// This event repository can be used by multiple services. Across difference process.
	e := eventbus.NewEventRepository[*pb.ImageUploadRequest, *pb.ImageUploadResponse](1024)

	pb.RegisterAssetServiceServer(s, &AssetService{})

	// When passing the event repository into the service, ONLY pass the pointer of the event repository.
	// Otherwise, it will cause an event repository duplication and result in malfunction event bus.
	pb.RegisterImageServiceServer(s, &ImageService{EventRepository: &e})
	pb.RegisterLogStreamServiceServer(s, &LogService{EventRepository: &e})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
