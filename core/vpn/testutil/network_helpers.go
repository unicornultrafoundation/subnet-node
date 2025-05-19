package testutil

import (
	"bytes"
	"sync"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/mock"
)

// MockNetworkPair creates a pair of mock networks for testing
type MockNetworkPair struct {
	// Host1's streams to Host2
	Host1ToHost2 map[string]*MockStream
	// Host2's streams to Host1
	Host2ToHost1 map[string]*MockStream
	// Mutex for thread safety
	mu sync.Mutex
}

// NewMockNetworkPair creates a new mock network pair
func NewMockNetworkPair() *MockNetworkPair {
	return &MockNetworkPair{
		Host1ToHost2: make(map[string]*MockStream),
		Host2ToHost1: make(map[string]*MockStream),
	}
}

// CreateStreamPair creates a pair of streams between the hosts
func (n *MockNetworkPair) CreateStreamPair(config *MockStreamConfig, protocol string) (*MockStream, *MockStream) {
	n.mu.Lock()
	defer n.mu.Unlock()

	stream1, stream2 := MockStreamPair(config)

	n.Host1ToHost2[protocol] = stream1
	n.Host2ToHost1[protocol] = stream2

	return stream1, stream2
}

// GetStream gets a stream from the network
func (n *MockNetworkPair) GetStream(fromHost1 bool, protocol string) *MockStream {
	n.mu.Lock()
	defer n.mu.Unlock()

	if fromHost1 {
		return n.Host1ToHost2[protocol]
	}
	return n.Host2ToHost1[protocol]
}

// MockStreamPair creates a pair of connected mock streams
func MockStreamPair(config *MockStreamConfig) (*MockStream, *MockStream) {
	if config == nil {
		config = &MockStreamConfig{}
	}

	// Create buffers for the streams
	buffer1 := new(bytes.Buffer)
	buffer2 := new(bytes.Buffer)

	// Create the streams
	stream1 := NewMockStream(config)
	stream2 := NewMockStream(config)

	// Set up stream1 to write to buffer1 and read from buffer2
	stream1.On("Write", mock.Anything).Return(func(p []byte) int {
		n, _ := buffer1.Write(p)
		return n
	}, nil)
	stream1.On("Read", mock.Anything).Return(func(p []byte) int {
		n, _ := buffer2.Read(p)
		return n
	}, nil)

	// Set up stream2 to write to buffer2 and read from buffer1
	stream2.On("Write", mock.Anything).Return(func(p []byte) int {
		n, _ := buffer2.Write(p)
		return n
	}, nil)
	stream2.On("Read", mock.Anything).Return(func(p []byte) int {
		n, _ := buffer1.Read(p)
		return n
	}, nil)

	// Set up other methods
	stream1.On("Close").Return(nil)
	stream2.On("Close").Return(nil)
	stream1.On("Reset").Return(nil)
	stream2.On("Reset").Return(nil)
	stream1.On("CloseRead").Return(nil)
	stream2.On("CloseRead").Return(nil)
	stream1.On("CloseWrite").Return(nil)
	stream2.On("CloseWrite").Return(nil)
	stream1.On("ID").Return("mock-stream-1")
	stream2.On("ID").Return("mock-stream-2")
	stream1.On("Protocol").Return(protocol.ID("/mock/1.0.0"))
	stream2.On("Protocol").Return(protocol.ID("/mock/1.0.0"))

	return stream1, stream2
}
