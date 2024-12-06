package core

import (
	"context"
	"testing"

	"github.com/ipfs/go-datastore"
	syncds "github.com/ipfs/go-datastore/sync"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

func TestInitialization(t *testing.T) {
	ctx := context.Background()
	// Good config
	good := []string{
		`
identity:
  peer_id: "12D3KooWLJCgSFv62DfQwNKTYxACMisLnCnGbYBp766p1K19PExx"
  privkey: "CAESQA0P3Td0UZ2sAoCflMNdUivxFbUdcuo+XraGbZk5EZEdm7ZnWjxFiBUG9M718wpSOzl8P4JDQe6vZ6w+F5S9VPE="
addresses:
  swarm:
    - /ip4/0.0.0.0/tcp/4001
  api:
    - /ip4/0.0.0.0/tcp/8080
`,
	}
	bad := []string{``}

	for i, cr := range good {
		c := config.NewC(logrus.New())
		err := c.LoadString(cr)
		if err != nil {
			t.Fatal(err)
		}

		r := &repo.Mock{
			C: c,
			D: syncds.MutexWrap(datastore.NewMapDatastore()),
		}
		n, err := NewNode(ctx, &BuildCfg{Repo: r})
		if n == nil || err != nil {
			t.Error("Should have constructed.", i, err)
		}
	}

	for i, cr := range bad {
		c := config.NewC(logrus.New())
		c.LoadString(cr)
		r := &repo.Mock{
			C: c,
			D: syncds.MutexWrap(datastore.NewMapDatastore()),
		}
		n, err := NewNode(ctx, &BuildCfg{Repo: r})
		if n != nil || err == nil {
			t.Error("Should have failed to construct.", i)
		}
	}

}
