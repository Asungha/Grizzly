package eventbroker

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type IEventBroker[T protoreflect.ProtoMessage] interface {
	ImplEventBroker()

	Subscribe() (string, chan T, error)
	SubscribeWithID(string) (chan T, error)
	Listen(ctx context.Context, client string, f func(T) error) error

	Publish(event T) error

	Unsubscribe(client string)

	Destroy()
}
