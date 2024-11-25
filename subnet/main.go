package subnet

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	sockets "github.com/libp2p/go-socket-activation"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/core/corehttp"
	"github.com/unicornultrafoundation/subnet-node/repo/snrepo"
)

var log = logrus.New().WithField("service", "subnet")

func Main(repoPath string, configPath string) {
	if err := run(repoPath, configPath); err != nil {
		log.Fatal(err)
	}
}

func run(repoPath string, configPath string) error {
	// let the user know we're going.
	fmt.Printf("Initializing subnetnode...\n")

	r, err := snrepo.OpenWithUserConfig(repoPath, configPath)
	if err != nil {
		// TODO handle case: daemon running
		// TODO handle case: repo doesn't exist or isn't initialized
		return err
	}

	node, err := core.NewNode(context.Background(), &core.BuildCfg{
		Repo:   r,
		Online: true,
		ExtraOpts: map[string]bool{
			"pubsub": true,
		},
	})

	if err != nil {
		// TODO handle case: daemon running
		// TODO handle case: repo doesn't exist or isn't initialized
		return err
	}

	defer node.Close()

	// construct api endpoint - every time
	apiErrc, err := serveHTTPApi(r.Config(), node)
	if err != nil {
		return err
	}

	// collect long-running errors and block for shutdown
	var errs error
	for err := range merge(apiErrc) {
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func serveHTTPApi(cfg *config.C, node *core.SubnetNode) (<-chan error, error) {
	listeners, err := sockets.TakeListeners("subnet.api")
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: socket activation failed: %s", err)
	}
	apiAddrs := cfg.GetStringSlice("addresses.api", []string{})

	listenerAddrs := make(map[string]bool, len(listeners))
	for _, listener := range listeners {
		listenerAddrs[string(listener.Multiaddr().Bytes())] = true
	}
	for _, addr := range apiAddrs {
		apiMaddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("serveHTTPApi: invalid API address: %q (err: %s)", addr, err)
		}
		if listenerAddrs[string(apiMaddr.Bytes())] {
			continue
		}

		apiLis, err := manet.Listen(apiMaddr)
		if err != nil {
			return nil, fmt.Errorf("serveHTTPApi: manet.Listen(%s) failed: %s", apiMaddr, err)
		}

		listenerAddrs[string(apiMaddr.Bytes())] = true
		listeners = append(listeners, apiLis)
	}

	for _, listener := range listeners {
		// we might have listened to /tcp/0 - let's see what we are listing on
		fmt.Printf("RPC API server listening on %s\n", listener.Multiaddr())
	}

	opts := []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("api"),
		corehttp.LogOption(),
		corehttp.RoutingOption(),
		corehttp.P2PProxyOption(),
		corehttp.OpenAPIOption(),
	}

	errc := make(chan error)
	var wg sync.WaitGroup
	for _, apiLis := range listeners {
		wg.Add(1)
		go func(lis manet.Listener) {
			defer wg.Done()
			errc <- corehttp.Serve(node, manet.NetListener(lis), opts...)
		}(apiLis)
	}

	go func() {
		wg.Wait()
		close(errc)
	}()

	return errc, nil
}

// merge does fan-in of multiple read-only error channels
// taken from http://blog.golang.org/pipelines
func merge(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	for _, c := range cs {
		if c != nil {
			wg.Add(1)
			go output(c)
		}
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
