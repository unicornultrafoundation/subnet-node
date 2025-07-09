package types

// Attribute represents key value pair
type Attribute struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Attributes []Attribute
