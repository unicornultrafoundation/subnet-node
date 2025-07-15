package testutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

// AccAddress provides an Account's Address bytes from a ed25519 generated
// private key.
func AccAddress(t testing.TB) common.Address {
	privKey := Key(t)
	return crypto.PubkeyToAddress(privKey.PublicKey)
}

func Key(t testing.TB) *ecdsa.PrivateKey {
	t.Helper()
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return privKey
}

func DeploymentID(t testing.TB) dtypes.DeploymentID {
	t.Helper()
	return dtypes.DeploymentID{
		Owner: AccAddress(t).String(),
		DSeq:  uint64(rand.Uint32()), // nolint: gosec
	}
}

func DeploymentIDForAccount(t testing.TB, addr common.Address) dtypes.DeploymentID {
	t.Helper()
	return dtypes.DeploymentID{
		Owner: addr.String(),
		DSeq:  uint64(rand.Uint32()), // nolint: gosec
	}
}

// DeploymentVersion provides a random sha256 sum for simulating Deployments.
func DeploymentVersion(t testing.TB) []byte {
	t.Helper()
	src := make([]byte, 128)
	_, err := cryptorand.Read(src)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(src)
	return sum[:]
}

func GroupID(t testing.TB) dtypes.GroupID {
	t.Helper()
	return dtypes.MakeGroupID(DeploymentID(t), rand.Uint32()) // nolint: gosec
}

func GroupIDForAccount(t testing.TB, addr common.Address) dtypes.GroupID {
	t.Helper()
	return dtypes.MakeGroupID(DeploymentIDForAccount(t, addr), rand.Uint32()) // nolint: gosec
}

func OrderID(t testing.TB) mtypes.OrderID {
	t.Helper()
	return mtypes.MakeOrderID(GroupID(t), rand.Uint32()) // nolint: gosec
}

func OrderIDForAccount(t testing.TB, addr common.Address) mtypes.OrderID {
	t.Helper()
	return mtypes.MakeOrderID(GroupIDForAccount(t, addr), rand.Uint32()) // nolint: gosec
}

func BidID(t testing.TB) mtypes.BidID {
	t.Helper()
	return mtypes.MakeBidID(OrderID(t), AccAddress(t))
}

func BidIDForAccount(t testing.TB, owner, provider common.Address) mtypes.BidID {
	t.Helper()
	return mtypes.MakeBidID(OrderIDForAccount(t, owner), provider)
}

func LeaseID(t testing.TB) mtypes.LeaseID {
	t.Helper()
	return mtypes.MakeLeaseID(BidID(t))
}

func LeaseIDForAccount(t testing.TB, owner, provider common.Address) mtypes.LeaseID {
	t.Helper()
	return mtypes.MakeLeaseID(BidIDForAccount(t, owner, provider))
}
