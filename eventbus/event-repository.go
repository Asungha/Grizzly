package eventbus

import (
	"context"
	"errors"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type IEventRepository[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] interface {
	EventRepositoryimpl()

	CreateTopic(topic string, config EventBusConfig) (*IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage], error)
	GetTopic(topic string) (*IEventBus[Req, Res], error)
	DeleteTopic(topic string) error

	Subscribe(topic string, client string) error
	ListenRequest(ctx context.Context, topic string, client string, f func(Req) error) error
	ListenResponse(ctx context.Context, topic string, client string, f func(Res) error) error

	PublishRequest(topic string, event Req) error
	PublishResponse(topic string, event Res) error
	Unsubscribe(topic string, client string) error
}

type EventRepository[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] struct {
	Topics map[string]*IEventBus[Req, Res] // Change the type to IEventBus[Req, Res] instead of *IEventBus[Req, Res]
}

func (r *EventRepository[Req, Res]) EventRepositoryimpl() {}

func (r *EventRepository[Req, Res]) CreateTopic(topic string, config EventBusConfig) (*IEventBus[Req, Res], error) {
	if r.Topics[topic] != nil {
		return nil, errors.New("topic already exists")
	}
	bus := NewEventBus[Req, Res](config)
	r.Topics[topic] = &bus
	return r.Topics[topic], nil
}

func (r *EventRepository[Req, Res]) GetTopic(topic string) (*IEventBus[Req, Res], error) {
	if r.Topics[topic] == nil {
		return nil, errors.New("topic not found")
	}
	return r.Topics[topic], nil
}

func (r *EventRepository[Req, Res]) DeleteTopic(topic string) error {
	if r.Topics[topic] == nil {
		return errors.New("topic not found when deleting")
	}

	// disconnect all clients
	(*r.Topics[topic]).Destroy()
	delete(r.Topics, topic)
	return nil
}

func (r *EventRepository[Req, Res]) Subscribe(topic string, client string) error {
	if r.Topics[topic] == nil {
		return errors.New("topic not found when subscribing")
	}
	(*r.Topics[topic]).SubscribeRequestPipe(client)
	return nil
}

func (r *EventRepository[Req, Res]) ListenRequest(ctx context.Context, topic string, client string, f func(Req) error) error {
	if r.Topics[topic] == nil {
		return errors.New("topic not found when listening request")
	}
	return (*r.Topics[topic]).ListenRequestPipe(ctx, client, f)
}

func (r *EventRepository[Req, Res]) ListenResponse(ctx context.Context, topic string, client string, f func(Res) error) error {
	if r.Topics[topic] == nil {
		return errors.New("topic not found when listening response")
	}
	return (*r.Topics[topic]).ListenResponsePipe(ctx, client, f)
}

func (r *EventRepository[Req, Res]) PublishRequest(topic string, event Req) error {
	if r.Topics[topic] == nil {
		return errors.New("topic not found when publishing request")
	}
	return (*r.Topics[topic]).PublishRequestPipe(event)
}

func (r *EventRepository[Req, Res]) PublishResponse(topic string, event Res) error {
	if r.Topics[topic] == nil {
		return errors.New("topic not found when publishing response")
	}
	return (*r.Topics[topic]).PublishResponsePipe(event)
}

func (r *EventRepository[Req, Res]) Unsubscribe(topic string, client string) error {
	if r.Topics[topic] == nil {
		return errors.New("topic not found when unsubscribing")
	}
	(*r.Topics[topic]).UnsubscribeRequestPipe(client)
	return nil
}

func NewEventRepository[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage](maxEvent int) IEventRepository[protoreflect.ProtoMessage, protoreflect.ProtoMessage] {
	return &EventRepository[protoreflect.ProtoMessage, protoreflect.ProtoMessage]{
		Topics: make(map[string]*IEventBus[protoreflect.ProtoMessage, protoreflect.ProtoMessage], maxEvent),
	}
}
