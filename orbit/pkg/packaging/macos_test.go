package packaging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMacOS(t *testing.T) {
	err := signPkg("/Users/zwass/Desktop/fleet-osquery.pkg", "foobar")
	require.NoError(t, err)
}
