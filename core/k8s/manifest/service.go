package manifest

import (
	"context"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/event"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

type Service struct {
	bus          pubsub.Bus
	logger       *logrus.Logger
	providerAddr common.Address
}

func NewService(bus pubsub.Bus, logger *logrus.Logger, providerAddr common.Address) *Service {
	return &Service{
		bus:          bus,
		logger:       logger,
		providerAddr: providerAddr,
	}
}

func (s *Service) Submit(ctx context.Context, deploymentID dtypes.DeploymentID, sdlManifest sdl.SDL) error {
	if err := s.ValidateRequest(ctx, deploymentID, sdlManifest); err != nil {
		return err
	}

	timestamp := time.Now().Unix()

	version, err := sdlManifest.Version()
	if err != nil {
		return err
	}

	deployment := dtypes.Deployment{
		DeploymentID: deploymentID,
		State:        dtypes.DeploymentActive,
		Version:      version,
		CreatedAt:    timestamp,
	}

	manifest, err := sdlManifest.Manifest()
	if err != nil {
		return err
	}

	dgroups, err := sdlManifest.DeploymentGroups()
	if err != nil {
		return err
	}

	groups := make([]dtypes.Group, 0, len(dgroups))

	for idx, spec := range dgroups {
		groups = append(groups, dtypes.Group{
			GroupID:   dtypes.MakeGroupID(deployment.ID(), uint32(idx+1)),
			State:     dtypes.GroupOpen,
			GroupSpec: *spec,
			CreatedAt: timestamp,
		})
	}

	lease := event.LeaseWon{
		LeaseID: mtypes.LeaseID{
			Owner:    deploymentID.Owner,
			DSeq:     deploymentID.DSeq,
			Provider: strings.ToLower(s.providerAddr.Hex()),
		},
		Group: &groups[0],
	}

	deploymentResponse := dtypes.QueryDeploymentResponse{
		Deployment: deployment,
		Groups:     groups,
	}

	s.logger.Debug("publishing manifest received for lease", "lease_id", lease.LeaseID)
	if err := s.bus.Publish(event.ManifestReceived{
		LeaseID:    lease.LeaseID,
		Group:      lease.Group,
		Manifest:   &manifest,
		Deployment: &deploymentResponse,
	}); err != nil {
		s.logger.Error("publishing event", "err", err, "lease", lease.LeaseID)
	}

	return nil
}
