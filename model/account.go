package model

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// UnbondingDelegations defines the structure for unbonding delegations.
type UnbondingDelegations struct {
	stakingtypes.UnbondingDelegation
	// DelegatorAddress string                                  `json:"delegator_address"`
	// ValidatorAddress string                                  `json:"validator_address"`
	// Entries          []stakingtypes.UnbondingDelegationEntry `json:"entries"`
	Moniker string `json:"moniker"`
}

// ModuleAccount defines the structure for module account information.
// 향후 사용할 수 있어서 삭제 안함(Jeonghwan)
type ModuleAccount struct {
	Name          string    `json:"name"`
	Permissions   []string  `json:"permissions"`
	Address       string    `json:"address"`
	AccountNumber uint64    `json:"account_number"`
	Coins         sdk.Coins `json:"coins"`
}
