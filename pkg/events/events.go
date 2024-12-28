package events

import (
	"sync"
)

type EventType string

const (
	// Raw events (from client/mud)
	EventRawInput  EventType = "raw_input"  // From client
	EventRawOutput EventType = "raw_output" // From MUD

	// Processed events (from LuaEngine)
	EventCommand      EventType = "command"
	EventConnect      EventType = "connect"
	EventDisconnect   EventType = "disconnect"
	EventOutput       EventType = "output"
	EventLog          EventType = "log"
	EventDebug        EventType = "debug"
	EventListBuffers  EventType = "list_buffers"
	EventSwitchBuffer EventType = "switch_buffer"
)

type Event struct {
	Type EventType
	Data interface{}
}

type Handler func(Event)

type EventProcessor struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler
}

func New() *EventProcessor {
	return &EventProcessor{
		handlers: make(map[EventType][]Handler),
	}
}

func (ep *EventProcessor) Subscribe(eventType EventType, handler Handler) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.handlers[eventType] = append(ep.handlers[eventType], handler)
}

func (ep *EventProcessor) Emit(event Event) {
	ep.mu.RLock()
	handlers := ep.handlers[event.Type]
	ep.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}
