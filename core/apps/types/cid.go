package apps

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// GenerateCID creates a CID using only the app ID
func (app *App) GenerateCID() (cid.Cid, error) {
	idBytes := app.ID.Bytes()

	mh, err := multihash.Sum(idBytes, multihash.SHA2_256, -1)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("failed to create multihash: %w", err)
	}

	return cid.NewCidV1(cid.DagCBOR, mh), nil
}
