package eventbus

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect" // "google.golang.org/protobuf/relect/protoreflect"
)

type IEventBus[
	Req protoreflect.ProtoMessage,
	Res protoreflect.ProtoMessage,
] interface {
	ImplEventbus()
	Init(EventBusConfig)

	GetConfig() EventBusConfig

	SubscribeRequestPipe(client string)
	SubscribeResponsePipe(client string)

	ListenRequestPipe(ctx context.Context, client string, f func(Req) error) error
	ListenResponsePipe(ctx context.Context, client string, f func(Res) error) error

	PublishRequestPipe(event Req) error
	PublishResponsePipe(event Res) error

	UnsubscribeRequestPipe(client string)
	UnsubscribeResponsePipe(client string)

	Destroy()
}
