//go:build !boringcrypto
// +build !boringcrypto

package subnet

var boringEnabled = func() bool { return false }
