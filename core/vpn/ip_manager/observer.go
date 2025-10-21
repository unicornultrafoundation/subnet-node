package ipmanager

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type IPManagerObserver interface {
	SendIP(ip string)
	Close()
	GetChannel() <-chan string
}

type observer struct {
	ipCh   chan string
	id     string
	closed bool
	mu     sync.RWMutex
}

func NewObserver() IPManagerObserver {
	return &observer{
		ipCh:   make(chan string, 10), // Buffered channel to prevent blocking
		id:     uuid.New().String(),
		closed: false,
	}
}

func (o *observer) SendIP(ip string) {
	o.mu.RLock()
	if o.closed {
		o.mu.RUnlock()
		return
	}

	select {
	case o.ipCh <- ip:
	default:
		// Channel is full, skip this IP update
	}
	o.mu.RUnlock()
}

func (o *observer) Close() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.closed {
		close(o.ipCh)
		o.closed = true
	}
}

func (o *observer) GetChannel() <-chan string {
	return o.ipCh
}

func (i *IPManagerImpl) WatchIP() (IPManagerObserver, error) {
	i.mu.Lock()
	if i.stopped {
		i.mu.Unlock()
		return nil, fmt.Errorf("IP manager is stopped")
	}

	observer := NewObserver()
	i.observerList = append(i.observerList, observer)
	i.mu.Unlock()

	// Start observer goroutine with proper context handling
	go func() {
		for {
			select {
			case ip := <-i.ipCh:
				observer.SendIP(ip)
			case <-i.stopChan:
				return
			case <-time.After(30 * time.Second):
				// Periodic check to prevent goroutine leaks
				continue
			}
		}
	}()

	return observer, nil
}
