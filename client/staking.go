package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmostation/cosmostation-coreum/custom"
	mbltypes "github.com/cosmostation/mintscan-backend-library/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"

	//cosmos-sdk
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetValidatorsByStatus 는 MBL의 GetValidatorsByStatus을 wrap한 함수 (codec 사용을 분리하기 위해)
// 필요한 함수를 우선 모듈에 맞게 정의한 후, 나중에 코어로 이전
// 코어로 분리가 가능할 것 같다.
func (c *Client) GetValidatorsByStatus(ctx context.Context, status stakingtypes.BondStatus) (validators []mdschema.Validator, err error) {
	res, err := c.GRPC.GetValidatorsByStatus(ctx, status, 2000)
	if err != nil {
		return []mdschema.Validator{}, nil
	}

	if res == nil {
		return []mdschema.Validator{}, nil
	}

	for i, val := range res.Validators {
		accAddr, err := mbltypes.ConvertAccAddrFromValAddr(val.OperatorAddress)
		if err != nil {
			return []mdschema.Validator{}, fmt.Errorf("failed to convert address from validator Address : %s", err)
		}

		var conspubkey cryptotypes.PubKey
		custom.AppCodec.UnpackAny(val.ConsensusPubkey, &conspubkey)

		valconspub, err := sdktypes.Bech32ifyAddressBytes(sdktypes.GetConfig().GetBech32ConsensusPubPrefix(), conspubkey.Bytes())
		if err != nil {
			return []mdschema.Validator{}, fmt.Errorf("failed to get consesnsus pubkey : %s", err)
		}

		// log.Println("conspubkey get cached value : ", val.ConsensusPubkey.GetCachedValue())
		// conspubkey, err := val.TmConsPubKey()
		// if err != nil {
		// 	return []schema.Validator{}, fmt.Errorf("failed to get consesnsus pubkey : %s", err)
		// }

		v := mdschema.Validator{
			Rank:                 i + 1,
			OperatorAddress:      val.OperatorAddress,
			Address:              accAddr,
			ConsensusPubkey:      valconspub,
			Proposer:             conspubkey.Address().String(),
			Jailed:               val.Jailed,
			Status:               int(val.Status),
			Tokens:               val.Tokens.String(),
			DelegatorShares:      val.DelegatorShares.String(),
			Moniker:              val.Description.Moniker,
			Identity:             val.Description.Identity,
			Website:              val.Description.Website,
			Details:              val.Description.Details,
			UnbondingHeight:      val.UnbondingHeight,
			UnbondingTime:        val.UnbondingTime,
			CommissionRate:       val.Commission.CommissionRates.Rate.String(),
			CommissionMaxRate:    val.Commission.CommissionRates.MaxRate.String(),
			CommissionChangeRate: val.Commission.CommissionRates.MaxChangeRate.String(),
			MinSelfDelegation:    val.MinSelfDelegation.String(),
			UpdateTime:           val.Commission.UpdateTime,
		}

		validators = append(validators, v)
	}

	return validators, nil
}

// GetValidatorsIdentities returns identities of all validators in the active chain
func (c *Client) GetValidatorsIdentities(vals []mdschema.Validator) (result []mdschema.Validator, err error) {
	for _, val := range vals {
		if val.Identity != "" {
			resp, err := c.KeyBase.R().Get("_/api/1.0/user/lookup.json?fields=pictures&key_suffix=" + val.Identity)
			if err != nil {
				return []mdschema.Validator{}, fmt.Errorf("failed to request identity: %s", err)
			}

			var keyBase mbltypes.KeyBase
			err = json.Unmarshal(resp.Body(), &keyBase)
			if err != nil {
				return []mdschema.Validator{}, fmt.Errorf("failed to unmarshal keybase: %s", err)
			}

			var url string
			if len(keyBase.Them) > 0 {
				for _, k := range keyBase.Them {
					url = k.Pictures.Primary.URL
				}

				v := mdschema.Validator{
					ID:         val.ID,
					KeybaseURL: url,
				}

				result = append(result, v)
			}
		}
	}

	return result, nil
}
