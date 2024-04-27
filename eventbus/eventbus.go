package eventbus

type EventBus[T any] struct {
	subscribers map[string][]chan<- T
}

func NewEventBus[T any]() *EventBus[T] {
	return &EventBus[T]{
		subscribers: make(map[string][]chan<- T),
	}
}

func (eb *EventBus[T]) Subscribe(eventType string, subscriber chan<- T) {
	eb.subscribers[eventType] = append(eb.subscribers[eventType], subscriber)
}

func (eb *EventBus[T]) Publish(eventType string, event T) {
	subscribers := eb.subscribers[eventType]
	for _, subscriber := range subscribers {
		subscriber <- event
	}
}

type EventRepository[T any] struct {
	eventBus map[string]*EventBus[T]
}

func (er *EventRepository[T]) GetEventBus(eventType string) *EventBus[T] {
	return er.eventBus[eventType]
}

func (er *EventRepository[T]) AddEventBus(eventType string) {
	er.eventBus[eventType] = NewEventBus[T]()
}

func NewEventRepository[T any]() *EventRepository[T] {
	return &EventRepository[T]{
		eventBus: make(map[string]*EventBus[T]),
	}
}
