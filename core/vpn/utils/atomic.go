package utils

import (
	"sync/atomic"
)

// AtomicBool is a boolean value that can be updated atomically
type AtomicBool struct {
	value atomic.Bool
}

// NewAtomicBool creates a new atomic boolean with the given initial value
func NewAtomicBool(initialValue bool) *AtomicBool {
	b := &AtomicBool{}
	b.value.Store(initialValue)
	return b
}

// Get returns the current value
func (b *AtomicBool) Get() bool {
	return b.value.Load()
}

// Set sets the value
func (b *AtomicBool) Set(value bool) {
	b.value.Store(value)
}

// CompareAndSwap sets the value to new only if the current value equals old
func (b *AtomicBool) CompareAndSwap(old, new bool) bool {
	return b.value.CompareAndSwap(old, new)
}

// AtomicCounter is a counter that can be updated atomically
type AtomicCounter struct {
	value atomic.Int64
}

// NewAtomicCounter creates a new atomic counter with the given initial value
func NewAtomicCounter(initialValue int64) *AtomicCounter {
	c := &AtomicCounter{}
	c.value.Store(initialValue)
	return c
}

// Get returns the current value
func (c *AtomicCounter) Get() int64 {
	return c.value.Load()
}

// Set sets the value
func (c *AtomicCounter) Set(value int64) {
	c.value.Store(value)
}

// Increment increments the counter and returns the new value
func (c *AtomicCounter) Increment() int64 {
	return c.value.Add(1)
}

// Decrement decrements the counter and returns the new value
func (c *AtomicCounter) Decrement() int64 {
	return c.value.Add(-1)
}

// Add adds the given value to the counter and returns the new value
func (c *AtomicCounter) Add(delta int64) int64 {
	return c.value.Add(delta)
}

// CompareAndSwap sets the value to new only if the current value equals old
func (c *AtomicCounter) CompareAndSwap(old, new int64) bool {
	return c.value.CompareAndSwap(old, new)
}
