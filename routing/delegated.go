package routing

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	drclient "github.com/ipfs/boxo/routing/http/client"
	"github.com/ipfs/boxo/routing/http/contentrouter"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p-kad-dht/fullrt"
	record "github.com/libp2p/go-libp2p-record"
	routinghelpers "github.com/libp2p/go-libp2p-routing-helpers"
	ic "github.com/libp2p/go-libp2p/core/crypto"
	host "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	ma "github.com/multiformats/go-multiaddr"
	version "github.com/unicornultrafoundation/subnet-node"
	"github.com/unicornultrafoundation/subnet-node/config"
	"go.opencensus.io/stats/view"
)

var log = logging.Logger("routing/delegated")

// Strings is a helper type that (un)marshals a single string to/from a single
// JSON string and a slice of strings to/from a JSON array of strings.
type Strings []string

// Addresses stores the (string) multiaddr addresses for the node.
type Addresses struct {
	Swarm          []string // addresses for the swarm to listen on
	Announce       []string // swarm addresses to announce to the network, if len > 0 replaces auto detected addresses
	AppendAnnounce []string // similar to Announce but doesn't overwrite auto detected addresses, they are just appended
	NoAnnounce     []string // swarm addresses not to announce to the network
	API            Strings  // address for the local API (RPC)
	Gateway        Strings  // address to listen on for IPFS HTTP object gateway
}

// convertInterfaceToAddresses converts an interface{} map to Addresses struct
func ConvertConfigToAddresses(cfg *config.C) (Addresses, error) {
	addressMap, ok := cfg.Get("addresses").(map[interface{}]interface{})
	if !ok {
		return Addresses{}, fmt.Errorf("input is not a map[interface{}]interface{}")
	}

	// Helper function to safely convert a field to a []string
	convertToStringSlice := func(v interface{}) ([]string, error) {
		if v == nil {
			return nil, nil
		}
		rawSlice, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected []interface{} but got %T", v)
		}
		result := make([]string, len(rawSlice))
		for i, elem := range rawSlice {
			str, ok := elem.(string)
			if !ok {
				return nil, fmt.Errorf("expected string but got %T at index %d", elem, i)
			}
			result[i] = str
		}
		return result, nil
	}

	var addresses Addresses

	var err error
	// Convert each field with type assertions
	if swarm, exists := addressMap["Swarm"]; exists {
		addresses.Swarm, err = convertToStringSlice(swarm)
		if err != nil {
			return Addresses{}, err
		}
	}
	if announce, exists := addressMap["Announce"]; exists {
		addresses.Announce, err = convertToStringSlice(announce)
		if err != nil {
			return Addresses{}, err
		}
	}
	if appendAnnounce, exists := addressMap["AppendAnnounce"]; exists {
		addresses.AppendAnnounce, err = convertToStringSlice(appendAnnounce)
		if err != nil {
			return Addresses{}, err
		}
	}
	if noAnnounce, exists := addressMap["NoAnnounce"]; exists {
		addresses.NoAnnounce, err = convertToStringSlice(noAnnounce)
		if err != nil {
			return Addresses{}, err
		}
	}
	if api, exists := addressMap["API"]; exists {
		addresses.API, err = convertToStringSlice(api)
		if err != nil {
			return Addresses{}, err
		}
	}
	if gateway, exists := addressMap["Gateway"]; exists {
		addresses.Gateway, err = convertToStringSlice(gateway)
		if err != nil {
			return Addresses{}, err
		}
	}

	return addresses, nil
}

type Router struct {
	// Router type ID. See RouterType for more info.
	Type RouterType

	// Parameters are extra configuration that this router might need.
	// A common one for HTTP router is "Endpoint".
	Parameters interface{}
}

type (
	Routers map[string]Router
	Methods map[MethodName]Method
)

func (m Methods) Check() error {
	// Check supported methods
	for _, mn := range MethodNameList {
		_, ok := m[mn]
		if !ok {
			return fmt.Errorf("method name %q is missing from Routing.Methods config param", mn)
		}
	}

	// Check unsupported methods
	for k := range m {
		seen := false
		for _, mn := range MethodNameList {
			if mn == k {
				seen = true
				break
			}
		}

		if seen {
			continue
		}

		return fmt.Errorf("method name %q is not a supported method on Routing.Methods config param", k)
	}

	return nil
}

// Type is the routing type.
// Depending of the type we need to instantiate different Routing implementations.
type RouterType string

const (
	RouterTypeHTTP       RouterType = "http"       // HTTP JSON API for delegated routing systems (IPIP-337).
	RouterTypeDHT        RouterType = "dht"        // DHT router.
	RouterTypeSequential RouterType = "sequential" // Router helper to execute several routers sequentially.
	RouterTypeParallel   RouterType = "parallel"   // Router helper to execute several routers in parallel.
)

