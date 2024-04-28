package controller

import (
	"sync"

	eventbus "github.com/Asungha/Grizzly/eventbus"
	utils "github.com/Asungha/Grizzly/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ServerStreamOptions struct {
	ClientId string
}

type ServerStreamFunction[V protoreflect.ProtoMessage] func(V) error

type ServerStreamOnReq[Req, Res protoreflect.ProtoMessage] func(Req) (Res, error)

type ServerStreamOnRes[Req, Res protoreflect.ProtoMessage] func(Req) (Res, error)

type ServerStreamFunctions[PipeReq, PipeRes, StreamRes protoreflect.ProtoMessage] struct {
	ConnectHandler func() error
	PipeReqHandler ServerStreamOnReq[PipeReq, StreamRes]
	PipeResHandler ServerStreamOnRes[PipeRes, StreamRes]
}
type ServerStream[Resp protoreflect.ProtoMessage] interface {
	Send(Resp) error
	grpc.ServerStream
}

type ServerStreamBuilder[
	PipeReq protoreflect.ProtoMessage,
	PipeRes protoreflect.ProtoMessage,
	StreamRes protoreflect.ProtoMessage,
] struct {
	res ServerStreamFunctions[PipeReq, PipeRes, StreamRes]
}

func (b *ServerStreamBuilder[PipeReq, PipeRes, StreamRes]) OnConnect(f func() error) *ServerStreamBuilder[PipeReq, PipeRes, StreamRes] {
	b.res.ConnectHandler = f
	return b
}

func (b *ServerStreamBuilder[PipeReq, PipeRes, StreamRes]) OnReqData(f ServerStreamOnReq[PipeReq, StreamRes]) *ServerStreamBuilder[PipeReq, PipeRes, StreamRes] {
	b.res.PipeReqHandler = f
	return b
}

func (b *ServerStreamBuilder[PipeReq, PipeRes, StreamRes]) OnResData(f ServerStreamOnRes[PipeRes, StreamRes]) *ServerStreamBuilder[PipeReq, PipeRes, StreamRes] {
	b.res.PipeResHandler = f
	return b
}

func (b *ServerStreamBuilder[PipeReq, PipeRes, StreamRes]) Build() ServerStreamFunctions[PipeReq, PipeRes, StreamRes] {
	return b.res
}

func NewServerStreamBuilder[
	PipeReq protoreflect.ProtoMessage,
	PipeRes protoreflect.ProtoMessage,
	StreamRes protoreflect.ProtoMessage,
]() *ServerStreamBuilder[PipeReq, PipeRes, StreamRes] {
	return &ServerStreamBuilder[PipeReq, PipeRes, StreamRes]{res: ServerStreamFunctions[PipeReq, PipeRes, StreamRes]{}}
}

// func NewClientStreamBuilder[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage]() *ClientStreamBuilder[Req, Res] {
// 	return &ClientStreamBuilder[Req, Res]{}
// }

func handleServerStream[
	PipeReq protoreflect.ProtoMessage,
	PipeRes protoreflect.ProtoMessage,
	StreamRes protoreflect.ProtoMessage,
](
	stream ServerStream[StreamRes],
	eventbus *eventbus.EventBus[PipeReq, PipeRes],
	functions ServerStreamFunctions[PipeReq, PipeRes, StreamRes],
	option *ServerStreamOptions,
) error {
	var wg sync.WaitGroup

	if functions.ConnectHandler != nil {
		err := functions.ConnectHandler()
		if err != nil {
			return utils.ServiceError(err)
		}
	}

	if functions.PipeReqHandler != nil {
		wg.Add(1)
		req := func() {
			defer wg.Done()
			(*eventbus).ListenRequestPipe(option.ClientId, func(iur PipeReq) error {
				result, err := functions.PipeReqHandler(iur)
				if err != nil {
					return err
				}
				if err := stream.Send(result); err != nil {
					return err
				}
				return nil
			})
		}
		go req()
	}

	if functions.PipeResHandler != nil {
		wg.Add(1)
		res := func() {
			defer wg.Done()
			(*eventbus).ListenResponsePipe(option.ClientId, func(iur PipeRes) error {
				result, err := functions.PipeResHandler(iur)
				if err != nil {
					return err
				}
				if err := stream.Send(result); err != nil {
					return err
				}
				return nil
			})
		}
		go res()
	}
	wg.Wait()
	return nil
}

func ServerStreamServer[
	PipeReq protoreflect.ProtoMessage,
	PipeRes protoreflect.ProtoMessage,
	StreamRes protoreflect.ProtoMessage,
](
	stream ServerStream[StreamRes],
	eventbus *eventbus.EventBus[PipeReq, PipeRes],
	functions ServerStreamFunctions[PipeReq, PipeRes, StreamRes],
) error {
	clientId := uuid.New().String()
	if functions.ConnectHandler != nil {
		err := functions.ConnectHandler()
		if err != nil {
			return utils.ServiceError(err)
		}
	}
	err := handleServerStream(stream, eventbus, functions, &ServerStreamOptions{ClientId: clientId})
	eventbus.UnsubscribeRequestPipe(clientId)
	eventbus.UnsubscribeResponsePipe(clientId)
	if err != nil {
		return err
	}
	return nil
}
