package verifier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"

	ggio "github.com/gogo/protobuf/io"
	proto "github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	pvtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/app/verifier"
)

const (
	defaultDifficulty = 6
)

type NodeType int

const (
	NodeVerifier NodeType = iota
	NodeProvider
)

// Pow struct to handle Proof of Work operations
type Pow struct {
	requests       map[string]time.Time
	ps             p2phost.Host
	p2p            *p2p.P2P
	qualifiedPeers map[string]bool
	nodeType       NodeType
	verifierPeers  map[string]bool
	mu             sync.RWMutex
}

// NewPow creates a new instance of Pow
func NewPow(nodeType NodeType, ps p2phost.Host, P2P *p2p.P2P) *Pow {
	pow := &Pow{
		ps:             ps,
		requests:       make(map[string]time.Time),
		p2p:            P2P,
		qualifiedPeers: make(map[string]bool),
		nodeType:       nodeType,
		verifierPeers:  make(map[string]bool),
	}

	if nodeType == NodeVerifier {
		ps.SetStreamHandler(atypes.ProtocolAppPoWResponse, pow.OnPoWResponse)
	} else if nodeType == NodeProvider {
		ps.SetStreamHandler(atypes.ProtocolAppPoWRequest, pow.OnPoWRequest)
	}

	return pow
}

// PerformPoW performs the Proof of Work by finding a nonce that satisfies the difficulty using multiple CPUs
func (p *Pow) performPoW(data string, difficulty int) (string, int) {
	var wg sync.WaitGroup
	numCPU := runtime.NumCPU()
	targetPrefix := strings.Repeat("0", difficulty)
	resultChan := make(chan struct {
		hash  string
		nonce int
	})
	stopChan := make(chan struct{})
	start := time.Now()

	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go func(startNonce int) {
			defer wg.Done()
			nonce := startNonce
			for {
				select {
				case <-stopChan:
					return
				default:
					input := fmt.Sprintf("%s:%d", data, nonce)
					hash := sha256.Sum256([]byte(input))
					hashHex := hex.EncodeToString(hash[:])

					// Check if the hash starts with `difficulty` number of zeros
					if strings.HasPrefix(hashHex, targetPrefix) {
						select {
						case resultChan <- struct {
							hash  string
							nonce int
						}{hashHex, nonce}:
							return
						case <-stopChan:
							return
						}
					}
					nonce += numCPU
				}
			}
		}(i)
	}

	result := <-resultChan
	close(stopChan)
	wg.Wait()

	log.Infof("✅ PoW CPU completed! Nonce: %d, Hash: %s, Time: %s\n", result.nonce, result.hash, time.Since(start))
	return result.hash, result.nonce
}

// VerifyPoW verifies the Proof of Work by checking the hash and nonce
func (p *Pow) verifyPoW(data string, nonce int, hash string, difficulty int) bool {
	input := fmt.Sprintf("%s:%d", data, nonce)
	expectedHash := sha256.Sum256([]byte(input))
	expectedHashHex := hex.EncodeToString(expectedHash[:])
	targetPrefix := strings.Repeat("0", int(difficulty))

	// Check if the hash has the correct format
	return strings.HasPrefix(expectedHashHex, targetPrefix) && expectedHashHex == hash
}

func (p *Pow) RequestPoW(peerID peer.ID) {
	powRequest := &pvtypes.PowRequest{
		Id:         uuid.New().String(),
		Difficulty: defaultDifficulty,
	}

	p.requests[powRequest.Id] = time.Now()

	ok := p.sendProtoMessage(peerID, atypes.ProtocolAppPoWRequest, powRequest)
	if !ok {
		log.Warnf("Failed to send PoW request to peer %s", peerID)
	} else {
		log.Infof("Sent PoW request to peer %s", peerID)
	}
}

func (p *Pow) RequestPoWFromPeers(peerIDs []peer.ID, workerCount int) {
	var wg sync.WaitGroup
	jobs := make(chan peer.ID, len(peerIDs))

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for peerID := range jobs {
				p.RequestPoW(peerID)
			}
		}()
	}

	for _, peerID := range peerIDs {
		jobs <- peerID
	}
	close(jobs)

	wg.Wait()
}

