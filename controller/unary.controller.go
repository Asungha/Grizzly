package controller

import (
	"context"
	"log"

	utils "github.com/Asungha/Grizzly/utils"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type UnaryFunction[T protoreflect.ProtoMessage, V protoreflect.ProtoMessage] func(context.Context, T) (V, error)
type ActionWrapper[T protoreflect.ProtoMessage, V protoreflect.ProtoMessage] func(context.Context, T, UnaryFunction[T, V]) (V, error)
type UnaryErrorFunction[T protoreflect.ProtoMessage] func(context.Context, T, error) error
type UnaryErrorInterceptorFunction[T protoreflect.ProtoMessage] func(context.Context, T, error) (T, error)

type UnaryFunctions[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] struct {
	IngressHandler   UnaryFunction[Req, Res]
	ErrorHandler     UnaryErrorFunction[Req]
	ErrorInterceptor UnaryErrorInterceptorFunction[Req]
}

func (f *UnaryFunctions[Req, Res]) HandleError(ctx context.Context, data Req, err error) error {
	if err != nil {
		if f.ErrorInterceptor != nil {
			data, err = f.ErrorInterceptor(ctx, data, err)
		}
		if f.ErrorHandler != nil {
			return f.ErrorHandler(ctx, data, err)
		}
	}
	return nil
}

type UnaryBuilder[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] struct {
	res UnaryFunctions[Req, Res]
}

func NewUnaryBuilder[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage]() *UnaryBuilder[Req, Res] {
	return &UnaryBuilder[Req, Res]{}
}

func (b *UnaryBuilder[Req, Res]) OnData(f UnaryFunction[Req, Res]) *UnaryBuilder[Req, Res] {
	b.res.IngressHandler = f
	return b
}

func (b *UnaryBuilder[Req, Res]) OnError(f UnaryErrorFunction[Req]) *UnaryBuilder[Req, Res] {
	b.res.ErrorHandler = f
	return b
}

func (b *UnaryBuilder[Req, Res]) BeforeReturnError(f UnaryErrorInterceptorFunction[Req]) *UnaryBuilder[Req, Res] {
	b.res.ErrorInterceptor = f
	return b
}

func (b *UnaryBuilder[Req, Res]) Build() UnaryFunctions[Req, Res] {
	return b.res
}

func UnaryServer[T protoreflect.ProtoMessage, V protoreflect.ProtoMessage](ctx context.Context, data T, functions UnaryFunctions[T, V], option *FunctionOptions) (V, error) {
	var placeholder V
	if option == nil {
		option = FunctionOptionsBuilder().InputValidation(USE_INPUT_VALIDATION).OutputValidation(USE_OUTPUT_VALIDATION)
	}
	f := func(ctx context.Context, data T) (V, error) {
		validator, err := protovalidate.New()
		if err != nil {
			remainError := functions.HandleError(ctx, data, err)
			if remainError != nil {
				return placeholder, remainError
			} else {
				return placeholder, utils.ServiceError(err)
			}
		}

		if option.inputValidation {
			if err := validator.Validate(data); err != nil {
				remainError := functions.HandleError(ctx, data, err)
				if remainError != nil {
					return placeholder, remainError
				} else {
					return placeholder, utils.ServiceError(err)
				}
			}
		}

		res, err := functions.IngressHandler(ctx, data)
		if err != nil {
			remainError := functions.HandleError(ctx, data, err)
			if remainError != nil {
				return placeholder, remainError
			} else {
				return placeholder, utils.ServiceError(err)
			}
		}
		if option.outputValidation {
			log.Printf("Validating output: %v", res)
			if err := validator.Validate(res); err != nil {
				remainError := functions.HandleError(ctx, data, err)
				if remainError != nil {
					return placeholder, remainError
				} else {
					return placeholder, utils.DataLossError(err)
				}
			}
		}
		return res, nil
	}
	res, err := UnaryFunction[T, V](f)(ctx, data)
	return res, err
}
