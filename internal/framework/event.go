package thinkgo

import (
	"fmt"
	"reflect"
	"sync"
)

// EventSystem provides ThinkPHP-style event handling.
// Listeners can be registered for events, and events can be triggered.
type EventSystem struct {
	mu        sync.RWMutex
	listeners map[string][]any
}

// NewEventSystem creates a new event system.
func NewEventSystem() *EventSystem {
	return &EventSystem{
		listeners: make(map[string][]any),
	}
}

// Listen registers a listener for the given event.
// listener can be a function or an EventListener interface.
func (e *EventSystem) Listen(event string, listener any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners[event] = append(e.listeners[event], listener)
}

// Trigger dispatches an event to all registered listeners.
// Returns any error from listeners.
func (e *EventSystem) Trigger(event string, args ...any) error {
	e.mu.RLock()
	listeners := e.listeners[event]
	e.mu.RUnlock()

	for _, listener := range listeners {
		if err := e.callListener(listener, args...); err != nil {
			return err
		}
	}
	return nil
}

// HasListeners checks if an event has listeners.
func (e *EventSystem) HasListeners(event string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.listeners[event]
	return ok && len(e.listeners[event]) > 0
}

// RemoveListeners removes all listeners for an event.
func (e *EventSystem) RemoveListeners(event string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.listeners, event)
}

// callListener invokes a listener with arguments.
func (e *EventSystem) callListener(listener any, args ...any) error {
	fn := reflect.ValueOf(listener)

	// If it's a function, call it
	if fn.Kind() == reflect.Func {
		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			in[i] = reflect.ValueOf(arg)
		}

		results := fn.Call(in)
		if len(results) > 0 {
			if err, ok := results[len(results)-1].Interface().(error); ok {
				return err
			}
		}
		return nil
	}

	// If it implements EventListener, call Handle
	if l, ok := listener.(EventListener); ok {
		return l.Handle(args...)
	}

	return fmt.Errorf("event: invalid listener type %T", listener)
}

// EventListener interface for class-based listeners.
type EventListener interface {
	Handle(args ...any) error
}

// Event is a base event struct.
type Event struct {
	Name string
	Data any
}

// NewEvent creates a new event.
func NewEvent(name string, data any) Event {
	return Event{Name: name, Data: data}
}
