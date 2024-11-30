package merkle

import "github.com/unicornultrafoundation/subnet-node/crypto/hash"

// TODO: make these have a large predefined capacity
var (
	leafPrefix  = []byte{0}
	innerPrefix = []byte{1}
)

// returns hash(<empty>)
func emptyHash() []byte {
	return hash.Sum([]byte{})
}

// returns hash(0x00 || leaf)
func leafHash(leaf []byte) []byte {
	return hash.Sum(append(leafPrefix, leaf...))
}

// returns tmhash(0x01 || left || right)
func innerHash(left []byte, right []byte) []byte {
	return hash.Sum(append(innerPrefix, append(left, right...)...))
}
