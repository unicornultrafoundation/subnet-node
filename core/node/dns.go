package node

import (
	"fmt"
	"math"
	"time"

	"github.com/ipfs/boxo/gateway"
	doh "github.com/libp2p/go-doh-resolver"
	madns "github.com/multiformats/go-multiaddr-dns"
	"github.com/unicornultrafoundation/subnet-node/config"
)

func DNSResolver(cfg *config.C) (*madns.Resolver, error) {
	var dohOpts []doh.Option
	dohOpts = append(dohOpts, doh.WithMaxCacheTTL(cfg.GetDuration("dns.max_cache_ttl", time.Duration(math.MaxUint32)*time.Second)))
	resolvers := make(map[string]string)
	iresolvers, ok := cfg.Get("dns.resolvers").(map[interface{}]interface{})
	if ok {
		for key, value := range iresolvers {
			strKey, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("invalid key type in dns.resolvers: %T", key)
			}
			strValue, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value type in dns.resolvers: %T", value)
			}
			resolvers[strKey] = strValue
		}
	}
	return gateway.NewDNSResolver(resolvers, dohOpts...)
}
