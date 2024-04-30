package eventbroker

import (
	"errors"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func mapContainsKey[T any](m map[string]T, key string) bool {
	_, ok := m[key]
	return ok
}

type IEventRepository[Req, Res protoreflect.ProtoMessage] interface {
	EventRepositoryimpl()

	CreateTopic(topic string, config EventBrokerConfig) (IEventBroker[Req], IEventBroker[Res], error)
	GetTopic(topic string) (IEventBroker[Req], IEventBroker[Res], error)
	DeleteTopic(topic string) error
}

type EventRepository[Req protoreflect.ProtoMessage, Res protoreflect.ProtoMessage] struct {
	ReqTopics map[string]*EventBroker[Req]
	ResTopics map[string]*EventBroker[Res]
}

func (r *EventRepository[Req, Res]) EventRepositoryimpl() {}

func (r *EventRepository[Req, Res]) CreateTopic(topic string, config EventBrokerConfig) (IEventBroker[Req], IEventBroker[Res], error) {
	if mapContainsKey(r.ReqTopics, topic) || mapContainsKey(r.ResTopics, topic) {
		return nil, nil, errors.New("topic already exists")
	}
	reqbus := NewEventBroker[Req]()
	resbus := NewEventBroker[Res]()
	if casted, ok := reqbus.(*EventBroker[Req]); config.PublishOnRequest && ok {
		r.ReqTopics[topic] = casted
	}
	if casted, ok := resbus.(*EventBroker[Res]); config.PublishOnResponse && ok {
		r.ResTopics[topic] = casted
	}
	// log.Printf("topics %v %v", r.ReqTopics[topic], r.ResTopics[topic])
	return r.ReqTopics[topic], r.ResTopics[topic], nil
}

func (r *EventRepository[Req, Res]) GetTopic(topic string) (IEventBroker[Req], IEventBroker[Res], error) {
	return r.ReqTopics[topic], r.ResTopics[topic], nil
}

func (r *EventRepository[Req, Res]) DeleteTopic(topic string) error {
	if r.ReqTopics != nil && mapContainsKey(r.ReqTopics, topic) {
		r.ReqTopics[topic].Destroy()
		delete(r.ReqTopics, topic)
	}
	if r.ResTopics != nil && mapContainsKey(r.ResTopics, topic) {
		r.ResTopics[topic].Destroy()
		delete(r.ResTopics, topic)
	}
	return nil
}

func NewEventRepository[Req, Res protoreflect.ProtoMessage](maxEvent int) IEventRepository[Req, Res] {
	return &EventRepository[Req, Res]{
		ReqTopics: make(map[string]*EventBroker[Req], maxEvent),
		ResTopics: make(map[string]*EventBroker[Res], maxEvent),
	}
}
