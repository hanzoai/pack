package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromEnvs(t *testing.T) {
	env, err := FromEnvs([]string{
		"VAR1=value1",
		"VAR2=value2",
		"RAILPACK_APT_PACKAGES=apt1,apt2",
	})

	require.NoError(t, err)
	require.Equal(t, env.GetVariable("VAR1"), "value1")
	require.Equal(t, env.GetVariable("VAR2"), "value2")
	require.Equal(t, env.GetVariable("RAILPACK_APT_PACKAGES"), "apt1,apt2")
}
