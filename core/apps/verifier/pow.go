package verifier

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Pow struct to handle Proof of Work operations
type Pow struct{}

// PerformPoW performs the Proof of Work by finding a nonce that satisfies the difficulty using multiple CPUs
func (p *Pow) PerformPoW(data string, difficulty int) (string, int) {
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

	fmt.Printf("✅ PoW CPU completed! Nonce: %d, Hash: %s, Time: %s\n", result.nonce, result.hash, time.Since(start))
	return result.hash, result.nonce
}

// VerifyPoW verifies the Proof of Work by checking the hash and nonce
func (p *Pow) VerifyPoW(data string, nonce int, hash string, difficulty int) bool {
	input := fmt.Sprintf("%s:%d", data, nonce)
	expectedHash := sha256.Sum256([]byte(input))
	expectedHashHex := hex.EncodeToString(expectedHash[:])
	targetPrefix := strings.Repeat("0", difficulty)

	// Check if the hash has the correct format
	return strings.HasPrefix(expectedHashHex, targetPrefix) && expectedHashHex == hash
}
