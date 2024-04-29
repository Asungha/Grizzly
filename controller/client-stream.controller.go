package controller

import (
	"errors"
	"io"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	utils "github.com/Asungha/Grizzly/utils"

	eventbus "github.com/Asungha/Grizzly/eventbus"
)

type ClientStreamFunction[V protoreflect.ProtoMessage] func(V) error
type ClientStreamPostFunction[T, V protoreflect.ProtoMessage] func(T, int) (V, error)
type ClientStreamErrorFunction[V protoreflect.ProtoMessage] func(V, error) error
type ClientStreamErrorInterceptorFunction[V protoreflect.ProtoMessage] func(V, error) (V, error)

// type ClientStreamWrapper[V protoreflect.ProtoMessage] func(V, ClientStreamFunction[V]) error

type ClientStream[Recv protoreflect.ProtoMessage, Resp protoreflect.ProtoMessage] interface {
	SendAndClose(Resp) error
	Recv() (Recv, error)
	grpc.ServerStream
}

type ClientStreamFunctions[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] struct {
	ConnectHandler   func() error
	IngressHandler   ClientStreamFunction[Req]
	StreamEndHandler ClientStreamPostFunction[Req, Res]
	DoneHandler      func() error
	ErrorHandler     ClientStreamErrorFunction[Req]
	ErrorInterceptor ClientStreamErrorInterceptorFunction[Req]

	config eventbus.EventBusConfig
}

type ClientStreamHandler[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] struct {
	res ClientStreamFunctions[Req, Res]
}

func NewClientStreamHandler[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage]() *ClientStreamHandler[Req, Res] {
	return &ClientStreamHandler[Req, Res]{}
}

func (b *ClientStreamHandler[Req, Res]) OnConnect(f func() error) *ClientStreamHandler[Req, Res] {
	b.res.ConnectHandler = f
	return b
}

func (b *ClientStreamHandler[Req, Res]) OnData(f ClientStreamFunction[Req]) *ClientStreamHandler[Req, Res] {
	b.res.IngressHandler = f
	return b
}

func (b *ClientStreamHandler[Req, Res]) OnStreamEnd(f ClientStreamPostFunction[Req, Res]) *ClientStreamHandler[Req, Res] {
	b.res.StreamEndHandler = f
	return b
}

func (b *ClientStreamHandler[Req, Res]) OnDone(f func() error) *ClientStreamHandler[Req, Res] {
	b.res.DoneHandler = f
	return b
}

func (b *ClientStreamHandler[Req, Res]) OnError(f ClientStreamErrorFunction[Req]) *ClientStreamHandler[Req, Res] {
	b.res.ErrorHandler = f
	return b
}

func (b *ClientStreamHandler[Req, Res]) BeforeReturnError(f ClientStreamErrorInterceptorFunction[Req]) *ClientStreamHandler[Req, Res] {
	b.res.ErrorInterceptor = f
	return b
}

func (b *ClientStreamHandler[Req, Res]) Export() ClientStreamFunctions[Req, Res] {
	return b.res
}

func (f *ClientStreamFunctions[Req, Res]) HandleError(data Req, err error) error {
	if err != nil {
		if f.ErrorInterceptor != nil {
			data, err = f.ErrorInterceptor(data, err)
		}
		if f.ErrorHandler != nil {
			return f.ErrorHandler(data, err)
		}
	}
	return nil
}

func handleClientStream[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage](
	stream ClientStream[Req, Res],
	functions ClientStreamFunctions[Req, Res],
	eventbus *eventbus.IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage],
	option *FunctionOptions,
) (Res, error) {
	var chunkCount int32 = 0
	validator, err := protovalidate.New()
	if err != nil {
		return utils.GetZero[Res](), utils.InternalError(err)
	}
	var lastChunk Req
	if option == nil {
		option = FunctionOptionsBuilder().InputValidation(USE_INPUT_VALIDATION).OutputValidation(USE_OUTPUT_VALIDATION)
	}
	for {
		buffer, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				break
			} else {
				remainError := functions.HandleError(buffer, err)
				if remainError == nil {
					return utils.GetZero[Res](), utils.ServiceError(err)
				} else {
					return utils.GetZero[Res](), remainError
				}
			}
		}

		lastChunk = buffer
		chunkCount++

		if option.inputValidation {
			if err := validator.Validate(buffer); err != nil {
				return utils.GetZero[Res](), utils.InvalidArgumentError(err)
			}
		}
		if eventbus != nil {
			(*eventbus).PublishRequestPipe(buffer)
		}
		err = functions.IngressHandler(buffer)
		if err != nil {
			return utils.GetZero[Res](), utils.ServiceError(err)
		}
	}

	var res Res

	if functions.StreamEndHandler != nil {
		postProcessRes, err := functions.StreamEndHandler(lastChunk, int(chunkCount))
		if err != nil {
			return utils.GetZero[Res](), utils.ServiceError(err)
		}
		res = postProcessRes
		if eventbus != nil {
			(*eventbus).PublishResponsePipe(res)
		}
	} else {
		return utils.GetZero[Res](), utils.InternalError(errors.New("no stream end handler"))
	}
	if functions.DoneHandler != nil {
		defer functions.DoneHandler()
	}
	return res, nil
}

/*
Wrapper for stream functions

@param stream Stream[Recv, Resp] - GRPC stream object
@param f ClientStreamFunction[V] - Function to call for each chunk of data stream
@param pf ClientStreamPostFunction[T, V] - Function to call after all data stream is received
@param option *FunctionOptions - Options for the function
*/
func ClientStreamServer[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage](
	stream ClientStream[Req, Res],
	functions ClientStreamFunctions[Req, Res],
	eventbus *eventbus.IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage],
	option *FunctionOptions,
) error {
	if functions.ConnectHandler != nil {
		err := functions.ConnectHandler()
		if err != nil {
			return utils.ServiceError(err)
		}
	}
	res, err := handleClientStream(stream, functions, eventbus, option)
	if err != nil {
		return err
	}
	if err := stream.SendAndClose(res); err != nil {
		return utils.ServiceError(err)
	}
	return nil
}
