package sdl

import (
	"errors"
	"fmt"
	"os"

	"github.com/blang/semver"
	"gopkg.in/yaml.v3"
)

var (
	errUninitializedConfig = errors.New("sdl: uninitialized")
	errSDLInvalidNoVersion = fmt.Errorf("%w: no version found", errSDLInvalid)
)

const (
	sdlVersionField = "version"
)

type SDL interface {
}

var _ SDL = (*sdl)(nil)

type sdl struct {
	Ver  semver.Version `yaml:"version,-"`
	data SDL            `yaml:"-"`
}

func (s *sdl) UnmarshalYAML(node *yaml.Node) error {
	var result sdl

	foundVersion := false
	for idx := range node.Content {
		if node.Content[idx].Value == sdlVersionField {
			var err error
			if result.Ver, err = semver.ParseTolerant(node.Content[idx+1].Value); err != nil {
				return err
			}
			foundVersion = true
			break
		}
	}

	if !foundVersion {
		return errSDLInvalidNoVersion
	}

	// nolint: gocritic
	if result.Ver.EQ(semver.MustParse("1.0.0")) {
		var decoded v1SDL
		if err := node.Decode(&decoded); err != nil {
			return err
		}

		result.data = &decoded
	} else {
		return fmt.Errorf("%w: config: unsupported version %q", errSDLInvalid, result.Ver)
	}

	*s = result

	return nil
}

// ReadFile read from given path and returns SDL instance
func ReadFile(path string) (SDL, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Read(buf)
}

// Read reads buffer data and returns SDL instance
func Read(buf []byte) (SDL, error) {
	obj := &sdl{}
	if err := yaml.Unmarshal(buf, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
