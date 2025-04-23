package packet

import (
	"errors"
)

// Common errors
var (
	ErrEmptyPacket          = errors.New("empty packet")
	ErrInvalidPacketLength  = errors.New("invalid packet length")
	ErrUnsupportedIPVersion = errors.New("unsupported IP version")
)
