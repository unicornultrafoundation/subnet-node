package subnetnode

import (
	"runtime"
)

// CurrentCommit is the current git commit, this is set as a ldflag in the Makefile.
var CurrentCommit string

// CurrentVersionNumber is the current application's version literal.
var CurrentVersionNumber string

var ApiVersion = "/subnet/" + CurrentVersionNumber + "/" //nolint

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
	Version string `json:"version"`
	Commit  string `json:"commit"`
	System  string `json:"system"`
	Golang  string `json:"golang"`
}

func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version: CurrentVersionNumber,
		Commit:  CurrentCommit,
		System:  runtime.GOARCH + "/" + runtime.GOOS,
		Golang:  runtime.Version(),
	}
}
