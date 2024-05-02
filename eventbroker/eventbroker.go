package eventbroker

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/protobuf/reflect/protoreflect"

	"errors"
)

type EventBrokerConfig struct {
	// Whether to publish request data into the broker
	PublishOnRequest bool

	// Whether to publish response data after the request is done into the broker
	PublishOnResponse bool
}

type EventBroker[
	Data protoreflect.ProtoMessage,
] struct {
	channels map[string]chan Data
	unsub    map[string]bool
	unsubMux sync.Mutex
}

func (c *EventBroker[Data]) ImplEventBroker() {}

func (c *EventBroker[Data]) Subscribe() (string, chan Data, error) {
	if c == nil {
		return "", nil, errors.New("no channels available")
	}
	if c.channels == nil {
		return "", nil, errors.New("no channels available")
	}
	clientId := uuid.New().String()
	c.channels[clientId] = make(chan Data)
	return clientId, c.channels[clientId], nil
}

func (c *EventBroker[Data]) SubscribeWithID(clientId string) (chan Data, error) {
	if c == nil {
		return nil, errors.New("no channels available")
	}
	if c.channels == nil {
		return nil, errors.New("no channels available")
	}
	if c.channels[clientId] == nil {
		c.channels[clientId] = make(chan Data)
	}
	return c.channels[clientId], nil
}
func (c *EventBroker[Data]) Listen(ctx context.Context, client string, f func(Data) error) error {
	if c == nil {
		return errors.New("no channels available")
	}
	channels, err := c.SubscribeWithID(client)
	if err != nil {
		return err
	}
	for {
		select {
		case event := <-channels: // might cause crash
			if fmt.Sprintf("%v", event) == "<nil>" {
				return errors.New("nil event in response pipe")
			}
			err := f(event)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *EventBroker[Data]) Publish(event Data) error {
	if c == nil {
		return errors.New("no channels available")
	}
	for client, channels := range c.channels {
		//check if channel is closed
		c.unsubMux.Lock()
		if isDeleted, ok := c.unsub[client]; ok && isDeleted {
			close(channels)
			delete(c.channels, client)
			c.unsubMux.Unlock()
			continue
		} else {
			channels <- event
		}
		c.unsubMux.Unlock()
	}
	return nil
}

func (c *EventBroker[Data]) Unsubscribe(client string) {
	c.unsubMux.Lock()
	defer c.unsubMux.Unlock()
	if c == nil {
		return
	}
	if c.channels == nil {
		return
	}
	if c.channels[client] == nil {
		return
	}
	// c.channels[client] = nil
	c.unsub[client] = true

}

func (c *EventBroker[Data]) Destroy() {
	if c == nil {
		return
	}
	if c.channels != nil {
		for client := range c.channels {
			if c.channels[client] == nil {
				continue
			}
			close(c.channels[client])
			delete(c.channels, client)
		}
	}
}

func NewEventBroker[
	Data protoreflect.ProtoMessage,
]() IEventBroker[Data] {
	res := &EventBroker[Data]{
		channels: make(map[string]chan Data),
		unsub:    make(map[string]bool),
		unsubMux: sync.Mutex{},
	}
	return res
}
