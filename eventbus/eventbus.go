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
			log.Printf("Received req event: %v", event)
			err := f(event)
			if err != nil {
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
			log.Printf("Received res event: %v", event)
			err := f(event)
			if err != nil {
				return err
			}
		}
	}
}

func (c *EventBus[Req, Res]) PublishRequestPipe(event Req) error {
	log.Printf("Current request pipe: %v", c.RequestPipe)
	if !c.config.AllowRequestPipeSubscription && c.RequestPipe != nil {
		log.Printf("Request pipe publishing is not allowed")
		return errors.New("request pipe publishing is not allowed")
	}
	err := c.RequestPipe.Publish(event)
	log.Printf("Published event: %v %v", event, err)
	return err
}

func (c *EventBus[Req, Res]) PublishResponsePipe(event Res) error {
	if !c.config.AllowResponsePipeSubscription && c.ResponsePipe != nil {
		return errors.New("response pipe publishing is not allowed")
	}
	err := c.ResponsePipe.Publish(event)
	log.Printf("Published event: %v %v", event, err)
	return err
}

func (c *EventBus[Req, Res]) UnsubscribeRequestPipe(client string) {
	c.RequestPipe.Unsubscribe(client)
}

func (c *EventBus[Req, Res]) UnsubscribeResponsePipe(client string) {
	c.ResponsePipe.Unsubscribe(client)
}

func NewEventBus[
	Req protoreflect.ProtoMessage,
	Res protoreflect.ProtoMessage,
](config EventBusConfig) IEventBus[Req, Res] {
	res := &EventBus[Req, Res]{}
	res.Init(config)
	return res
}
