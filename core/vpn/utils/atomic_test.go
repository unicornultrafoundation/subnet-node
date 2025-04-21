package utils

import (
	"testing"
)

func TestAtomicBool(t *testing.T) {
	// Test initial value
	b := NewAtomicBool(true)
	if !b.Get() {
		t.Errorf("Expected initial value to be true")
	}

	// Test Set
	b.Set(false)
	if b.Get() {
		t.Errorf("Expected value to be false after Set(false)")
	}

	// Test CompareAndSwap
	if b.CompareAndSwap(true, false) {
		t.Errorf("CompareAndSwap should return false when current value doesn't match old")
	}
	if !b.CompareAndSwap(false, true) {
		t.Errorf("CompareAndSwap should return true when current value matches old")
	}
	if !b.Get() {
		t.Errorf("Expected value to be true after CompareAndSwap(false, true)")
	}
}

func TestAtomicCounter(t *testing.T) {
	// Test initial value
	c := NewAtomicCounter(42)
	if c.Get() != 42 {
		t.Errorf("Expected initial value to be 42, got %d", c.Get())
	}

	// Test Set
	c.Set(100)
	if c.Get() != 100 {
		t.Errorf("Expected value to be 100 after Set(100), got %d", c.Get())
	}

	// Test Increment
	if c.Increment() != 101 {
		t.Errorf("Expected Increment to return 101, got %d", c.Get())
	}

	// Test Decrement
	if c.Decrement() != 100 {
		t.Errorf("Expected Decrement to return 100, got %d", c.Get())
	}

	// Test Add
	if c.Add(42) != 142 {
		t.Errorf("Expected Add(42) to return 142, got %d", c.Get())
	}

	// Test CompareAndSwap
	if c.CompareAndSwap(100, 200) {
		t.Errorf("CompareAndSwap should return false when current value doesn't match old")
	}
	if !c.CompareAndSwap(142, 200) {
		t.Errorf("CompareAndSwap should return true when current value matches old")
	}
	if c.Get() != 200 {
		t.Errorf("Expected value to be 200 after CompareAndSwap(142, 200), got %d", c.Get())
	}
}
