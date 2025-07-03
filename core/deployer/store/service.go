package store

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
)

type Service struct {
	Datastore datastore.Datastore
	logger    *logrus.Logger
}

func NewService(ds datastore.Datastore, logger *logrus.Logger) *Service {
	return &Service{
		Datastore: ds,
		logger:    logger,
	}
}

func (s *Service) Start(ctx context.Context) error {
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}
