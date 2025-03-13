package connect

import (
	"context"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("service", "connect")

func Main(peer, app string, ports []string) {
	if err := run(peer, app, ports); err != nil {
		log.Fatal(err)
	}
}

func run(peerAddr, appId string, ports []string) error {
	ctx := context.Background()
	h, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}

	// Parse peer address
	peerInfo, err := peer.AddrInfoFromString(peerAddr)
	if err != nil {
		log.Fatal("Invalid peer address:", err)
	}

	// Connect to peer
	if err := h.Connect(ctx, *peerInfo); err != nil {
		log.Fatal("Failed to connect to peer:", err)
	}

	// Start forwarding traffic for each mapping
	for _, mapping := range ports {
		ports := strings.Split(mapping, ":")
		if len(ports) != 2 {
			log.Fatal("Invalid port mapping format")
		}
		go forwardTraffic(h, peerInfo.ID, ports[0], appId, ports[1])
	}

	select {} // Keep running
}
