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

// GetDHTKey returns the DHT key for storing/retrieving peer IDs
func (app *App) GetDHTKey() (string, error) {
	c, err := app.GenerateCID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/app/%s", c.String()), nil
}
