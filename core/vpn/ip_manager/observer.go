package ipmanager

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type IPManagerObserver interface {
	SendIP(ip string)
	Close()
	GetChannel() <-chan string
}

type observer struct {
	ipCh    chan string
	id      string
	closed  bool
	mu      sync.RWMutex
	onClose func(string)
}

func NewObserver(onClose func(string)) IPManagerObserver {
	return &observer{
		ipCh:    make(chan string, 10), // Buffered channel to prevent blocking
		id:      uuid.New().String(),
		closed:  false,
		onClose: onClose,
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
		if o.onClose != nil {
			o.onClose(o.id)
		}
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

	observer := NewObserver(func(id string) {
		// Remove the observer when it is closed
		i.removeObserver(id)
	})
	i.observerList = append(i.observerList, observer)
	i.mu.Unlock()

	return observer, nil
}
