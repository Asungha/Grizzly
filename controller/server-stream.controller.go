package controller

import (
	"context"
	"reflect"
	"sync"

	eventbroker "github.com/Asungha/Grizzly/eventbroker"
	utils "github.com/Asungha/Grizzly/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ServerStreamOptions struct {
	ClientId string

	// Set as true if you want to listen to the data without sending a strem response
	SniffingOnly bool
}

type ServerStreamFunction[V protoreflect.ProtoMessage] func(V) error

type ServerStreamOnData[Req, Res protoreflect.ProtoMessage] func(Req) (Res, error)

type IServerStreamFunctions[I, O protoreflect.ProtoMessage] interface {
	ConnectHandler() func() error
	DataHandler() ServerStreamOnData[I, O]

	SetConnectHandler(func() error)
	SetDataHandler(ServerStreamOnData[I, O])
}
type ServerStreamFunctions[I, O protoreflect.ProtoMessage] struct {
	connectHandler func() error
	dataHandler    ServerStreamOnData[I, O]
}

func (f ServerStreamFunctions[I, O]) ConnectHandler() func() error {
	return f.connectHandler
}

func (f ServerStreamFunctions[I, O]) DataHandler() ServerStreamOnData[I, O] {
	return f.dataHandler
}

func (f *ServerStreamFunctions[I, O]) SetConnectHandler(connectHandler func() error) {
	f.connectHandler = connectHandler
}

func (f *ServerStreamFunctions[I, O]) SetDataHandler(dataHandler ServerStreamOnData[I, O]) {
	f.dataHandler = dataHandler
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
		return status.Error(codes.InvalidArgument, "Invalid argument")
	}
	return (s.Stream).Send(res.(Res))
}

// type ServerStreamHandler[
// 	StreamCap protoreflect.ProtoMessage,
// 	StreamRes protoreflect.ProtoMessage,
// ] struct {
// 	res IServerStreamFunctions[StreamCap, StreamRes]
// }

// func (b *ServerStreamHandler[StreamCap, StreamRes]) OnConnect(f func() error) *ServerStreamHandler[StreamCap, StreamRes] {
// 	b.res.SetConnectHandler(f)
// 	return b
// }

// func (b *ServerStreamHandler[StreamCap, StreamRes]) OnData(f ServerStreamOnData[StreamCap, StreamRes]) *ServerStreamHandler[StreamCap, StreamRes] {
// 	if casted, ok := f.(ServerStreamOnData[protoreflect.ProtoMessage, protoreflect.ProtoMessage]); !ok {
// 		log.Printf("Invalid argument")
// 		return b
// 	} else {
// 		b.res.SetDataHandler(casted)
// 	}
// }

// func (b *ServerStreamHandler[StreamCap, StreamRes]) Export() IServerStreamFunctions[StreamCap, StreamRes] {
// 	return b.res
// }

func NewServerStreamHandler[
	StreamCap protoreflect.ProtoMessage,
	StreamRes protoreflect.ProtoMessage,
]() IServerStreamFunctions[StreamCap, StreamRes] {
	return &ServerStreamFunctions[StreamCap, StreamRes]{}
}

// func NewClientStreamBuilder[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage]() *ClientStreamBuilder[Req, Res] {
// 	return &ClientStreamBuilder[Req, Res]{}
// }

func handleServerStream[
	StreamCap protoreflect.ProtoMessage,
	StreamRes protoreflect.ProtoMessage,
](
	ctx context.Context,
	stream ServerStream,
	broker eventbroker.IEventBroker[StreamCap],
	functions IServerStreamFunctions[StreamCap, StreamRes],
	option *ServerStreamOptions,
) error {
	if functions.ConnectHandler() != nil {
		err := functions.ConnectHandler()()
		if err != nil {
			return utils.ServiceError(err)
		}
	}

	var b *eventbroker.EventBroker[StreamCap]

	if v, ok := broker.(*eventbroker.EventBroker[StreamCap]); !ok || v == nil {
		return status.Error(codes.InvalidArgument, "No event broker")
	} else {
		b = v
	}

	if functions.DataHandler() != nil {
		err := b.Listen(ctx, option.ClientId, func(iur StreamCap) error {
			result, resErr := functions.DataHandler()(iur)
			if reflect.ValueOf(result).IsZero() {
				if resErr != nil {
					return resErr
				}
				return nil
			}
			if err := stream.Send(result); err != nil {
				return err
			}
			if resErr != nil {
				return resErr
			}
			return nil
		})
		if err != nil {
			return utils.ServiceError(err)
		}
	}
	return nil
}

