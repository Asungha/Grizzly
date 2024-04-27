package eventbus

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type IEventBus[
	Req protoreflect.ProtoMessage,
	Res protoreflect.ProtoMessage,
] interface {
	ImplCotroller()
	Init(EventBusConfig)

	SubscribeRequestPipe(client string)
	SubscribeResponsePipe(client string)

	ListenRequestPipe(client string, f func(Req) error) error
	ListenResponsePipe(client string, f func(Res) error) error

	PublishRequestPipe(event Req) error
	PublishResponsePipe(event Res) error

	UnsubscribeRequestPipe(client string)
	UnsubscribeResponsePipe(client string)
}
