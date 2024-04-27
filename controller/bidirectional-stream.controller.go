package controller

import (
	"io"
	"log"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	utils "github.com/Asungha/Grizzly/utils"
)

type BidirectionalStreamFunction[V protoreflect.ProtoMessage] func(V) error
type BidirectionalPostFunction[T, V protoreflect.ProtoMessage] func(T, int) (V, error)
type BidirectionalErrorFunction[V protoreflect.ProtoMessage] func(V, error) error
type BidirectionalErrorInterceptorFunction[V protoreflect.ProtoMessage] func(V, error) (V, error)

// type BidirectionalWrapper[V protoreflect.ProtoMessage] func(V, BidirectionalFunction[V]) error

type BidirectionalStream[Recv protoreflect.ProtoMessage, Resp protoreflect.ProtoMessage] interface {
	SendAndClose(Resp) error
	Recv() (Recv, error)
	Send(Resp) error
	grpc.ServerStream
}

type BidirectionalFunctions[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] struct {
	DataHandler      BidirectionalStreamFunction[Req]
	StreamEndHandler BidirectionalPostFunction[Req, Res]
	ErrorHandler     BidirectionalErrorFunction[Req]
	ErrorInterceptor BidirectionalErrorInterceptorFunction[Req]
}

type BidirectionalBuilder[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] struct {
	res BidirectionalFunctions[Req, Res]
}

func NewBidirectionalBuilder[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage]() *BidirectionalBuilder[Req, Res] {
	return &BidirectionalBuilder[Req, Res]{}
}

func (b *BidirectionalBuilder[Req, Res]) OnData(f BidirectionalStreamFunction[Req]) *BidirectionalBuilder[Req, Res] {
	b.res.DataHandler = f
	return b
}

func (b *BidirectionalBuilder[Req, Res]) OnEnd(f BidirectionalPostFunction[Req, Res]) *BidirectionalBuilder[Req, Res] {
	b.res.StreamEndHandler = f
	return b
}

func (b *BidirectionalBuilder[Req, Res]) OnError(f BidirectionalErrorFunction[Req]) *BidirectionalBuilder[Req, Res] {
	b.res.ErrorHandler = f
	return b
}

func (b *BidirectionalBuilder[Req, Res]) BeforeReturnError(f BidirectionalErrorInterceptorFunction[Req]) *BidirectionalBuilder[Req, Res] {
	b.res.ErrorInterceptor = f
	return b
}

func (b *BidirectionalBuilder[Req, Res]) Build() BidirectionalFunctions[Req, Res] {
	return b.res
}

func (f *BidirectionalFunctions[Req, Res]) HandleError(data Req, err error) error {
	if err != nil {
		if f.ErrorInterceptor != nil {
			log.Println("Error interceptor called")
			data, err = f.ErrorInterceptor(data, err)
		}
		if f.ErrorHandler != nil {
			log.Println("Error handler called")
			return f.ErrorHandler(data, err)
		}
	}
	return nil
}

func bidirectionalhandle[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage](
	stream BidirectionalStream[Req, Res],
	functions BidirectionalFunctions[Req, Res],
	option *FunctionOptions,
) (*Res, error) {
	var chunkCount int32 = 0
	validator, err := protovalidate.New()
	if err != nil {
		return nil, utils.InternalError(err)
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
					return nil, utils.ServiceError(err)
				} else {
					return nil, remainError
				}
			}
		}

		lastChunk = buffer
		chunkCount++

		if option.inputValidation {
			if err := validator.Validate(buffer); err != nil {
				remainError := functions.HandleError(buffer, err)
				if remainError != nil {
					return nil, utils.DataLossError(err)
				} else {
					return nil, remainError
				}
			}
		}
		err = functions.DataHandler(buffer)
		if err != nil {
			remainError := functions.HandleError(buffer, err)
			if remainError != nil {
				return nil, utils.ServiceError(err)
			} else {
				return nil, remainError
			}
		}
	}

	var res Res

	if functions.StreamEndHandler != nil {
		postProcessRes, err := functions.StreamEndHandler(lastChunk, int(chunkCount))
		if err != nil {
			remainError := functions.HandleError(lastChunk, err)
			if remainError != nil {
				return nil, utils.ServiceError(err)
			} else {
				return nil, remainError
			}
		}
		res = postProcessRes
	}
	return &res, nil
}

/*
Wrapper for stream functions

@param stream Stream[Recv, Resp] - GRPC stream object
@param f BidirectionalFunction[V] - Function to call for each chunk of data stream
@param pf BidirectionalPostFunction[T, V] - Function to call after all data stream is received
@param option *FunctionOptions - Options for the function
*/
func BidirectionalServer[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage](
	stream BidirectionalStream[Req, Res],
	functions BidirectionalFunctions[Req, Res],
	option *FunctionOptions,
) error {
	res, err := bidirectionalhandle(stream, functions, option)
	if err != nil {
		return err
	}
	if err := stream.SendAndClose(*res); err != nil {
		return utils.ServiceError(err)
	}
	return nil
}
