package eventbus

import (
	"context"
	"fmt"
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

func (c *EventBus[Req, Res]) ImplEventbus() {}

func (c *EventBus[Req, Res]) Init(config EventBusConfig) {
	log.Printf("config %v", config)
	c.config = config
	if config.AllowRequestPipeSubscription == true {
		log.Printf("allowing request pipe subscription")
		c.RequestPipe = NewInternalPipe[Req]()
	}
	if config.AllowResponsePipeSubscription == true {
		log.Printf("allowing response pipe subscription")
		c.ResponsePipe = NewInternalPipe[Res]()
	}
}

func (c *EventBus[Req, Res]) SubscribeRequestPipe(client string) {
	c.RequestPipe.Subscribe(client)
}

func (c *EventBus[Req, Res]) SubscribeResponsePipe(client string) {
	c.ResponsePipe.Subscribe(client)
}

func (c *EventBus[Req, Res]) ListenRequestPipe(ctx context.Context, client string, f func(Req) error) error {
	if c.RequestPipe == nil {
		return errors.New("no request pipe available")
	}
	channel := c.RequestPipe.Subscribe(client)
	for {
		select {
		case event := <-channel: // might cause crash
			if fmt.Sprintf("%v", event) == "<nil>" {
				log.Printf("nil event in request pipe")
				return errors.New("nil event in response pipe")
			}
			err := f(event)
			if err != nil {
				log.Printf("error in request pipe: %v", err)
				return err
			}
		case sig := <-c.RequestPipe.GetSignal():
			if sig == PIPE_DESTROY {
				log.Printf("Req pipe destroyed")
				return errors.New("pipe destroyed")
			}
		case <-ctx.Done():
			log.Printf("Req: context done")
			return nil
		}
	}
}

func (c *EventBus[Req, Res]) ListenResponsePipe(ctx context.Context, client string, f func(Res) error) error {
	if c.ResponsePipe == nil {
		return errors.New("no response pipe available")
	}
	channel := c.ResponsePipe.Subscribe(client)
	for {
		select {
		case event := <-channel: // might cause crash
			// Don't know why I have to do this
			if fmt.Sprintf("%v", event) == "<nil>" {
				log.Printf("nil event in request pipe")
				return errors.New("nil event in response pipe")
			}
			err := f(event)
			if err != nil {
				return err
			}
		case sig := <-c.ResponsePipe.GetSignal():
			if sig == PIPE_DESTROY {
				log.Printf("Res pipe destroyed")
				return errors.New("pipe destroyed")
			}
		case <-ctx.Done():
			log.Printf("Res: context done")
			return nil
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
	if c.RequestPipe == nil {
		return
	}
	c.RequestPipe.Unsubscribe(client)
}

func (c *EventBus[Req, Res]) UnsubscribeResponsePipe(client string) {
	if c.ResponsePipe == nil {
		return
	}
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

func (c *EventBus[Req, Res]) GetConfig() EventBusConfig {
	return c.config
}

func NewEventBus[
	Req protoreflect.ProtoMessage,
	Res protoreflect.ProtoMessage,
](config EventBusConfig) IEventBus[Req, Res] {
	res := &EventBus[Req, Res]{} // Change to return a pointer to res
	res.Init(config)
	return res
}
