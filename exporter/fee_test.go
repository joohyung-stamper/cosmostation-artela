package exporter

import (
	"testing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGetFee(t *testing.T) {
	beginTxID := int64(2725400)
	p, fees, dailyfees, err := ex.GetFees(beginTxID)
	require.NoError(t, err)
	t.Log(p, fees, dailyfees)
}

func TestMap(t *testing.T) {

	a := make([]sdktypes.Coin, 0)
	a = append(a, sdktypes.NewCoin("uatom", sdktypes.NewInt(0)))
	a = append(a, sdktypes.NewCoin("uatom", sdktypes.NewInt(0)))
	a = append(a, sdktypes.NewCoin("uatom", sdktypes.NewInt(0)))
	a = append(a, sdktypes.NewCoin("utest", sdktypes.NewInt(3)))

	// as := sdktypes.NewCoins(a...)

	mm := make(map[string]sdktypes.Int)

	for i := range a {
		uc, ok := mm[a[i].Denom]
		if !ok {
			uc = sdktypes.ZeroInt()
		}
		sum := uc.Add(a[i].Amount)
		mm[a[i].Denom] = sum
	}

	for k, v := range mm {
		t.Log(k, v)
	}

	b := make([]sdktypes.Coin, 0)
	b = append(b, sdktypes.NewCoin("uatom", sdktypes.NewInt(0)))
	b = append(b, sdktypes.NewCoin("utest", sdktypes.NewInt(3)))
	for i := range b {
		uc, ok := mm[b[i].Denom]
		if !ok {
			uc = sdktypes.ZeroInt()
		}
		sum := uc.Add(b[i].Amount)
		mm[b[i].Denom] = sum
	}

	for k, v := range mm {
		t.Log(k, v)
	}

}
