//go:build tools

package tools

// nolint
import (
	_ "github.com/regen-network/cosmos-proto/protoc-gen-gocosmos"
	_ "github.com/vektra/mockery/v2"
	_ "k8s.io/code-generator"
)