type DHTMode string

const (
	DHTModeServer DHTMode = "server"
	DHTModeClient DHTMode = "client"
	DHTModeAuto   DHTMode = "auto"
)

type MethodName string

const (
	MethodNameProvide       MethodName = "provide"
	MethodNameFindProviders MethodName = "find-providers"
	MethodNameFindPeers     MethodName = "find-peers"
	MethodNameGetIPNS       MethodName = "get-ipns"
	MethodNamePutIPNS       MethodName = "put-ipns"
)

var MethodNameList = []MethodName{MethodNameProvide, MethodNameFindPeers, MethodNameFindProviders, MethodNameGetIPNS, MethodNamePutIPNS}

type HTTPRouterParams struct {
	// Endpoint is the URL where the routing implementation will point to get the information.
	Endpoint string

	// MaxProvideBatchSize determines the maximum amount of CIDs sent per batch.
	// Servers might not accept more than 100 elements per batch. 100 elements by default.
	MaxProvideBatchSize int

	// MaxProvideConcurrency determines the number of threads used when providing content. GOMAXPROCS by default.
	MaxProvideConcurrency int
}

func (hrp *HTTPRouterParams) FillDefaults() {
	if hrp.MaxProvideBatchSize == 0 {
		hrp.MaxProvideBatchSize = 100
	}

	if hrp.MaxProvideConcurrency == 0 {
		hrp.MaxProvideConcurrency = runtime.GOMAXPROCS(0)
	}
}

type DHTRouterParams struct {
	Mode                 DHTMode
	AcceleratedDHTClient bool
	PublicIPNetwork      bool
}

type ComposableRouterParams struct {
	Routers []ConfigRouter
	Timeout time.Duration
}

type ConfigRouter struct {
	RouterName   string
	Timeout      time.Duration
	IgnoreErrors bool
	ExecuteAfter time.Duration
}

type Method struct {
	RouterName string
}

func Parse(cfg *config.C, extraDHT *ExtraDHTParams, extraHTTP *ExtraHTTPParams) (routing.Routing, error) {

	createdRouters := make(map[string]routing.Routing)
	finalRouter := &Composer{}
	methods := cfg.Get(("routing.methods")).(map[interface{}]interface{})
	routers := cfg.Get(("routing.routers")).(map[interface{}]interface{})

	// Create all needed routers from method names
	for methodName, routerName := range methods {

		router, err := parse(make(map[string]bool), createdRouters, routerName.(string), routers, extraDHT, extraHTTP)
		if err != nil {
			return nil, err
		}

		switch MethodName(methodName.(string)) {
		case MethodNamePutIPNS:
			finalRouter.PutValueRouter = router
		case MethodNameGetIPNS:
			finalRouter.GetValueRouter = router
		case MethodNameFindPeers:
			finalRouter.FindPeersRouter = router
		case MethodNameFindProviders:
			finalRouter.FindProvidersRouter = router
		case MethodNameProvide:
			finalRouter.ProvideRouter = router
		}

		log.Info("using method ", routerName, " with router ", routerName)
	}

	return finalRouter, nil
}

func convertRouter(r map[interface{}]interface{}) (Router, error) {
	rtype, ok := r["type"].(string)
	if !ok {
		return Router{}, errors.New("invalid or missing router type")
	}

	parameters, ok := r["parameters"].(map[interface{}]interface{})
	if !ok {
		return Router{}, errors.New("invalid or missing parameters")
	}

	p := &Router{
		Type: RouterType(rtype),
	}

	switch RouterType(rtype) {
	case RouterTypeHTTP:
		maxProvideBatchSize, _ := parameters["max_provide_batch_size"].(int)
		maxProvideConcurrency, _ := parameters["max_provide_concurrency"].(int)
		endpoint, _ := parameters["endpoint"].(string)

		p.Parameters = &HTTPRouterParams{
			Endpoint:              endpoint,
			MaxProvideBatchSize:   maxProvideBatchSize,
			MaxProvideConcurrency: maxProvideConcurrency,
		}
	case RouterTypeDHT:
		mode, _ := parameters["mode"].(DHTMode)
		acceleratedDHTClient, _ := parameters["accelerated_dht_client"].(bool)
		publicIPNetwork, _ := parameters["public_ip_network"].(bool)

		p.Parameters = &DHTRouterParams{
			Mode:                 mode,
			AcceleratedDHTClient: acceleratedDHTClient,
			PublicIPNetwork:      publicIPNetwork,
		}
	case RouterTypeSequential, RouterTypeParallel:
		timeout, _ := parameters["timeout"].(int64)
		routers, _ := parameters["routers"].([]interface{})

		composableRouterParams := &ComposableRouterParams{
			Timeout: time.Duration(timeout),
			Routers: []ConfigRouter{},
		}

		for _, r := range routers {
			rr, ok := r.(map[interface{}]interface{})
			if !ok {
				return Router{}, errors.New("invalid router configuration")
			}

			rname, _ := rr["name"].(string)
			rtimeout, _ := rr["timeout"].(int64)
			rIgnoreErrors, _ := rr["ignore_errors"].(bool)
			executeAfter, _ := rr["execute_after"].(int64)

			composableRouterParams.Routers = append(composableRouterParams.Routers, ConfigRouter{
				RouterName:   rname,
				Timeout:      time.Duration(rtimeout),
				IgnoreErrors: rIgnoreErrors,
				ExecuteAfter: time.Duration(executeAfter),
			})
		}
		p.Parameters = composableRouterParams
	default:
		return Router{}, fmt.Errorf("unsupported router type: %s", rtype)
	}
	return *p, nil
}

