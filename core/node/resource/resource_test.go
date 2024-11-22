package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResource(t *testing.T) {
	r, err := GetResource()
	assert.NoError(t, err)
	assert.Greater(t, r.CPU.Count, 0)
}
