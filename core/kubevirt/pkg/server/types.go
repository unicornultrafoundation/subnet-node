package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rancher/lasso/pkg/controller"
	steveauth "github.com/rancher/steve/pkg/auth"
	steveserver "github.com/rancher/steve/pkg/server"
	"github.com/rancher/wrangler/v3/pkg/generic"
	"github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/api/vm"
	"github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/config"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubevirtServer struct {
	Context    context.Context
	RESTConfig *rest.Config
	steve      *steveserver.Server
	scaled     *config.Scaled
	Handler    http.Handler
}

func New(ctx context.Context, clientConfig clientcmd.ClientConfig, options config.Options) (*KubevirtServer, error) {

	var err error

	server := &KubevirtServer{
		Context: ctx,
	}

	server.RESTConfig, err = clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	if err := server.generateSteveServer(options); err != nil {
		return nil, err
	}

	server.Handler = server.steve.Handler

	return server, nil
}

func (s *KubevirtServer) generateSteveServer(options config.Options) error {
	factory, err := controller.NewSharedControllerFactoryFromConfig(s.RESTConfig, Scheme)
	if err != nil {
		return err
	}

	factoryOpts := &generic.FactoryOptions{
		SharedControllerFactory: factory,
	}

	var scaled *config.Scaled

	s.Context, scaled, err = config.SetupScaled(s.Context, s.RESTConfig, factoryOpts)
	if err != nil {
		return err
	}

	router, err := NewRouter(scaled, s.RESTConfig, options)
	if err != nil {
		return err
	}

	s.scaled = scaled

	s.steve, err = steveserver.New(s.Context, s.RESTConfig, &steveserver.Options{
		Router: router.Routes,
	})

	if err != nil {
		return err
	}

	authMiddleware := steveauth.ToMiddleware(steveauth.AuthenticatorFunc(steveauth.AlwaysAdmin))
	s.Handler = authMiddleware(s.steve)

	vm.RegisterSchema(s.scaled, s.steve, options)

	return nil
}

func (s *KubevirtServer) StartControllers() error {

	scaled := s.scaled

	fmt.Println("scaled controllers", scaled)

	if err := scaled.VirtFactory.Start(s.Context, 1); err != nil {
		return fmt.Errorf("failed to start virt controllers: %v", err)
	}

	s.steve.StartAggregation(s.Context)
	return nil
}
