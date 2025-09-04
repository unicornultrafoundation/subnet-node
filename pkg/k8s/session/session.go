package session

import (
	"github.com/sirupsen/logrus"
	ptypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/v1"
)

// Session interface wraps Log, Client, Provider and ForModule methods
type Session interface {
	Log() *logrus.Logger
	Provider() *ptypes.Provider
	ForModule(string) Session
}

// New returns new session instance with provided details
func New(log *logrus.Logger, provider *ptypes.Provider) Session {
	return session{
		provider: provider,
		log:      log,
	}
}

type session struct {
	provider *ptypes.Provider
	log      *logrus.Logger
}

func (s session) Log() *logrus.Logger {
	return s.log
}

func (s session) Provider() *ptypes.Provider {
	return s.provider
}

func (s session) ForModule(name string) Session {
	s.log = s.log.WithField("module", name).Logger
	return s
}
