package custom

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

var (
	// config로 빼자
	NonNativeAssets = []string{}
	PowerReduction  = sdktypes.NewIntFromUint64(1000000) // 1e6
)
