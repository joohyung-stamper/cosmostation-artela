package client

import (
	"context"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	mbltypes "github.com/cosmostation/mintscan-backend-library/types"
)

var (
	pageLimit = uint64(100)
)

// GetBaseAccountTotalAsset returns coins against bonded-denom from a delegator.
// returns spendable, delegated, undelegated, rewards, commission
func (c *Client) GetBaseAccountTotalAsset(address string) (sdktypes.Coin, sdktypes.Coin, sdktypes.Coin, sdktypes.Coin, sdktypes.Coin, error) {
	ctx := context.Background()
	denom, err := c.GRPC.GetBondDenom(ctx)
	if err != nil {
		return sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, err
	}

	available := sdktypes.NewCoin(denom, sdktypes.NewInt(0))
	delegated := sdktypes.NewCoin(denom, sdktypes.NewInt(0))
	undelegated := sdktypes.NewCoin(denom, sdktypes.NewInt(0))
	rewards := sdktypes.NewCoin(denom, sdktypes.NewInt(0))
	commission := sdktypes.NewCoin(denom, sdktypes.NewInt(0))

	resAvailable, err := c.GRPC.GetBalance(ctx, denom, address)
	if err != nil {
		return sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, err
	}
	if resAvailable != nil {
		available = available.Add(*resAvailable)
	}

	// Get total delegated coins.
	delegatorDelegationsResp, err := c.GRPC.GetDelegatorDelegations(ctx, address, pageLimit)
	if err != nil {
		return sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, err
	}

	for _, delegation := range delegatorDelegationsResp.DelegationResponses {
		delegated = delegated.Add(delegation.Balance)
	}

	// Get total undelegated coins.
	unbondingDelegationsResp, err := c.GRPC.GetDelegatorUnbondingDelegations(ctx, address, pageLimit)
	if err != nil {
		return sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, err
	}

	for _, undelegation := range unbondingDelegationsResp.UnbondingResponses {
		for _, e := range undelegation.Entries {
			undelegated = undelegated.Add(sdktypes.NewCoin(denom, e.Balance))
		}
	}

	// total Rewards
	totalRewardsResp, err := c.GRPC.GetDelegationTotalRewards(ctx, address)
	if err != nil {
		return sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, err
	}
	if totalRewardsResp != nil {
		rewards = rewards.Add(sdktypes.NewCoin(denom, totalRewardsResp.Total.AmountOf(denom).TruncateInt()))
	}

	valAddr, err := mbltypes.ConvertValAddrFromAccAddr(address)
	if err != nil {
		return sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, err
	}

	// Get total commission
	commissions, err := c.GRPC.GetValidatorCommission(ctx, valAddr)
	if err != nil {
		return sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, sdktypes.Coin{}, err
	}

	for _, c := range commissions.Commission {
		comm, _ := c.TruncateDecimal()
		commission = commission.Add(comm)
	}

	return available, delegated, undelegated, rewards, commission, nil
}