func parse(visited map[string]bool,
	createdRouters map[string]routing.Routing,
	routerName string,
	routersCfg map[interface{}]interface{},
	extraDHT *ExtraDHTParams,
	extraHTTP *ExtraHTTPParams,
) (routing.Routing, error) {
	// check if we already created it
	r, ok := createdRouters[routerName]
	if ok {
		return r, nil
	}
	// check if we are in a dep loop
	if visited[routerName] {
		return nil, fmt.Errorf("dependency loop creating router with name %q", routerName)
	}

	// set node as visited
	visited[routerName] = true

	cfg, ok := routersCfg[routerName].(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("config for router with name %q not found", routerName)
	}

	rtype := cfg["type"].(string)
	cfgrouter, err := convertRouter(cfg)
	if err != nil {
		return nil, err
	}
	var router routing.Routing
	switch RouterType(rtype) {
	case RouterTypeHTTP:
		router, err = httpRoutingFromConfig(cfgrouter, extraHTTP)
	case RouterTypeDHT:
		router, err = dhtRoutingFromConfig(cfgrouter, extraDHT)
	case RouterTypeParallel:
		crp := cfgrouter.Parameters.(*ComposableRouterParams)
		var pr []*routinghelpers.ParallelRouter
		for _, cr := range crp.Routers {
			ri, err := parse(visited, createdRouters, cr.RouterName, routersCfg, extraDHT, extraHTTP)
			if err != nil {
				return nil, err
			}

			pr = append(pr, &routinghelpers.ParallelRouter{
				Router:                  ri,
				IgnoreError:             cr.IgnoreErrors,
				DoNotWaitForSearchValue: true,
				Timeout:                 cr.Timeout,
				ExecuteAfter:            cr.ExecuteAfter,
			})

		}

		router = routinghelpers.NewComposableParallel(pr)
	case RouterTypeSequential:
		crp := cfgrouter.Parameters.(*ComposableRouterParams)
		var sr []*routinghelpers.SequentialRouter
		for _, cr := range crp.Routers {
			ri, err := parse(visited, createdRouters, cr.RouterName, routersCfg, extraDHT, extraHTTP)
			if err != nil {
				return nil, err
			}

			sr = append(sr, &routinghelpers.SequentialRouter{
				Router:      ri,
				IgnoreError: cr.IgnoreErrors,
				Timeout:     cr.Timeout,
			})

		}

		router = routinghelpers.NewComposableSequential(sr)
	default:
		return nil, fmt.Errorf("unknown router type %q", rtype)
	}

	if err != nil {
		return nil, err
	}

	createdRouters[routerName] = router

	log.Info("created router ", routerName, " with params ", cfgrouter.Parameters)

	return router, nil
}

type ExtraHTTPParams struct {
	PeerID     string
	Addrs      []string
	PrivKeyB64 string
}

func ConstructHTTPRouter(endpoint string, peerID string, addrs []string, privKey string) (routing.Routing, error) {
	return httpRoutingFromConfig(
		Router{
			Type: "http",
			Parameters: HTTPRouterParams{
				Endpoint: endpoint,
			},
		},
		&ExtraHTTPParams{
			PeerID:     peerID,
			Addrs:      addrs,
			PrivKeyB64: privKey,
		},
	)
}

