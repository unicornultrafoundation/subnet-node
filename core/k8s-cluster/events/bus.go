package events

import (
	"context"
	"fmt"
	"sync"
)

// EventHandler defines a function that handles an event
type EventHandler[T any] func(context.Context, T) error

// EventBus defines the interface for event bus with generic type support
type EventBus[T any] interface {
	// Subscribe subscribes to an event type
	Subscribe(eventType string, handler EventHandler[T]) error
	// Unsubscribe unsubscribes from an event type
	Unsubscribe(eventType string, handler EventHandler[T]) error
	// Publish publishes an event
	Publish(ctx context.Context, eventType string, data T) error
	// Start initializes the event bus
	Start(ctx context.Context) error
	// Stop cleans up the event bus
	Stop(ctx context.Context) error
}

// DefaultEventBus implements the EventBus interface with generic type support
type DefaultEventBus[T any] struct {
	handlers map[string][]EventHandler[T]
	mu       sync.RWMutex
	started  bool
	stopCh   chan struct{}
}

// NewEventBus creates a new event bus
func NewEventBus[T any]() EventBus[T] {
	return &DefaultEventBus[T]{
		handlers: make(map[string][]EventHandler[T]),
		stopCh:   make(chan struct{}),
	}
}

// Subscribe subscribes to an event type
func (b *DefaultEventBus[T]) Subscribe(eventType string, handler EventHandler[T]) error {
	if !b.started {
		return fmt.Errorf("event bus is not started")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	handlers := b.handlers[eventType]
	handlers = append(handlers, handler)
	b.handlers[eventType] = handlers
	return nil
}

// Unsubscribe unsubscribes from an event type
func (b *DefaultEventBus[T]) Unsubscribe(eventType string, handler EventHandler[T]) error {
	if !b.started {
		return fmt.Errorf("event bus is not started")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	handlers := b.handlers[eventType]
	for i, h := range handlers {
		if &h == &handler {
			handlers = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
	b.handlers[eventType] = handlers
	return nil
}

// Publish publishes an event
func (b *DefaultEventBus[T]) Publish(ctx context.Context, eventType string, data T) error {
	if !b.started {
		return fmt.Errorf("event bus is not started")
	}

	b.mu.RLock()
	handlers := b.handlers[eventType]
	b.mu.RUnlock()

	var errs []error
	for _, handler := range handlers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-b.stopCh:
			return fmt.Errorf("event bus is stopped")
		default:
			if err := handler(ctx, data); err != nil {
				errs = append(errs, fmt.Errorf("handler error: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors in event handlers: %v", errs)
	}
	return nil
}

// Start initializes the event bus
func (b *DefaultEventBus[T]) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.started {
		return fmt.Errorf("event bus is already started")
	}

	b.started = true
	return nil
}

// Stop cleans up the event bus
func (b *DefaultEventBus[T]) Stop(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.started {
		return fmt.Errorf("event bus is not started")
	}

	close(b.stopCh)
	b.started = false
	return nil
}
