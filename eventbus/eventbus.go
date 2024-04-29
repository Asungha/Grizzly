package eventbus

import (
	"context"
	"fmt"

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

func (c *EventBus[Req, Res]) ListenRequestPipe(ctx context.Context, client string, f func(Req) error) error {
	if c.RequestPipe == nil {
		return errors.New("no request pipe available")
	}
	channel := c.RequestPipe.Subscribe(client)
	for {
		select {
		case event := <-channel: // might cause crash
			if fmt.Sprintf("%v", event) == "<nil>" {
				return errors.New("nil event in response pipe")
			}
			err := f(event)
			if err != nil {
				return err
			}
		case sig := <-c.RequestPipe.GetSignal():
			if sig == PIPE_DESTROY {
				return errors.New("pipe destroyed")
			}
		case <-ctx.Done():
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
				return errors.New("nil event in response pipe")
			}
			err := f(event)
			if err != nil {
				return err
			}
		case sig := <-c.ResponsePipe.GetSignal():
			if sig == PIPE_DESTROY {
				return errors.New("pipe destroyed")
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *EventBus[Req, Res]) PublishRequestPipe(event Req) error {
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
