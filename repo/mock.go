package repo

import "github.com/unicornultrafoundation/subnet-node/config"

// Mock is not thread-safe.
type Mock struct {
	C *config.C
	D Datastore
}

func (m *Mock) Config() *config.C {
	return m.C
}

func (m *Mock) Close() error { return m.D.Close() }

func (m *Mock) Datastore() Datastore { return m.D }

func (m *Mock) SwarmKey() ([]byte, error) {
	return nil, nil
}

func (m *Mock) Path() string {
	return ""
}
