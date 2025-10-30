package vbox_service

import (
	"context"
	"strings"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/config"
)

// MockDatastore implements the datastore.Datastore interface for testing
type MockDatastore struct {
	data map[datastore.Key][]byte
}

// NewMockDatastore creates a new mock datastoreD
func NewMockDatastore() *MockDatastore {
	return &MockDatastore{
		data: make(map[datastore.Key][]byte),
	}
}

// Put stores a value in the mock datastore
func (m *MockDatastore) Put(ctx context.Context, key datastore.Key, value []byte) error {
	m.data[key] = value
	return nil
}

// Get retrieves a value from the mock datastore
func (m *MockDatastore) Get(ctx context.Context, key datastore.Key) ([]byte, error) {
	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return nil, datastore.ErrNotFound
}

// Has checks if a key exists in the mock datastore
func (m *MockDatastore) Has(ctx context.Context, key datastore.Key) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

// Delete removes a key from the mock datastore
func (m *MockDatastore) Delete(ctx context.Context, key datastore.Key) error {
	delete(m.data, key)
	return nil
}

// Query performs a query on the mock datastore
func (m *MockDatastore) Query(ctx context.Context, q query.Query) (query.Results, error) {
	var results []query.Entry

	for key, value := range m.data {
		// Simple prefix matching
		if q.Prefix != "" && !strings.HasPrefix(key.String(), q.Prefix) {
			continue
		}

		results = append(results, query.Entry{
			Key:   key.String(),
			Value: value,
		})
	}

	return query.ResultsWithEntries(q, results), nil
}

// Sync ensures all data is written to storage
func (m *MockDatastore) Sync(ctx context.Context, prefix datastore.Key) error {
	return nil
}

// Close closes the mock datastore
func (m *MockDatastore) Close() error {
	return nil
}

// Batch returns a batch interface
func (m *MockDatastore) Batch(ctx context.Context) (datastore.Batch, error) {
	return &MockBatch{datastore: m}, nil
}

// GetSize returns the size of a value
func (m *MockDatastore) GetSize(ctx context.Context, key datastore.Key) (int, error) {
	if value, exists := m.data[key]; exists {
		return len(value), nil
	}
	return 0, datastore.ErrNotFound
}

// MockBatch implements the datastore.Batch interface
type MockBatch struct {
	datastore *MockDatastore
	ops       []batchOp
}

type batchOp struct {
	op    string
	key   datastore.Key
	value []byte
}

// Put adds a put operation to the batch
func (b *MockBatch) Put(ctx context.Context, key datastore.Key, value []byte) error {
	b.ops = append(b.ops, batchOp{"put", key, value})
	return nil
}

// Delete adds a delete operation to the batch
func (b *MockBatch) Delete(ctx context.Context, key datastore.Key) error {
	b.ops = append(b.ops, batchOp{"delete", key, nil})
	return nil
}

// Commit executes all operations in the batch
func (b *MockBatch) Commit(ctx context.Context) error {
	for _, op := range b.ops {
		switch op.op {
		case "put":
			b.datastore.data[op.key] = op.value
		case "delete":
			delete(b.datastore.data, op.key)
		}
	}
	return nil
}

// MockVBoxConfig creates a mock VBoxConfig for testing
func MockVBoxConfig() *config.VBoxConfig {
	return &config.VBoxConfig{
		BaseFolder: "~/VirtualBox VMs",
		ImagesDir:  "~/VirtualBox VMs/Images",
	}
}
