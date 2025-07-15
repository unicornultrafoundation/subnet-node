package types

import "github.com/unicornultrafoundation/subnet-node/core/types"

type GroupSpec struct {
	Name      string          `json:"name" yaml:"name"`
	Resources types.Resources `json:"resources" yaml:"resources"`
}
