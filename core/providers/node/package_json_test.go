package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageJson(t *testing.T) {
	packageJson := NewPackageJson()
	assert.Equal(t, packageJson.HasScript("start"), false)
}
