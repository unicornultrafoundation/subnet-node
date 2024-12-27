package subnet

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	sockets "github.com/libp2p/go-socket-activation"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/sirupsen/logrus"
	ninit "github.com/unicornultrafoundation/subnet-node/cmd/init"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/core/corehttp"
	"github.com/unicornultrafoundation/subnet-node/repo/snrepo"
)

var log = logrus.New().WithField("service", "subnet")

func Main(repoPath string, configPath *string) {
	if err := run(repoPath, configPath); err != nil {
		log.Fatal(err)
	}
}

func run(repoPath string, configPath *string) error {
	// let the user know we're going.
	fmt.Printf("Initializing Subnet Node...\n")

	if !snrepo.IsInitialized(repoPath) {
		_, err := ninit.Init(repoPath, os.Stdout)
		if err != nil {
			return err
		}
	}

	r, err := snrepo.Open(repoPath, configPath)
	if err != nil {
		// TODO handle case: daemon running
		// TODO handle case: repo doesn't exist or isn't initialized
		return err
	}

	defer r.Close()

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

	printLibp2pPorts(node)

	// construct api endpoint - every time
	apiErrc, err := serveHTTPApi(r.Config(), node)
	if err != nil {
		return err
	}

	// node.Repo.Config().RegisterReloadCallback(func(c *config.C) {
	// 	node.Close()
	// })

	// collect long-running errors and block for shutdown
	var errs error
	for err := range merge(apiErrc) {
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func printLibp2pPorts(node *core.SubnetNode) {
	if !node.IsOnline {
		fmt.Println("Swarm not listening, running in offline mode.")
		return
	}

	ifaceAddrs, err := node.PeerHost.Network().InterfaceListenAddresses()
	if err != nil {
		log.Errorf("failed to read listening addresses: %s", err)
	}

	// Multiple libp2p transports can use same port.
	// Deduplicate all listeners and collect unique IP:port (udp|tcp) combinations
	// which is useful information for operator deploying Kubo in TCP/IP infra.
	addrMap := make(map[string]map[string]struct{})
	re := regexp.MustCompile(`^/(?:ip[46]|dns(?:[46])?)/([^/]+)/(tcp|udp)/(\d+)(/.*)?$`)
	for _, addr := range ifaceAddrs {
		matches := re.FindStringSubmatch(addr.String())
		if matches != nil {
			hostname := matches[1]
			protocol := strings.ToUpper(matches[2])
			port := matches[3]
			var host string
			if matches[0][:4] == "/ip6" {
				host = fmt.Sprintf("[%s]:%s", hostname, port)
			} else {
				host = fmt.Sprintf("%s:%s", hostname, port)
			}
			if _, ok := addrMap[host]; !ok {
				addrMap[host] = make(map[string]struct{})
			}
			addrMap[host][protocol] = struct{}{}
		}
	}

	// Produce a sorted host:port list
	hosts := make([]string, 0, len(addrMap))
	for host := range addrMap {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	// Print listeners
	for _, host := range hosts {
		protocolsSet := addrMap[host]
		protocols := make([]string, 0, len(protocolsSet))
		for protocol := range protocolsSet {
			protocols = append(protocols, protocol)
		}
		sort.Strings(protocols)
		fmt.Printf("Swarm listening on %s (%s)\n", host, strings.Join(protocols, "+"))
	}
	fmt.Printf("Run 'subnet id' to inspect announced and discovered multiaddrs of this node.\n")
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
		corehttp.APIOption(),
		corehttp.StatusOption(), // Add the StatusOption here
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
