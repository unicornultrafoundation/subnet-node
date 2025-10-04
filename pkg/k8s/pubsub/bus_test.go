package pubsub_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
)

func newAddress() []byte {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	if err != nil {
		panic(err)
	}
	return crypto.PubkeyToAddress(privKey.PublicKey).Bytes()
}

func TestBus(t *testing.T) {
	bus := pubsub.NewBus()
	defer bus.Close()

	ev := newEvent(newAddress())

	assert.NoError(t, bus.Publish(ev))

	sub1, err := bus.Subscribe()
	require.NoError(t, err)

	sub2, err := bus.Subscribe()
	require.NoError(t, err)

	assert.NoError(t, bus.Publish(ev))

	select {
	case newEv := <-sub1.Events():
		assert.Equal(t, ev, newEv)
	case <-pubsub.AfterThreadStart(t):
		require.Fail(t, "time out")
	}

	select {
	case newEv := <-sub2.Events():
		assert.Equal(t, ev, newEv)
	case <-pubsub.AfterThreadStart(t):
		require.Fail(t, "time out")
	}

	sub2.Close()

	select {
	case <-sub2.Done():
	case <-pubsub.AfterThreadStart(t):
		require.Fail(t, "time out")
	}

	assert.NoError(t, bus.Publish(ev))

	select {
	case newEv := <-sub1.Events():
		assert.Equal(t, ev, newEv)
	case <-pubsub.AfterThreadStart(t):
		require.Fail(t, "time out")
	}

	select {
	case <-sub2.Events():
		require.Fail(t, "spurious event")
	case <-pubsub.AfterThreadStart(t):
	}

	bus.Close()

	select {
	case <-sub1.Done():
	case <-pubsub.AfterThreadStart(t):
		require.Fail(t, "time out")
	}

	assert.Equal(t, pubsub.ErrNotRunning, bus.Publish(ev))

}

func TestClone(t *testing.T) {
	bus := pubsub.NewBus()
	defer bus.Close()

	did1 := newAddress()
	ev1 := newEvent(did1)

	did2 := newAddress()
	ev2 := newEvent(did2)

	assert.NoError(t, bus.Publish(ev1))

	sub1, err := bus.Subscribe()
	require.NoError(t, err)

	select {
	case <-sub1.Events():
		require.Fail(t, "spurious event")
	case <-pubsub.AfterThreadStart(t):
	}

	assert.NoError(t, bus.Publish(ev1))
	assert.NoError(t, bus.Publish(ev2))

	// allow event propagation
	pubsub.SleepForThreadStart(t)

	// clone subscription
	sub2, err := sub1.Clone()
	require.NoError(t, err)

	// both subscriptions should receive both events

	for i, pev := range []pubsub.Event{ev1, ev2} {
		select {
		case ev := <-sub1.Events():
			assert.Equal(t, pev, ev, "sub1 event %v", i+1)
		case <-pubsub.AfterThreadStart(t):
			require.Fail(t, "timeout sub1 event %v", i+1)
		}

		select {
		case ev := <-sub2.Events():
			assert.Equal(t, pev, ev, "sub2 event %v", i+1)
		case <-pubsub.AfterThreadStart(t):
			require.Fail(t, "timeout sub2 event %v", i+1)
		}
	}

	// sub1 should close sub2
	sub1.Close()

	select {
	case <-sub2.Done():
	case <-pubsub.AfterThreadStart(t):
		require.Fail(t, "time out closing sub2")
	}

	select {
	case <-sub1.Done():
	case <-pubsub.AfterThreadStart(t):
		require.Fail(t, "time out closing sub1")
	}

}

type testEvent []byte

func newEvent(addr []byte) testEvent {
	return testEvent(addr)
}
