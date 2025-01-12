package events

import (
	"sync"
)

type EventType string

const (
	// Raw events (from client/mud)
	EventRawInput  EventType = "raw_input"  // From client
	EventRawOutput EventType = "raw_output" // From MUD

	// Connection events
	EventConnect      EventType = "connect"      // Request to connect
	EventConnected    EventType = "connected"    // Connection established
	EventDisconnect   EventType = "disconnect"   // Request to disconnect
	EventDisconnected EventType = "disconnected" // Connection closed

	// Processed events (from LuaEngine)
	EventCommand      EventType = "command"
	EventOutput       EventType = "output"
	EventLog          EventType = "log"
	EventDebug        EventType = "debug"
	EventListBuffers  EventType = "list_buffers"
	EventSwitchBuffer EventType = "switch_buffer"

	// Client lifecycle events
	EventQuit EventType = "quit" // Request to quit the client
)

type Event struct {
	Type EventType
	Data interface{}
}

type Handler func(Event)

type EventProcessor struct {
	eventChan chan Event
	handlers  map[EventType][]Handler
	mu        sync.RWMutex
}

func New() *EventProcessor {
	ep := &EventProcessor{
		eventChan: make(chan Event, 1024),
		handlers:  make(map[EventType][]Handler),
	}
	// Start the dispatch loop
	go ep.run()
	return ep
}

func (ep *EventProcessor) run() {
	for event := range ep.eventChan {
		ep.mu.RLock()
		handlers := ep.handlers[event.Type]
		ep.mu.RUnlock()

		for _, h := range handlers {
			h(event)
		}
	}
}

func (ep *EventProcessor) Subscribe(eventType EventType, handler Handler) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.handlers[eventType] = append(ep.handlers[eventType], handler)
}

func (ep *EventProcessor) Emit(event Event) {
	ep.eventChan <- event
}
