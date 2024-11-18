package coreiface

import "errors"

var (
	ErrOffline      = errors.New("this action must be run in online mode, try running 'subnet daemon' first")
	ErrNotSupported = errors.New("operation not supported")
)
