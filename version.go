package subnetnode

import (
	"runtime"
)

// CurrentCommit is the current git commit, this is set as a ldflag in the Makefile.
var CurrentCommit string

// CurrentVersionNumber is the current application's version literal.
const CurrentVersionNumber = "0.0.1-dev"

const ApiVersion = "/subnet/" + CurrentVersionNumber + "/" //nolint

// Note: This will end in `/` when no commit is available. This is expected.
func GetUserAgentVersion() string {
	userAgent := "subnet/" + CurrentVersionNumber + "/" + CurrentCommit
	if userAgentSuffix != "" {
		if CurrentCommit != "" {
			userAgent += "/"
		}
		userAgent += userAgentSuffix
	}
	return userAgent
}

var userAgentSuffix string

func SetUserAgentSuffix(suffix string) {
	userAgentSuffix = suffix
}

type VersionInfo struct {
	Version string
	Commit  string
	System  string
	Golang  string
}

func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version: CurrentVersionNumber,
		Commit:  CurrentCommit,
		System:  runtime.GOARCH + "/" + runtime.GOOS, // TODO: Precise version here
		Golang:  runtime.Version(),
	}
}
