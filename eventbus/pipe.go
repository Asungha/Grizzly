package eventbus

import (
	"log"
)

type IPipe[T any] interface {
	IPipeimpl()
	Init()

	Publish(event T) error
	Subscribe(client string) chan T
	Unsubscribe(client string)
}

type InternalPipe[T any] struct {
	EventBuses map[string]chan T
}

func (s *InternalPipe[T]) IPipeimpl() {}

func (s *InternalPipe[T]) Init() {
	s.EventBuses = make(map[string]chan T, 10)
}

func NewInternalPipe[T any]() *InternalPipe[T] {
	res := &InternalPipe[T]{}
	res.Init()
	return res
}

func (s *InternalPipe[T]) Publish(event T) error {
	log.Printf("Prepare publishing event: %v", event)
	log.Printf("Current channels: %v", s.EventBuses)
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
		return s.EventBuses[client]
	}
	log.Printf("Creating a new channel for %s", client)
	s.EventBuses[client] = make(chan T)
	log.Printf("Current channels: %v", s.EventBuses)
	return s.EventBuses[client]
}

func (s *InternalPipe[T]) Unsubscribe(client string) {
	delete(s.EventBuses, client)
}

func (s *InternalPipe[T]) Listen(client string) chan T {
	return s.EventBuses[client]
}
