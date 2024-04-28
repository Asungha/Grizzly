package eventbus

import (
	"log"

	"google.golang.org/protobuf/reflect/protoreflect"

	"errors"
)

type EventBusConfig struct {
	AllowRequestPipeSubscription  bool
	AllowResponsePipeSubscription bool
}

type EventBus[
	Req protoreflect.ProtoMessage,
	Res protoreflect.ProtoMessage,
] struct {
	RequestPipe  IPipe[Req]
	ResponsePipe IPipe[Res]

	config EventBusConfig
}

func (c *EventBus[Req, Res]) ImplCotroller() {}

func (c *EventBus[Req, Res]) Init(config EventBusConfig) {
	c.config = config
	if config.AllowRequestPipeSubscription {
		c.RequestPipe = NewInternalPipe[Req]()
	}
	if config.AllowResponsePipeSubscription {
		c.ResponsePipe = NewInternalPipe[Res]()
	}
}

func (c *EventBus[Req, Res]) SubscribeRequestPipe(client string) {
	c.RequestPipe.Subscribe(client)
}

func (c *EventBus[Req, Res]) SubscribeResponsePipe(client string) {
	c.ResponsePipe.Subscribe(client)
}

func (c *EventBus[Req, Res]) ListenRequestPipe(client string, f func(Req) error) error {
	if c.RequestPipe == nil {
		return errors.New("no request pipe available")
	}
	channel := c.RequestPipe.Subscribe(client)
	for {
		select {
		case event := <-channel:
			// check if event is nil
			if !event.ProtoReflect().IsValid() {
				return errors.New("event is nil")
			}
			err := f(event)
			if err != nil {
				log.Printf("error in request pipe: %v", err)
				return err
			}
		}
	}
}

func (c *EventBus[Req, Res]) ListenResponsePipe(client string, f func(Res) error) error {
	if c.ResponsePipe == nil {
		return errors.New("no response pipe available")
	}
	channel := c.ResponsePipe.Subscribe(client)
	for {
		select {
		case event := <-channel:
			if !event.ProtoReflect().IsValid() {
				return errors.New("event is nil")
			}
			err := f(event)
			if err != nil {
				return err
			}
		}
	}
}

func (c *EventBus[Req, Res]) PublishRequestPipe(event Req) error {
	log.Printf("publishing request pipe %v", event)
	if !c.config.AllowRequestPipeSubscription || c.RequestPipe == nil {
		return errors.New("request pipe publishing is not allowed")
	}
	err := c.RequestPipe.Publish(event)
	return err
}

func (c *EventBus[Req, Res]) PublishResponsePipe(event Res) error {
	if !c.config.AllowResponsePipeSubscription || c.ResponsePipe == nil {
		return errors.New("response pipe publishing is not allowed")
	}
	err := c.ResponsePipe.Publish(event)
	return err
}

func (c *EventBus[Req, Res]) UnsubscribeRequestPipe(client string) {
	c.RequestPipe.Unsubscribe(client)
}

func (c *EventBus[Req, Res]) UnsubscribeResponsePipe(client string) {
	c.ResponsePipe.Unsubscribe(client)
}

func (c *EventBus[Req, Res]) Destroy() {
	if c.RequestPipe != nil {
		c.RequestPipe.CloseAll()
	}
	if c.ResponsePipe != nil {
		c.ResponsePipe.CloseAll()
	}
}

func NewEventBus[
	Req protoreflect.ProtoMessage,
	Res protoreflect.ProtoMessage,
](config EventBusConfig) *EventBus[Req, Res] {
	res := &EventBus[Req, Res]{}
	res.Init(config)
	return res
}
