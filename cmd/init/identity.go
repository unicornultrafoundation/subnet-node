package init

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/crypto"
	ic "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	IdentityTag     = "Identity"
	PrivKeyTag      = "PrivKey"
	PrivKeySelector = IdentityTag + "." + PrivKeyTag
)

// Identity tracks the configuration of the local node's identity.
type Identity struct {
	PeerID  string
	PrivKey string `json:",omitempty"`
}

// DecodePrivateKey is a helper to decode the users PrivateKey.
func (i *Identity) DecodePrivateKey(passphrase string) (ic.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(i.PrivKey)
	if err != nil {
		return nil, err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	return ic.UnmarshalPrivateKey(pkb)
}

// CreateIdentity initializes a new identity.
func CreateIdentity(out io.Writer) (Identity, error) {
	// TODO guard higher up
	ident := Identity{}

	var sk crypto.PrivKey
	var pk crypto.PubKey

	fmt.Fprintf(out, "generating ED25519 keypair...")
	priv, pub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return ident, err
	}

	sk = priv
	pk = pub
	fmt.Fprintf(out, "done\n")

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	skbytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		return ident, err
	}
	ident.PrivKey = base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return ident, err
	}
	ident.PeerID = id.String()
	fmt.Fprintf(out, "peer identity: %s\n", ident.PeerID)
	return ident, nil
}
