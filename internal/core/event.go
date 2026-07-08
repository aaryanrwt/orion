package core

import (
	"sync"
)

// EventType represents the category of the bus message.
type EventType string

const (
	EventDeviceOnline  EventType = "device_online"
	EventDeviceOffline EventType = "device_offline"
)

// Event represents a decoupled message payload.
type Event struct {
	Type    EventType
	Payload interface{}
}

// EventBus implements a simple thread-safe publish-subscribe message broker.
type EventBus struct {
	mu     sync.RWMutex
	subs   map[EventType][]chan Event
	closed bool
}

// NewEventBus instantiates a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[EventType][]chan Event),
	}
}

// Subscribe registers a channel to receive events of a specific type.
func (eb *EventBus) Subscribe(t EventType) chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, 10)
	eb.subs[t] = append(eb.subs[t], ch)
	return ch
}

// Publish distributes an event to all active subscribers.
func (eb *EventBus) Publish(ev Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if eb.closed {
		return
	}

	for _, ch := range eb.subs[ev.Type] {
		select {
		case ch <- ev:
		default:
			// Non-blocking drop if channel buffer is full
		}
	}
}

// Close teardowns all event subscription channels.
func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return
	}
	eb.closed = true

	for _, list := range eb.subs {
		for _, ch := range list {
			close(ch)
		}
	}
	eb.subs = make(map[EventType][]chan Event)
}
