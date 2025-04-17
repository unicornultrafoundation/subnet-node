package vpn

import (
	"io"
	"strconv"

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
						// Reject all requests that are to unallowed ports
						unallowedPort, exist := s.unallowedPorts[dstPort]
						if exist && unallowedPort {
							continue
						}

						// Reject all requests which destination ports are not in range 30000-65535
						if *packetInfo.DstPort < 30000 || *packetInfo.DstPort > 65535 {
							continue
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
