package vpn

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/songgao/water"
)

// HandleP2PTraffic listens for incoming P2P streams
func (s *Service) HandleP2PTraffic(iface *water.Interface) {
	s.PeerHost.SetStreamHandler(VPNProtocol, func(stream network.Stream) {
		go func() {
			defer stream.Close()

			buf := make([]byte, s.mtu)
			for {
				n, err := stream.Read(buf)
				if err != nil {
					if err == io.EOF {
						log.Debugf("P2P stream EOF from %s", stream.Conn().RemotePeer())
					} else {
						log.Errorf("error reading P2P stream: %v", err)
					}
					return
				}

				packet := make([]byte, n)
				copy(packet, buf[:n])

				packetInfo, err := ExtractIPAndPorts(packet)
				if err != nil {
					log.Debugf("error extracting packet info: %v", err)
					continue
				} else {
					if packetInfo.DstPort != nil {
						dstPort := strconv.Itoa(*packetInfo.DstPort)
						// Reject all requests that are in unallowed ports
						unallowedPort, exist := s.unallowedPorts[dstPort]
						if exist && unallowedPort {
							continue
						}

						if s.IsProvider {
							// Reject all requests that aren't in exposed app ports list
							allowedPort, exist := s.exposedPorts[dstPort]
							if !exist || !allowedPort {
								continue
							}
						}
					}

					_, err = iface.Write(packet)
					if err != nil {
						log.Errorf("failed to write to iface %s: %v", iface.Name(), err)
						return
					}
				}
			}
		}()
	})
}

func (s *Service) startUpdatingExposedPorts(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Debugf("Starting updating exposed app ports...")
			s.UpdateAllAppExposedPorts(ctx)
		case <-s.stopChan:
			log.Infof("Stopping updating exposed app ports...")
			return
		case <-ctx.Done():
			log.Infof("Context canceled, stopping updating exposed app ports...")
			return
		}
	}
}

func (s *Service) UpdateAllAppExposedPorts(ctx context.Context) error {
	exposedPorts := make(map[string]bool)
	runningAppList, err := s.Apps.GetRunningAppListProto(ctx)
	if err != nil {
		return fmt.Errorf("failed to get running app list for get their ports: %v", err)
	}

	for _, appIdBytes := range runningAppList.AppIds {
		appId := new(big.Int).SetBytes(appIdBytes)

		container, err := s.Apps.ContainerInspect(ctx, appId)
		if err != nil {
			log.Debugf("failed to get container from appId %s: %v", appId, err)
			continue
		}

		for _, ports := range container.NetworkSettings.Ports {
			for _, port := range ports {
				exposedPorts[port.HostPort] = true
			}
		}
	}

	s.exposedPorts = exposedPorts

	return nil
}
