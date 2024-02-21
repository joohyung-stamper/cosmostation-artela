package model

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ActiveValidator is a state when a validator is bonded.
	ActiveValidator = "active"

	// InactiveValidator is a state when a validator is either unbonding or unbonded.
	InactiveValidator = "inactive"

	// MissingAllBlocks is a number of missing blocks when a validator is in unbonding or unbonded state.
	MissingAllBlocks = 100
)

// Uptime defines the structure for a validator's liveness of the last 100 blocks.
type Uptime struct {
	Address      string `json:"address"`
	MissedBlocks int    `json:"missed_blocks"`
	OverBlocks   int    `json:"over_blocks"`
}

// ValidatorDelegations defines the structure for validator's delegations.
type ValidatorDelegations struct {
	DelegatorAddress string  `json:"delegator_address"`
	ValidatorAddress string  `json:"validator_address"`
	Shares           sdk.Dec `json:"shares"`
	Amount           string  `json:"amount"`
}
