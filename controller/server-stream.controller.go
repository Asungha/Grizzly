package controller

import (
	"context"
	"log"
	"sync"

	eventbus "github.com/Asungha/Grizzly/eventbus"
	utils "github.com/Asungha/Grizzly/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ServerStreamOptions struct {
	ClientId string
}

type ServerStreamFunction[V protoreflect.ProtoMessage] func(V) error

type ServerStreamOnReq[Req, Res protoreflect.ProtoMessage] func(Req) (Res, error)

type ServerStreamOnRes[Req, Res protoreflect.ProtoMessage] func(Req) (Res, error)

type IServerStreamFunctions interface {
	ConnectHandler() func() error
	PipeReqHandler() ServerStreamOnReq[protoreflect.ProtoMessage, protoreflect.ProtoMessage]
	PipeResHandler() ServerStreamOnRes[protoreflect.ProtoMessage, protoreflect.ProtoMessage]

	SetConnectHandler(func() error)
	SetPipeReqHandler(ServerStreamOnReq[protoreflect.ProtoMessage, protoreflect.ProtoMessage])
	SetPipeResHandler(ServerStreamOnRes[protoreflect.ProtoMessage, protoreflect.ProtoMessage])
}
type ServerStreamFunctions[PipeReq, PipeRes, StreamRes protoreflect.ProtoMessage] struct {
	connectHandler func() error
	pipeReqHandler ServerStreamOnReq[protoreflect.ProtoMessage, protoreflect.ProtoMessage]
	pipeResHandler ServerStreamOnRes[protoreflect.ProtoMessage, protoreflect.ProtoMessage]
}

func (f ServerStreamFunctions[PipeReq, PipeRes, StreamRes]) ConnectHandler() func() error {
	return f.connectHandler
}

func (f ServerStreamFunctions[PipeReq, PipeRes, StreamRes]) PipeReqHandler() ServerStreamOnReq[protoreflect.ProtoMessage, protoreflect.ProtoMessage] {
	return f.pipeReqHandler
}

func (f ServerStreamFunctions[PipeReq, PipeRes, StreamRes]) PipeResHandler() ServerStreamOnRes[protoreflect.ProtoMessage, protoreflect.ProtoMessage] {
	return f.pipeResHandler
}

func (f *ServerStreamFunctions[PipeReq, PipeRes, StreamRes]) SetConnectHandler(connectHandler func() error) {
	f.connectHandler = connectHandler
}

func (f *ServerStreamFunctions[PipeReq, PipeRes, StreamRes]) SetPipeReqHandler(pipeReqHandler ServerStreamOnReq[protoreflect.ProtoMessage, protoreflect.ProtoMessage]) {
	f.pipeReqHandler = pipeReqHandler
}

func (f *ServerStreamFunctions[PipeReq, PipeRes, StreamRes]) SetPipeResHandler(pipeResHandler ServerStreamOnRes[protoreflect.ProtoMessage, protoreflect.ProtoMessage]) {
	f.pipeResHandler = pipeResHandler
}

type ServerStream interface {
	Send(protoreflect.ProtoMessage) error
	grpc.ServerStream
}

type TServerStream[Resp protoreflect.ProtoMessage] interface {
	Send(Resp) error
	grpc.ServerStream
}

type ServerStreamWrapper[Res protoreflect.ProtoMessage] struct {
	Stream TServerStream[Res]
	grpc.ServerStream
}

func (s *ServerStreamWrapper[Res]) Send(res protoreflect.ProtoMessage) error {
	// Check if the type of res is the same as the type of Stream
	if _, ok := res.(Res); !ok {
		log.Printf("Invalid argument")
		return status.Error(codes.InvalidArgument, "Invalid argument")
	}
	return (s.Stream).Send(res.(Res))
}

type ServerStreamHandler[
	PipeReq protoreflect.ProtoMessage,
	PipeRes protoreflect.ProtoMessage,
	StreamRes protoreflect.ProtoMessage,
] struct {
	res IServerStreamFunctions
}

func (b *ServerStreamHandler[PipeReq, PipeRes, StreamRes]) OnConnect(f func() error) *ServerStreamHandler[PipeReq, PipeRes, StreamRes] {
	b.res.SetConnectHandler(f)
	return b
}

func (b *ServerStreamHandler[PipeReq, PipeRes, StreamRes]) OnReqData(f ServerStreamOnReq[protoreflect.ProtoMessage, protoreflect.ProtoMessage]) *ServerStreamHandler[PipeReq, PipeRes, StreamRes] {
	b.res.SetPipeReqHandler(f)
	return b
}

func (b *ServerStreamHandler[PipeReq, PipeRes, StreamRes]) OnResData(f ServerStreamOnRes[protoreflect.ProtoMessage, protoreflect.ProtoMessage]) *ServerStreamHandler[PipeReq, PipeRes, StreamRes] {
	b.res.SetPipeResHandler(f)
	return b
}

func (b *ServerStreamHandler[PipeReq, PipeRes, StreamRes]) Export() IServerStreamFunctions {
	return b.res
}

func NewServerStreamHandler[
	PipeReq protoreflect.ProtoMessage,
	PipeRes protoreflect.ProtoMessage,
	StreamRes protoreflect.ProtoMessage,
]() *ServerStreamHandler[PipeReq, PipeRes, StreamRes] {
	return &ServerStreamHandler[PipeReq, PipeRes, StreamRes]{res: &ServerStreamFunctions[PipeReq, PipeRes, StreamRes]{}}
}