// func BindServerStream[
// 	PipeReq protoreflect.ProtoMessage,
// 	PipeRes protoreflect.ProtoMessage,
// 	StreamRes protoreflect.ProtoMessage,
// ](
// 	spec ServerStreamTask[PipeReq, PipeRes, StreamRes],
// ) error {
// 	return serverStreamHandler(spec)
// }

type IServerStreamTask[StreamCap, StreamRes protoreflect.ProtoMessage] interface {
	IServerStreamTaskImpl()

	Handler() IServerStreamFunctions[StreamCap, StreamRes]
	EventBroker() eventbroker.IEventBroker[StreamCap]
	Stream() ServerStream

	OnlySniffing()
	IsOnlySniffing() bool

	SetHandler(IServerStreamFunctions[StreamCap, StreamRes])
	Seteventbroker(eventbroker.IEventBroker[StreamCap])
	// SetStream(ServerStream)
}

type ServerStreamTask[
	StreamCap protoreflect.ProtoMessage,
	StreamRes protoreflect.ProtoMessage,
] struct {
	stream      ServerStream
	eventbroker eventbroker.IEventBroker[StreamCap]
	functions   IServerStreamFunctions[StreamCap, StreamRes]

	onlySniffing bool
}

func (s *ServerStreamTask[StreamCap, StreamRes]) IServerStreamTaskImpl() {}

func (s *ServerStreamTask[StreamCap, StreamRes]) Handler() IServerStreamFunctions[StreamCap, StreamRes] {
	return s.functions
}

func (s *ServerStreamTask[StreamCap, StreamRes]) EventBroker() eventbroker.IEventBroker[StreamCap] {
	return s.eventbroker
}

func (s *ServerStreamTask[StreamCap, StreamRes]) Stream() ServerStream {
	return s.stream
}

func (s *ServerStreamTask[StreamCap, StreamRes]) SetHandler(functions IServerStreamFunctions[StreamCap, StreamRes]) {
	s.functions = functions
}

func (s *ServerStreamTask[StreamCap, StreamRes]) Seteventbroker(eventbroker eventbroker.IEventBroker[StreamCap]) {
	s.eventbroker = eventbroker
}

func (s *ServerStreamTask[StreamCap, StreamRes]) OnlySniffing() {
	s.onlySniffing = true
}

func (s *ServerStreamTask[StreamCap, StreamRes]) IsOnlySniffing() bool {
	return s.onlySniffing
}

// func (s *ServerStreamTask[StreamCap, StreamRes]) SetStream(stream ServerStream) {
// 	s.stream = stream
// }

type BindingOption struct {
	OnlySniffing bool
}

type ServerStreamBinder struct {
	wg     *sync.WaitGroup
	stream ServerStream
	Funcs  []func(*ServerStreamBinder, context.Context)
	errs   chan error
}

func NewServerStreamBinder(stream ServerStream) *ServerStreamBinder {
	return &ServerStreamBinder{wg: &sync.WaitGroup{}, stream: stream, errs: make(chan error)}
}

func Bind[StreamCap, StreamRes protoreflect.ProtoMessage](binder *ServerStreamBinder, task IServerStreamTask[StreamCap, StreamRes]) {
	t := ServerStreamTask[StreamCap, StreamRes]{
		eventbroker:  task.EventBroker(),
		functions:    task.Handler(),
		stream:       binder.stream,
		onlySniffing: task.IsOnlySniffing(),
	}
	binder.Funcs = append(binder.Funcs, func(b *ServerStreamBinder, ctx context.Context) {
		// err := serverStreamHandler[StreamCap, StreamRes](ctx, binder.wg, &t)
		// if err != nil {
		//	binder.errs <- err
		//	log.Printf("task Error %v", err)
		// }
		defer b.wg.Done()
		bus := t.EventBroker()
		clientId := uuid.New().String()
		functions := t.Handler()

		opts := &ServerStreamOptions{ClientId: clientId, SniffingOnly: false}

		err := handleServerStream(ctx, t.Stream(), bus, functions, opts) // Use the casted functions
		bus.Unsubscribe(clientId)
		bus.Unsubscribe(clientId)
		if err != nil {
			b.errs <- err
		}
	})
}

func (b *ServerStreamBinder) Serve() error {
	var done chan bool = make(chan bool)
	ctx, cancle := context.WithCancel(context.Background())
	for _, f := range b.Funcs {
		b.wg.Add(1)
		go f(b, ctx)
	}
	go func() {
		b.wg.Wait()
		done <- true
	}()

	// Wait for either an error or the stream to end
	select {
	case err := <-b.errs:
		cancle()
		return err
	case <-done:
		cancle()
		return status.Error(codes.OK, "Server stream ended successfully")
	}
}
