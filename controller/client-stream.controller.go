package controller

import (
	"errors"
	"io"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	utils "github.com/Asungha/Grizzly/utils"

	eventbroker "github.com/Asungha/Grizzly/eventbroker"
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

	config eventbroker.EventBrokerConfig
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

func (f *ClientStreamFunctions[Req, Res]) HandleError(data Req, err error, defaultError error) error {
	if f.ErrorInterceptor != nil {
		data, err = f.ErrorInterceptor(data, err)
	}
	if f.ErrorHandler != nil {
		return f.ErrorHandler(data, err)
	}
	return defaultError
}

func handleClientStream[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage](
	stream ClientStream[Req, Res],
	functions ClientStreamFunctions[Req, Res],
	onReqBus *eventbroker.IEventBroker[Req],
	onResBus *eventbroker.IEventBroker[Res],
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
				return utils.GetZero[Res](), functions.HandleError(lastChunk, err, utils.ServiceError(err))
			}
		}

		lastChunk = buffer
		chunkCount++

		if option.inputValidation {
			if err := validator.Validate(buffer); err != nil {
				return utils.GetZero[Res](), functions.HandleError(buffer, err, utils.InvalidArgumentError(err))
			}
		}
		if onReqBus != nil {
			(*onReqBus).Publish(buffer)
		}
		err = functions.IngressHandler(buffer)
		if err != nil {
			return utils.GetZero[Res](), functions.HandleError(buffer, err, utils.ServiceError(err))
		}
	}

	var res Res

	if functions.StreamEndHandler != nil {
		postProcessRes, err := functions.StreamEndHandler(lastChunk, int(chunkCount))
		if err != nil {
			return utils.GetZero[Res](), utils.ServiceError(err)
		}
		res = postProcessRes
		if onResBus != nil {
			(*onResBus).Publish(res)
		}
	} else {
		return utils.GetZero[Res](), utils.InternalError(errors.New("no stream end handler"))
	}
	if option.outputValidation {
		if err := validator.Validate(res); err != nil {
			return utils.GetZero[Res](), utils.DataLossError(err)
		}
	}
	if functions.DoneHandler != nil {
		err := functions.DoneHandler()
		if err != nil {
			return utils.GetZero[Res](), utils.ServiceError(err)
		}
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
	onReqBus *eventbroker.IEventBroker[Req],
	onResBus *eventbroker.IEventBroker[Res],
	option *FunctionOptions,
) error {
	if functions.ConnectHandler != nil {
		err := functions.ConnectHandler()
		if err != nil {
			return utils.ServiceError(err)
		}
	}
	res, err := handleClientStream(stream, functions, onReqBus, onResBus, option)
	if err != nil {
		return err
	}
	if option.outputValidation {
		validator, err := protovalidate.New()
		if err != nil {
			return utils.InternalError(err)
		}
		if err := validator.Validate(res); err != nil {
			return utils.DataLossError(err)
		}
	}
	if err := stream.SendAndClose(res); err != nil {
		return utils.ServiceError(err)
	}
	return nil
}
