package events

import (
	"sync"
)

// EventSystem defines the interface for event handling
type EventSystem interface {
	Subscribe(eventType EventType, handler Handler)
	Emit(event Event)
}

type EventType string

const (
	// System events
	EventConnect      EventType = "connect"      // Request to connect
	EventConnected    EventType = "connected"    // Connection established
	EventDisconnect   EventType = "disconnect"   // Request to disconnect
	EventDisconnected EventType = "disconnected" // Connection closed
	EventQuit         EventType = "quit"         // Quit application
    
	// UI events
	EventRedraw       EventType = "redraw"       // Request UI redraw
	EventResize       EventType = "resize"       // Terminal resize
	EventScroll       EventType = "scroll"       // Scroll viewport
	EventOutput       EventType = "output"       // Output text to display
	EventInput        EventType = "input"        // Input text from user
    
	// Buffer management
	EventListBuffers  EventType = "list_buffers"
	EventSwitchBuffer EventType = "switch_buffer"

	// Command events
	EventCommand     EventType = "command"      // Command to be sent to the server
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

func (t EventType) String() string {
	return string(t)
}
