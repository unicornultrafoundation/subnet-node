package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/steve/pkg/server/router"
	"github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/config"
	"k8s.io/client-go/rest"
)

type Router struct {
	scaled     *config.Scaled
	restConfig *rest.Config
	options    config.Options
}

func NewRouter(scaled *config.Scaled, restConfig *rest.Config, options config.Options) (*Router, error) {
	return &Router{
		scaled:     scaled,
		restConfig: restConfig,
		options:    options,
	}, nil
}

func (r *Router) Routes(h router.Handlers) http.Handler {
	m := mux.NewRouter()

	// TODO: Add routes

	return m
}
