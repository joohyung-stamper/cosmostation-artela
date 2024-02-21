package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBaseAccountTotalAsset(t *testing.T) {
	address := "cosmos12nrtzmzxred3pkmzqwf99ccfkn7wdvaemnjhrl"
	// address := "cosmos1x5wgh6vwye60wv3dtshs9dmqggwfx2ldnqvev0"
	a, b, c, d, e, err := cli.GetBaseAccountTotalAsset(address)
	require.NoError(t, err)
	t.Log("available:", a)
	t.Log("delegated:", b)
	t.Log("undelegated:", c)
	t.Log("rewards:", d)
	t.Log("commission:", e)
}
