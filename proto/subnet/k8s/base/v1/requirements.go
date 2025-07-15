package v1

import (
	"gopkg.in/yaml.v3"
)

func (m *PlacementRequirements) String() string {
	res, _ := yaml.Marshal(m)
	return string(res)
}
