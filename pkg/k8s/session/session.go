package session

import (
	"github.com/sirupsen/logrus"
	aclient "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/client"
	ptypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/v1"
)

// Session interface wraps Log, Client, Provider and ForModule methods
type Session interface {
	Log() *logrus.Logger
	Client() aclient.Client
	Provider() *ptypes.Provider
	ForModule(string) Session
	CreatedAtBlockHeight() int64
}

// New returns new session instance with provided details
func New(log *logrus.Logger, client aclient.Client, provider *ptypes.Provider, createdAtBlockHeight int64) Session {
	return session{
		client:               client,
		provider:             provider,
		log:                  log,
		createdAtBlockHeight: createdAtBlockHeight,
	}
}

type session struct {
	client               aclient.Client
	provider             *ptypes.Provider
	log                  *logrus.Logger
	createdAtBlockHeight int64
}

func (s session) Log() *logrus.Logger {
	return s.log
}

func (s session) Client() aclient.Client {
	return s.client
}

func (s session) Provider() *ptypes.Provider {
	return s.provider
}

func (s session) ForModule(name string) Session {
	s.log = s.log.WithField("module", name).Logger
	return s
}

func (s session) CreatedAtBlockHeight() int64 {
	return s.createdAtBlockHeight
}
