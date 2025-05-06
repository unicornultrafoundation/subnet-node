package testutil

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/mock"
)

// AnyByteSlice is a matcher for any byte slice
var AnyByteSlice = func() interface{} {
	return mock.MatchedBy(func(b []byte) bool {
		return true
	})
}

// AnyContext is a matcher for any context
var AnyContext = func() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		return true
	})
}

// AnyPeerID is a matcher for any peer ID
var AnyPeerID = func() interface{} {
	return mock.MatchedBy(func(p peer.ID) bool {
		return true
	})
}
