package eventbus

import (
	"log"
)

type IPipe[T any] interface {
	IPipeimpl()
	Init(...any)

	Publish(key string, event T) error
	Subscribe(key string, event T) *chan T
}

type InternalPipe[T any] struct {
	EventBuses map[string]chan T
}

func (s *InternalPipe[T]) IPipeimpl() {}

func (s *InternalPipe[T]) Init(event string) {
	s.EventBuses = make(map[string]chan T, 10)
}

func (s *InternalPipe[T]) NewInternalPipe() *InternalPipe[T] {
	return &InternalPipe[T]{}
}

func (s *InternalPipe[T]) Publish(event T) error {
	for _, eventBus := range s.EventBuses {
		log.Printf("Publishing event to %v", eventBus)
		select {
		case eventBus <- event:
		default:
		}
	}
	return nil
}

func (s *InternalPipe[T]) Subscribe(client string) chan T {
	if s.EventBuses[client] != nil {
		log.Printf("Channel already exists for %s", client)
		// return s.EventBuses[client]
		return nil
	}
	log.Printf("Creating a new channel for %s", client)
	s.EventBuses[client] = make(chan T)
	return s.EventBuses[client]
}

func (s *InternalPipe[T]) Unsubscribe(client string) {
	delete(s.EventBuses, client)
}

func (s *InternalPipe[T]) Listen(client string) chan T {
	return s.EventBuses[client]
}

type RedisConfig struct {
	Host string
	Port int
}

type RedisPipe[T any] struct{}

func (s *RedisPipe[T]) IPipeimpl() {}

func (s *RedisPipe[T]) Init(config RedisConfig) {

}

func (s *RedisPipe[T]) Send(event T) error {
	return nil
}

func (s *RedisPipe[T]) Channel() *chan T {
	return nil
}
