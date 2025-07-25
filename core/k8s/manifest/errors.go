package manifest

import "errors"

var (
	// ErrShutdownTimerExpired for a terminating deployment
	ErrShutdownTimerExpired = errors.New("shutdown timer expired")
	// ErrManifestVersion indicates that the given manifest's version does not
	// match the blockchain Version value.
	ErrManifestVersion         = errors.New("manifest version validation failed")
	ErrNoManifestForDeployment = errors.New("manifest not yet received for that deployment")
	ErrNoLeaseForDeployment    = errors.New("no lease for deployment")
	errNoGroupForLease         = errors.New("group not found")
	errManifestRejected        = errors.New("manifest rejected")
)
