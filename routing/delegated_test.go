package routing

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/config"
)

func TestParser(t *testing.T) {
	require := require.New(t)

	pid, sk, err := generatePeerID()
	require.NoError(err)

	cfg := config.NewC(logrus.New())
	err = cfg.LoadString(`
routing: 
    routers:
        r1:
            type: http
            parameters:
                endpoint: test
        r2:
            type: parallel
            parameters:
                routers:
                    - name: r1
    methods:
        find-peers: r1
        find-providers: r1
        get-ipns: r1
        put-ipns: r2
        provide: r2
    
`)
	require.NoError(err)
	router, err := Parse(cfg, &ExtraDHTParams{}, &ExtraHTTPParams{
		PeerID:     string(pid),
		PrivKeyB64: sk,
	})

	require.NoError(err)

	comp, ok := router.(*Composer)
	require.True(ok)

	require.Equal(comp.FindPeersRouter, comp.FindProvidersRouter)
	require.Equal(comp.ProvideRouter, comp.PutValueRouter)
}

func TestParserRecursive(t *testing.T) {
	require := require.New(t)

	pid, sk, err := generatePeerID()
	require.NoError(err)

	cfg := config.NewC(logrus.New())
	err = cfg.LoadString(`
routing: 
    routers:
        r1:
            type: http
            parameters:
                endpoint: test
        r2:
            type: parallel
            parameters:
                routers:
                    - name: r1
    methods:
        find-peers: r1
        find-providers: r1
        get-ipns: r1
        put-ipns: r2
        provide: r2
    
`)
	require.NoError(err)
	router, err := Parse(cfg, &ExtraDHTParams{}, &ExtraHTTPParams{
		PeerID:     string(pid),
		PrivKeyB64: sk,
	})

	require.NoError(err)

	_, ok := router.(*Composer)
	require.True(ok)
}

func TestParserRecursiveLoop(t *testing.T) {
	require := require.New(t)
	cfg := config.NewC(logrus.New())
	err := cfg.LoadString(`
    routing: 
        routers:
            composable1:
                type: sequential
                parameters:
                    routers:
                        - name: composable2
            composable2:
                type: parallel
                parameters:
                    routers:
                        - name: composable1
        methods:
            find-peers: composable2
            find-providers: composable2
            get-ipns: composable2
            put-ipns: composable2
            provide: composable2
        
    `)
	require.NoError(err)
	_, err = Parse(cfg, &ExtraDHTParams{}, nil)

	require.ErrorContains(err, "dependency loop creating router with name \"composable2\"")
}

func generatePeerID() (string, string, error) {
	sk, pk, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return "", "", err
	}

	bytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		return "", "", err
	}

	enc := base64.StdEncoding.EncodeToString(bytes)
	if err != nil {
		return "", "", err
	}

	pid, err := peer.IDFromPublicKey(pk)
	return pid.String(), enc, err
}