func (p *Pow) OnPoWResponse(s network.Stream) {
	msg := &pvtypes.PowResponse{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Error(err)
		return
	}
	s.Close()

	err = proto.Unmarshal(buf, msg)
	if err != nil {
		log.Error(err)
		return
	}

	peerID := s.Conn().RemotePeer().String()
	powRequest, exists := p.requests[msg.Id]
	if !exists {
		log.Warnf("Received unexpected PoW response from %s", peerID)
		return
	}
	duration := time.Since(powRequest)
	expectedMaxPoWTime := time.Duration(msg.Nonce/5000)*time.Millisecond + 500*time.Millisecond // Adjust this based on your requirements
	log.Infof("Received PoW response from %s. Duration: %s, Nonce: %d, Hash: %s, Expected Max Time: %s", peerID, duration, msg.Nonce, msg.Hash, expectedMaxPoWTime)
	// Verify PoW
	if p.verifyPoW(msg.Id, int(msg.Nonce), msg.Hash, 6) {
		if duration <= expectedMaxPoWTime {
			log.Infof("PoW verified successfully for peer %s", peerID)
			p.qualifiedPeers[peerID] = true
		} else {
			log.Warnf("PoW verification failed for peer %s: exceeded maximum time", peerID)
		}
	} else {
		log.Warnf("PoW verification failed for peer %s", peerID)
	}

	delete(p.requests, peerID)
}

func (p *Pow) OnPoWRequest(s network.Stream) {
	peerID := s.Conn().RemotePeer().String()
	if !p.IsVerifierPeer(peerID) {
		log.Warnf("PoW request from non-verifier peer %s rejected", peerID)
		s.Reset()
		return
	}

	msg := &pvtypes.PowRequest{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Error(err)
		return
	}
	s.Close()

	err = proto.Unmarshal(buf, msg)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Received PoW request from %s. Data: %s, Difficulty: %d", s.Conn().RemotePeer(), msg.Id, msg.Difficulty)

	hash, nonce := p.performPoW(msg.Id, int(msg.Difficulty))

	response := &pvtypes.PowResponse{
		Id:    msg.Id,
		Hash:  hash,
		Nonce: int64(nonce),
	}

	ok := p.sendProtoMessage(s.Conn().RemotePeer(), atypes.ProtocolAppPoWResponse, response)
	if !ok {
		log.Warnf("Failed to send PoW response to peer %s", s.Conn().RemotePeer())
	} else {
		log.Infof("Sent PoW response to peer %s", s.Conn().RemotePeer())
	}
}

func (p *Pow) UpdateVerifierPeers(peerIDs []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.verifierPeers = make(map[string]bool)
	for _, peerID := range peerIDs {
		p.verifierPeers[peerID] = true
	}
}

func (p *Pow) AddVerifierPeer(peerID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.verifierPeers[peerID] = true
}

func (p *Pow) RemoveVerifierPeer(peerID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.verifierPeers, peerID)
}

func (p *Pow) IsVerifierPeer(peerID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.verifierPeers[peerID]
}

func (p *Pow) GetQualifiedPeers() []string {
	peers := make([]string, 0, len(p.qualifiedPeers))
	for peerID := range p.qualifiedPeers {
		peers = append(peers, peerID)
	}
	return peers
}

func (p *Pow) IsPeerQualified(peerID string) bool {
	return p.qualifiedPeers[peerID]
}

func (p *Pow) Clear() {
	p.requests = make(map[string]time.Time)
	p.qualifiedPeers = make(map[string]bool)
}

func (pow *Pow) sendProtoMessage(id peer.ID, p protocol.ID, data proto.Message) bool {
	s, err := pow.ps.NewStream(context.Background(), id, p)
	if err != nil {
		log.Error(err)
		return false
	}
	defer s.Close()

	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.Error(err)
		s.Reset()
		return false
	}
	return true
}
