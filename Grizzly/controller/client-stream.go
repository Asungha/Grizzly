package controller

import (
	. "grizzly/utils"
	"io"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	IngressHandler   ClientStreamFunction[Req]
	StreamEndHandler ClientStreamPostFunction[Req, Res]
	ErrorHandler     ClientStreamErrorFunction[Req]
	ErrorInterceptor ClientStreamErrorInterceptorFunction[Req]
}

type ClientStreamBuilder[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] struct {
	res ClientStreamFunctions[Req, Res]
}

func NewClientStreamBuilder[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage]() *ClientStreamBuilder[Req, Res] {
	return &ClientStreamBuilder[Req, Res]{}
}

func (b *ClientStreamBuilder[Req, Res]) OnData(f ClientStreamFunction[Req]) *ClientStreamBuilder[Req, Res] {
	b.res.IngressHandler = f
	return b
}

func (b *ClientStreamBuilder[Req, Res]) OnEnd(f ClientStreamPostFunction[Req, Res]) *ClientStreamBuilder[Req, Res] {
	b.res.StreamEndHandler = f
	return b
}

func (b *ClientStreamBuilder[Req, Res]) OnError(f ClientStreamErrorFunction[Req]) *ClientStreamBuilder[Req, Res] {
	b.res.ErrorHandler = f
	return b
}

func (b *ClientStreamBuilder[Req, Res]) BeforeReturnError(f ClientStreamErrorInterceptorFunction[Req]) *ClientStreamBuilder[Req, Res] {
	b.res.ErrorInterceptor = f
	return b
}

func (b *ClientStreamBuilder[Req, Res]) Build() ClientStreamFunctions[Req, Res] {
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
	option *FunctionOptions,
) error {
	var chunkCount int32 = 0
	validator, err := protovalidate.New()
	if err != nil {
		return InternalError(err)
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
				if remainError != nil {
					return ServiceError(err)
				} else {
					return remainError
				}
			}
		}

		lastChunk = buffer
		chunkCount++

		if option.inputValidation {
			if err := validator.Validate(buffer); err != nil {
				remainError := functions.HandleError(buffer, err)
				if remainError != nil {
					return DataLossError(err)
				} else {
					return remainError
				}
			}
		}
		err = functions.IngressHandler(buffer)
		if err != nil {
			remainError := functions.HandleError(buffer, err)
			if remainError != nil {
				return ServiceError(err)
			} else {
				return remainError
			}
		}
	}

	var res Res

	if functions.StreamEndHandler != nil {
		postProcessRes, err := functions.StreamEndHandler(lastChunk, int(chunkCount))
		if err != nil {
			remainError := functions.HandleError(lastChunk, err)
			if remainError != nil {
				return ServiceError(err)
			} else {
				return remainError
			}
		}
		res = postProcessRes
	}
	stream.SendAndClose(res)
	return nil
}