// func NewClientStreamBuilder[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage]() *ClientStreamBuilder[Req, Res] {
// 	return &ClientStreamBuilder[Req, Res]{}
// }

func handleServerStream(
	stream ServerStream,
	eventbus *eventbus.IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage],
	functions IServerStreamFunctions,
	option *ServerStreamOptions,
) error {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	ctx, cancle := context.WithCancel(context.Background())

	errChan := make(chan error, 1)

	var req func()
	var res func()

	if functions.ConnectHandler != nil && functions.ConnectHandler() != nil {
		err := functions.ConnectHandler()()
		if err != nil {
			return utils.ServiceError(err)
		}
	}

	if functions.PipeReqHandler != nil && functions.PipeReqHandler() != nil && (*eventbus).GetConfig().AllowRequestPipeSubscription {
		log.Println("PipeReqHandler is not nil")
		req = func() {
			defer wg.Done()
			(*eventbus).ListenRequestPipe(ctx, option.ClientId, func(iur protoreflect.ProtoMessage) error {
				result, err := functions.PipeReqHandler()(iur)
				if err != nil {
					errChan <- err
					cancle()
					return err
				}
				if err := stream.Send(result.(protoreflect.ProtoMessage)); err != nil {
					errChan <- err
					cancle()
					return err
				}
				return nil
			})
			cancle()
			log.Printf("Stop listening request pipe")
		}
	}

	if functions.PipeResHandler != nil && functions.PipeResHandler() != nil && (*eventbus).GetConfig().AllowResponsePipeSubscription {
		log.Println("PipeResHandler is not nil")
		res = func() {
			defer wg.Done()
			(*eventbus).ListenResponsePipe(ctx, option.ClientId, func(iur protoreflect.ProtoMessage) error {
				result, err := functions.PipeResHandler()(iur)
				if err != nil {
					errChan <- err
					cancle()
					return err
				}
				if err := stream.Send(result.(protoreflect.ProtoMessage)); err != nil {
					errChan <- err
					cancle()
					return err
				}
				return nil
			})
			log.Printf("Stop listening response pipe")
			cancle()
		}
	}

	if req != nil {
		wg.Add(1)
		go req()
	}
	if res != nil {
		wg.Add(1)
		go res()
	}

	done := make(chan bool)

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		return nil
	case err := <-errChan:
		cancle()
		return err
	}
}

func serverStreamHandler(binderWG *sync.WaitGroup, spec IServerStreamSpec) error {
	defer binderWG.Done()
	bus := spec.Eventbus()
	clientId := uuid.New().String()
	log.Printf("Subscribing client %s", clientId)
	functions := spec.Handler()
	if functions.ConnectHandler != nil && functions.ConnectHandler() != nil {
		err := functions.ConnectHandler()()
		if err != nil {
			return utils.ServiceError(err)
		}
	}
	err := handleServerStream(spec.Stream(), bus, functions, &ServerStreamOptions{ClientId: clientId}) // Use the casted functions
	log.Printf("Unsubscribing client %s", clientId)
	(*bus).UnsubscribeRequestPipe(clientId)
	(*bus).UnsubscribeResponsePipe(clientId)
	log.Printf("Client %s unsubscribed", clientId)
	if err != nil {
		log.Printf("Server stream ended with error")
		log.Printf(err.Error())
		return err
	}
	log.Printf("Server stream ended successfully")
	return status.Error(codes.Aborted, "Server stream ended")
}

// func BindServerStream[
// 	PipeReq protoreflect.ProtoMessage,
// 	PipeRes protoreflect.ProtoMessage,
// 	StreamRes protoreflect.ProtoMessage,
// ](
// 	spec ServerStreamSpec[PipeReq, PipeRes, StreamRes],
// ) error {
// 	return serverStreamHandler(spec)
// }

type IServerStreamSpec interface {
	IServerStreamSpecImpl()

	Handler() IServerStreamFunctions
	Eventbus() *eventbus.IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage]
	Stream() ServerStream

	SetHandler(IServerStreamFunctions)
	SetEventbus(*eventbus.IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage])
	SetStream(ServerStream)
}

type ServerStreamSpec[
	StreamRes protoreflect.ProtoMessage,
] struct {
	stream    ServerStream
	eventbus  *eventbus.IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage]
	functions IServerStreamFunctions
}

func (s *ServerStreamSpec[StreamRes]) IServerStreamSpecImpl() {}

func (s *ServerStreamSpec[StreamRes]) Handler() IServerStreamFunctions {
	return s.functions
}

func (s *ServerStreamSpec[StreamRes]) Eventbus() *eventbus.IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage] {
	return s.eventbus
}

func (s *ServerStreamSpec[StreamRes]) Stream() ServerStream {
	return s.stream
}

func (s *ServerStreamSpec[StreamRes]) SetHandler(functions IServerStreamFunctions) {
	s.functions = functions
}

func (s *ServerStreamSpec[StreamRes]) SetEventbus(eventbus *eventbus.IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage]) {
	s.eventbus = eventbus
}

func (s *ServerStreamSpec[StreamRes]) SetStream(stream ServerStream) {
	s.stream = stream
}

type ServerStreamBinder struct {
	wg sync.WaitGroup
}

func NewServerStreamBinder() *ServerStreamBinder {
	return &ServerStreamBinder{wg: sync.WaitGroup{}}
}

func (b *ServerStreamBinder) Bind(spec IServerStreamSpec) *ServerStreamBinder {
	b.wg.Add(1)
	go serverStreamHandler(&b.wg, spec)
	return b
}

func (b *ServerStreamBinder) Serve() {
	b.wg.Wait()
}
