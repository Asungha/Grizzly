package eventbus

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
	for _, eventBus := range s.EventBuses {
		select {
		case eventBus <- event:
		default:
		}
	}
	return nil
}

func (s *InternalPipe[T]) Subscribe(client string) chan T {
	if s.EventBuses[client] != nil {
		return s.EventBuses[client]
	}
	s.EventBuses[client] = make(chan T)
	return s.EventBuses[client]
}

func (s *InternalPipe[T]) Unsubscribe(client string) {
	delete(s.EventBuses, client)
}

func (s *InternalPipe[T]) Listen(client string) chan T {
	return s.EventBuses[client]
}
