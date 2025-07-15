//go:build tools

package tools

// nolint
import (
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/vektra/mockery/v2"
	_ "k8s.io/code-generator"
)
