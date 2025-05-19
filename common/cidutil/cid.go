package cidutil

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// GenerateCID creates a CID using the provided bytes
// It uses SHA2-256 hash and DagCBOR codec by default
func GenerateCID(idBytes []byte) (cid.Cid, error) {
	mh, err := multihash.Sum(idBytes, multihash.SHA2_256, -1)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("failed to create multihash: %w", err)
	}

	return cid.NewCidV1(cid.DagCBOR, mh), nil
}