func httpRoutingFromConfig(conf Router, extraHTTP *ExtraHTTPParams) (routing.Routing, error) {
	params := conf.Parameters.(*HTTPRouterParams)
	if params.Endpoint == "" {
		return nil, NewParamNeededErr("Endpoint", conf.Type)
	}

	params.FillDefaults()

	// Increase per-host connection pool since we are making lots of concurrent requests.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 500
	transport.MaxIdleConnsPerHost = 100

	delegateHTTPClient := &http.Client{
		Transport: &drclient.ResponseBodyLimitedTransport{
			RoundTripper: transport,
			LimitBytes:   1 << 20,
		},
	}

	key, err := decodePrivKey(extraHTTP.PrivKeyB64)
	if err != nil {
		return nil, err
	}

	addrInfo, err := createAddrInfo(extraHTTP.PeerID, extraHTTP.Addrs)
	if err != nil {
		return nil, err
	}

	cli, err := drclient.New(
		params.Endpoint,
		drclient.WithHTTPClient(delegateHTTPClient),
		drclient.WithIdentity(key),
		drclient.WithProviderInfo(addrInfo.ID, addrInfo.Addrs),
		drclient.WithUserAgent(version.GetUserAgentVersion()),
		drclient.WithStreamResultsRequired(),
		drclient.WithDisabledLocalFiltering(false),
	)
	if err != nil {
		return nil, err
	}

	cr := contentrouter.NewContentRoutingClient(
		cli,
		contentrouter.WithMaxProvideBatchSize(params.MaxProvideBatchSize),
		contentrouter.WithMaxProvideConcurrency(params.MaxProvideConcurrency),
	)

	err = view.Register(drclient.OpenCensusViews...)
	if err != nil {
		return nil, fmt.Errorf("registering HTTP delegated routing views: %w", err)
	}

	return &httpRoutingWrapper{
		ContentRouting:    cr,
		PeerRouting:       cr,
		ValueStore:        cr,
		ProvideManyRouter: cr,
	}, nil
}

func decodePrivKey(keyB64 string) (ic.PrivKey, error) {
	pk, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, err
	}

	return ic.UnmarshalPrivateKey(pk)
}

func createAddrInfo(peerID string, addrs []string) (peer.AddrInfo, error) {
	pID, err := peer.Decode(peerID)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	var mas []ma.Multiaddr
	for _, a := range addrs {
		m, err := ma.NewMultiaddr(a)
		if err != nil {
			return peer.AddrInfo{}, err
		}

		mas = append(mas, m)
	}

	return peer.AddrInfo{
		ID:    pID,
		Addrs: mas,
	}, nil
}

type ExtraDHTParams struct {
	BootstrapPeers []peer.AddrInfo
	Host           host.Host
	Validator      record.Validator
	Datastore      datastore.Batching
	Context        context.Context
}

func dhtRoutingFromConfig(conf Router, extra *ExtraDHTParams) (routing.Routing, error) {
	params, ok := conf.Parameters.(*DHTRouterParams)
	if !ok {
		return nil, errors.New("incorrect params for DHT router")
	}

	if params.AcceleratedDHTClient {
		return createFullRT(extra)
	}

	var mode dht.ModeOpt
	switch params.Mode {
	case DHTModeAuto:
		mode = dht.ModeAuto
	case DHTModeClient:
		mode = dht.ModeClient
	case DHTModeServer:
		mode = dht.ModeServer
	default:
		return nil, fmt.Errorf("invalid DHT mode: %q", params.Mode)
	}

	return createDHT(extra, params.PublicIPNetwork, mode)
}

func createDHT(params *ExtraDHTParams, public bool, mode dht.ModeOpt) (routing.Routing, error) {
	var opts []dht.Option

	if public {
		opts = append(opts, dht.QueryFilter(dht.PublicQueryFilter),
			dht.RoutingTableFilter(dht.PublicRoutingTableFilter),
			dht.RoutingTablePeerDiversityFilter(dht.NewRTPeerDiversityFilter(params.Host, 2, 3)))
	} else {
		opts = append(opts, dht.ProtocolExtension(dual.LanExtension),
			dht.QueryFilter(dht.PrivateQueryFilter),
			dht.RoutingTableFilter(dht.PrivateRoutingTableFilter))
	}

	opts = append(opts,
		dht.Concurrency(10),
		dht.Mode(mode),
		dht.Datastore(params.Datastore),
		dht.Validator(params.Validator),
		dht.BootstrapPeers(params.BootstrapPeers...))

	return dht.New(
		params.Context, params.Host, opts...,
	)
}

func createFullRT(params *ExtraDHTParams) (routing.Routing, error) {
	return fullrt.NewFullRT(params.Host,
		dht.DefaultPrefix,
		fullrt.DHTOption(
			dht.Validator(params.Validator),
			dht.Datastore(params.Datastore),
			dht.BootstrapPeers(params.BootstrapPeers...),
			dht.BucketSize(20),
		),
	)
}
