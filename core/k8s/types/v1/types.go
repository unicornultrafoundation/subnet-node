package v1

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/pkg/errors"
	inventoryV1 "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/inventory/v1"
	eventsv1 "k8s.io/api/events/v1"
)

type ProviderResourceEvent string

const (
	ProviderResourceAdd    = ProviderResourceEvent("add")
	ProviderResourceUpdate = ProviderResourceEvent("update")
	ProviderResourceDelete = ProviderResourceEvent("delete")
)

var (
	// ErrInsufficientCapacity is the new error when capacity is insufficient
	ErrInsufficientCapacity  = errors.New("insufficient capacity")
	ErrGroupResourceMismatch = errors.New("group resource mismatch")
)

// ServiceLog stores name, stream and scanner
type ServiceLog struct {
	Name    string
	Stream  io.ReadCloser
	Scanner *bufio.Scanner
}

type InventoryOptions struct {
	DryRun bool
}

type InventoryOption func(*InventoryOptions) *InventoryOptions

type Inventory interface {
	Adjust(ReservationGroup, ...InventoryOption) error
	Metrics() inventoryV1.Metrics
	Snapshot() inventoryV1.Cluster
	Dup() Inventory
}

type EventsWatcher interface {
	Shutdown()
	Done() <-chan struct{}
	ResultChan() <-chan *eventsv1.Event
	SendEvent(*eventsv1.Event) bool
}

type eventsFeed struct {
	ctx    context.Context
	cancel func()
	feed   chan *eventsv1.Event
}

var _ EventsWatcher = (*eventsFeed)(nil)

func NewEventsFeed(ctx context.Context) EventsWatcher {
	ctx, cancel := context.WithCancel(ctx)
	return &eventsFeed{
		ctx:    ctx,
		cancel: cancel,
		feed:   make(chan *eventsv1.Event),
	}
}

func (e *eventsFeed) Shutdown() {
	e.cancel()
}

func (e *eventsFeed) Done() <-chan struct{} {
	return e.ctx.Done()
}

func (e *eventsFeed) SendEvent(evt *eventsv1.Event) bool {
	select {
	case e.feed <- evt:
		return true
	case <-e.ctx.Done():
		return false
	}
}

func (e *eventsFeed) ResultChan() <-chan *eventsv1.Event {
	return e.feed
}

type ExecResult interface {
	ExitCode() int
}

// FilterGPUInterface ensures interface values are always lower case
// generalizes sxm* to sxm
func FilterGPUInterface(val string) string {
	val = strings.ToLower(val)

	if strings.HasPrefix(val, "sxm") {
		val = "sxm"
	}

	return val
}
