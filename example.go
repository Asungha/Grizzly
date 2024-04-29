package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/Asungha/Grizzly/asset_stub"
	"github.com/Asungha/Grizzly/eventbus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	controller "github.com/Asungha/Grizzly/controller"
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
	EventRepository *eventbus.IEventRepository[protoreflect.ProtoMessage, protoreflect.ProtoMessage]
	AssetService    *AssetService
}

type LogService struct {
	pb.UnimplementedLogStreamServiceServer
	EventRepository *eventbus.IEventRepository[protoreflect.ProtoMessage, protoreflect.ProtoMessage]
}

// GRPC function impplementation example

// Unary function
func (s *AssetService) GetAsset(ctx context.Context, data *pb.GetAssetRequest) (*pb.GetAssetResponse, error) {
	builder := controller.NewUnaryBuilder[*pb.GetAssetRequest, *pb.GetAssetResponse]()
	// What to do when there's a request from client
	// Request data will be passed to the function anf response will be returned immediately
	builder.OnData(func(ctx context.Context, gar *pb.GetAssetRequest) (*pb.GetAssetResponse, error) {
		return &pb.GetAssetResponse{Asset: &pb.Asset{AssetDetail: &pb.AssetDetail{Area: 1.23}}}, nil
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
	// topicId := uuid.New().String()
	topicId := "image_upload"
	bus, _ := (*s.EventRepository).CreateTopic(
		topicId,
		eventbus.EventBusConfig{AllowRequestPipeSubscription: true, AllowResponsePipeSubscription: true},
	)

	// Create client stream builder
	builder := controller.NewClientStreamHandler[*pb.ImageUploadRequest, *pb.ImageUploadResponse]()
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
		log.Printf("Stream ended")
		asset, err := s.AssetService.GetAsset(context.Background(), &pb.GetAssetRequest{AssetId: "123456"})
		if err != nil {
			return nil, err
		}
		log.Printf("Asset: %v", asset)
		return &pb.ImageUploadResponse{Url: fmt.Sprintf("%f", asset.Asset.AssetDetail.Area)}, nil
	})
	// What to do after the request is done
	// Usually use for clean up session data or logging
	builder.OnDone(func() error {
		log.Printf("Delete topic: %v", topicId)
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
		builder.Export(),
		bus, // Use nil if you don't want to broadcast the event
		controller.FunctionOptionsBuilder().InputValidation(false).OutputValidation(false),
	)
}

// Server Stream function
// Listen to the event bus and send the data to the client
// event is came from other process that publish the event to the event bus
func (s *LogService) Listen(reqData *pb.LogRequest, stream pb.LogStreamService_ListenServer) error {
	// Get topic bus that client needed
	// Autehntication is highly recommended before allow client to listen to the topic
	// Don't trust client input. Otherwise, it will be a huge security issue
	topic := reqData.Key
	bus, err := (*s.EventRepository).GetTopic(topic)
	if err != nil {
		return err
	}

	// Defind the handler for the server stream

	handler := controller.NewServerStreamHandler[*pb.ImageUploadRequest, *pb.ImageUploadResponse, *pb.LogResponse]()
	// What to do when there're request from client within image upload service
	handler.OnReqData(func(iur protoreflect.ProtoMessage) (protoreflect.ProtoMessage, error) {
		// Always check the type of the data before using it
		data := iur.(*pb.ImageUploadRequest)
		result := &pb.LogResponse{Value: data.Filename}
		return result, nil
	})
	// What to do when there're response from server within image upload service
	handler.OnResData(func(iur protoreflect.ProtoMessage) (protoreflect.ProtoMessage, error) {
		result := &pb.LogResponse{Value: "Stream ended"}
		return result, nil
	})

	// Let's try using 2 handlers
	handler2 := controller.NewServerStreamHandler[*pb.ImageUploadRequest, *pb.ImageUploadResponse, *pb.LogResponse]()
	handler2.OnResData(func(iur protoreflect.ProtoMessage) (protoreflect.ProtoMessage, error) {
		result := &pb.LogResponse{Value: "Stream ended from 2"}
		return result, nil
	})

	// Create server stream specs
	// Each spec will be use to create goroutine to handle stream concurrently
	// You can have multiple spec to handle multiple event type in the same grpc request

	// It's possible to share data between specs via channel. But watch out for deadlock.
	// So, it's better to avoid sharing data between specs.

	// Event bus of each specs didn't have to be the same.
	// This allow you to have different event bus for different event. But using the same stream connection.
	spec := controller.ServerStreamSpec[*pb.LogResponse]{}
	spec.SetEventbus(bus)
	spec.SetHandler(handler.Export()) // First handler here
	spec.SetStream(&controller.ServerStreamWrapper[*pb.LogResponse]{Stream: stream})

	spec2 := controller.ServerStreamSpec[*pb.LogResponse]{}
	spec2.SetEventbus(bus)
	spec2.SetHandler(handler2.Export()) // Another handler here
	spec2.SetStream(&controller.ServerStreamWrapper[*pb.LogResponse]{Stream: stream})

	// This will block the process until the stream is done
	// Beware of usage
	binder := controller.NewServerStreamBinder()

	// Bind the spec to the binder
	// both spec will be run concurrently. And send the data to the client via the same stream connection
	binder.Bind(&spec).Bind(&spec2).Serve()
	return nil
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
	// In the future, This will be able to change transport to Redis or Kafka without changing the service code.
	// To support distributed system.
	e := eventbus.NewEventRepository[*pb.ImageUploadRequest, *pb.ImageUploadResponse](1024)

	assetService := &AssetService{}

	// To use other service within the service, you can pass the service as a parameter.
	// Make sure it's a pointer

	// When passing the event repository into the service, ONLY pass the pointer of the event repository.
	// Otherwise, it will cause an event repository duplication and result in malfunction event bus.

	// You can put grpc client in it if you want.
	// This will allow you to call function from other server.
	imageService := &ImageService{EventRepository: &e, AssetService: assetService}
	logService := &LogService{EventRepository: &e}

	// Register the service to the grpc server
	pb.RegisterAssetServiceServer(s, assetService)
	pb.RegisterImageServiceServer(s, imageService)
	pb.RegisterLogStreamServiceServer(s, logService)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
