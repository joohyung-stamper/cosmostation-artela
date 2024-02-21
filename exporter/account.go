package exporter

import (
	"context"
	"fmt"

	// cosmos-sdk
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// mbl
	mbltypes "github.com/cosmostation/mintscan-backend-library/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"

	// tendermint
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"

	"go.uber.org/zap"
)

// getAccounts
func (ex *Exporter) getAccounts(block *tmctypes.ResultBlock, txResps []*sdk.TxResponse) (accounts []mdschema.AccountCoin, err error) {
	if len(txResps) <= 0 {
		return []mdschema.AccountCoin{}, nil
	}

	for _, txResp := range txResps {
		// Other than code equals to 0, it is failed transaction.
		if txResp.Code != 0 {
			return []mdschema.AccountCoin{}, nil
		}

		// stdTx, ok := tx.Tx.(auth.StdTx)
		// if !ok {
		// 	return []schema.AccountCoin{}, fmt.Errorf("unsupported tx type: %s", tx.Tx)
		// }

		msgs := txResp.GetTx().GetMsgs()

		for _, msg := range msgs {
			/*
				tx 내 많은 메세지가 존재하고, 동일한 어카운트 대한 조회가 반복적으로 이루어질 수 있음.
				따라서 tx내 모든 메세지를 파싱하고 난 뒤 중복된 어카운트를 제거하여 1회만 계정별 조회를 할 필요가 있음
			*/
			switch m := msg.(type) {
			case *banktypes.MsgSend:
				zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), txResp.TxHash)

				// msgSend := m.(bank.MsgSend)

				fromAcct, err := ex.Client.CliCtx.GetAccount(m.FromAddress)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				toAcct, err := ex.Client.CliCtx.GetAccount(m.ToAddress)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				exportedAccts := []sdkclient.Account{
					fromAcct, toAcct,
				}

				accounts, err = ex.getAccountAllAssets(exportedAccts, txResp.TxHash, txResp.Timestamp)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

			case *banktypes.MsgMultiSend:
				zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), txResp.TxHash)

				// msgMultiSend := m.(bank.MsgMultiSend)

				var exportedAccts []sdkclient.Account

				for _, input := range m.Inputs {
					inputAcct, err := ex.Client.CliCtx.GetAccount(input.Address)
					if err != nil {
						return []mdschema.AccountCoin{}, err
					}

					exportedAccts = append(exportedAccts, inputAcct)
				}

				for _, output := range m.Outputs {
					outputAcct, err := ex.Client.CliCtx.GetAccount(output.Address)
					if err != nil {
						return []mdschema.AccountCoin{}, err
					}

					exportedAccts = append(exportedAccts, outputAcct)
				}

				accounts, err = ex.getAccountAllAssets(exportedAccts, txResp.TxHash, txResp.Timestamp)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

			case *stakingtypes.MsgDelegate:
				zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), txResp.TxHash)

				// msgDelegate := m.(staking.MsgDelegate)

				delegatorAddr, err := ex.Client.CliCtx.GetAccount(m.DelegatorAddress)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				valAccAddr, err := mbltypes.ConvertAccAddrFromValAddr(m.DelegatorAddress)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				valAddr, err := ex.Client.CliCtx.GetAccount(valAccAddr)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				exportedAccts := []sdkclient.Account{
					delegatorAddr, valAddr,
				}

				accounts, err = ex.getAccountAllAssets(exportedAccts, txResp.TxHash, txResp.Timestamp)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

			case *stakingtypes.MsgUndelegate:
				zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), txResp.TxHash)

				// msgUndelegate := m.(staking.MsgUndelegate)

				delegatorAddr, err := ex.Client.CliCtx.GetAccount(m.DelegatorAddress)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				valAccAddr, err := mbltypes.ConvertAccAddrFromValAddr(m.DelegatorAddress)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				valAddr, err := ex.Client.CliCtx.GetAccount(valAccAddr)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				exportedAccts := []sdkclient.Account{
					delegatorAddr, valAddr,
				}

				accounts, err = ex.getAccountAllAssets(exportedAccts, txResp.TxHash, txResp.Timestamp)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

			case *stakingtypes.MsgBeginRedelegate:
				zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), txResp.TxHash)

				// msgBeginRedelegate := m.(staking.MsgBeginRedelegate)

				delegatorAddr, err := ex.Client.CliCtx.GetAccount(m.DelegatorAddress)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				valSrcAccAddr, err := mbltypes.ConvertAccAddrFromValAddr(m.ValidatorSrcAddress)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				valDstAccAddr, err := mbltypes.ConvertAccAddrFromValAddr(m.ValidatorDstAddress)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				srcAddr, err := ex.Client.CliCtx.GetAccount(valSrcAccAddr)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				dstAddr, err := ex.Client.CliCtx.GetAccount(valDstAccAddr)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

				exportedAccts := []sdkclient.Account{
					delegatorAddr, srcAddr, dstAddr,
				}

				accounts, err = ex.getAccountAllAssets(exportedAccts, txResp.TxHash, txResp.Timestamp)
				if err != nil {
					return []mdschema.AccountCoin{}, err
				}

			default:
				continue
			}
		}
	}

	return accounts, nil
}

func (ex *Exporter) getAccountAllAssets(exportedAccts []sdkclient.Account, txHashStr, txTime string) (accounts []mdschema.AccountCoin, err error) {
	// chainID, err := ex.Client.GetNetworkChainID()
	if err != nil {
		return []mdschema.AccountCoin{}, err
	}

	denom, err := ex.Client.GRPC.GetBondDenom(context.Background())
	if err != nil {
		return []mdschema.AccountCoin{}, err
	}

	latestBlockHeight, err := ex.Client.RPC.GetLatestBlockHeight()
	if err != nil {
		return []mdschema.AccountCoin{}, err
	}

	block, err := ex.Client.RPC.GetBlock(latestBlockHeight)
	if err != nil {
		return []mdschema.AccountCoin{}, err
	}

	for _, account := range exportedAccts {
		switch acc := account.(type) {
		case *authtypes.BaseAccount:
			zap.S().Infof("Account type: %T | Account: %s", acc, account.GetAddress())

			// acc := account.(*authtypes.BaseAccount)

			available, rewards, commission, delegated, undelegated, err := ex.Client.GetBaseAccountTotalAsset(acc.GetAddress().String())
			if err != nil {
				return []mdschema.AccountCoin{}, err
			}

			total := sdk.NewCoin(denom, sdk.NewInt(0))

			// Sum up all coins that exist in an account.
			total = total.Add(available).
				Add(delegated).
				Add(undelegated).
				Add(rewards).
				Add(commission)

			acct := mdschema.AccountCoin{
				// ChainID:           chainID,
				Address: acc.Address,
				// AccountNumber:     acc.AccountNumber,
				// AccountType:       types.BaseAccount,
				Denom:        denom,
				Total:        total.Amount.String(),
				Available:    available.Amount.String(),
				Rewards:      rewards.Amount.String(),
				Commission:   commission.Amount.String(),
				Delegated:    delegated.Amount.String(),
				Undelegated:  undelegated.Amount.String(),
				Vested:       "0",
				Vesting:      "0",
				FailedVested: "0",
				LastTx:       txHashStr,
				LastTxTime:   txTime,
				// CreationTime:      block.Block.Time.String(),
			}

			accounts = append(accounts, acct)

		case *authtypes.ModuleAccount:
			zap.S().Infof("Account type: %T | Account: %s", acc, account.GetAddress())

			// acc := account.(authtypes.ModuleAccountI)

			available, rewards, commission, delegated, undelegated, err := ex.Client.GetBaseAccountTotalAsset(acc.GetAddress().String())
			if err != nil {
				return []mdschema.AccountCoin{}, err
			}

			total := sdk.NewCoin(denom, sdk.NewInt(0))

			// Sum up all coins that exist in an account.
			total = total.Add(available).
				Add(delegated).
				Add(undelegated).
				Add(rewards).
				Add(commission)

			acct := mdschema.AccountCoin{
				// ChainID:           chainID,
				Address: acc.GetAddress().String(),
				// AccountNumber:     acc.GetAccountNumber(),
				// AccountType:       types.ModuleAccount,
				Denom:        denom,
				Total:        total.Amount.String(),
				Available:    available.Amount.String(),
				Rewards:      rewards.Amount.String(),
				Commission:   commission.Amount.String(),
				Delegated:    delegated.Amount.String(),
				Undelegated:  undelegated.Amount.String(),
				Vested:       "0",
				Vesting:      "0",
				FailedVested: "0",
				LastTx:       txHashStr,
				LastTxTime:   txTime,
				// CreationTime: block.Block.Time.String(),
			}

			accounts = append(accounts, acct)

		case *authvestingtypes.PeriodicVestingAccount:
			zap.S().Infof("Account type: %T | Account: %s", acc, account.GetAddress())

			// acc := account.(*authvestingtypes.PeriodicVestingAccount)

			available, rewards, commission, delegated, undelegated, err := ex.Client.GetBaseAccountTotalAsset(acc.GetAddress().String())
			if err != nil {
				return []mdschema.AccountCoin{}, err
			}

			vesting := sdk.NewCoin(denom, sdk.NewInt(0))
			vested := sdk.NewCoin(denom, sdk.NewInt(0))

			vestingCoins := acc.GetVestingCoins(block.Block.Time)
			vestedCoins := acc.GetVestedCoins(block.Block.Time)
			delegatedVesting := acc.GetDelegatedVesting()

			// When total vesting amount is greater than or equal to delegated vesting amount, then
			// there is still a room to delegate. Otherwise, vesting should be zero.
			if len(vestingCoins) > 0 {
				if vestingCoins.IsAllGTE(delegatedVesting) {
					vestingCoins = vestingCoins.Sub(delegatedVesting...)
					for _, vc := range vestingCoins {
						if vc.Denom == denom {
							vesting = vesting.Add(vc)
						}
					}
				}
			}

			if len(vestedCoins) > 0 {
				for _, vc := range vestedCoins {
					if vc.Denom == denom {
						vested = vested.Add(vc)
					}
				}
			}

			total := sdk.NewCoin(denom, sdk.NewInt(0))

			// Sum up all coins that exist in an account.
			total = total.Add(available).
				Add(delegated).
				Add(undelegated).
				Add(rewards).
				Add(commission).
				Add(vesting)

			acct := mdschema.AccountCoin{
				// ChainID:           chainID,
				Address: acc.Address,
				// AccountNumber:     acc.AccountNumber,
				// AccountType:       types.PeriodicVestingAccount,
				Denom:        denom,
				Total:        total.Amount.String(),
				Available:    available.Amount.String(),
				Rewards:      rewards.Amount.String(),
				Commission:   commission.Amount.String(),
				Delegated:    delegated.Amount.String(),
				Undelegated:  undelegated.Amount.String(),
				Vested:       "0",
				Vesting:      "0",
				FailedVested: "0",
				LastTx:       txHashStr,
				LastTxTime:   txTime,
				// CreationTime:      block.Block.Time.String(),
			}

			accounts = append(accounts, acct)

		case *authvestingtypes.DelayedVestingAccount:
			zap.S().Infof("Account type: %T | Account: %s", acc, account.GetAddress())

			// acc := account.(*authvestingtypes.DelayedVestingAccount)

			available, rewards, commission, delegated, undelegated, err := ex.Client.GetBaseAccountTotalAsset(acc.GetAddress().String())
			if err != nil {
				return []mdschema.AccountCoin{}, err
			}

			vesting := sdk.NewCoin(denom, sdk.NewInt(0))
			vested := sdk.NewCoin(denom, sdk.NewInt(0))

			vestingCoins := acc.GetVestingCoins(block.Block.Time)
			vestedCoins := acc.GetVestedCoins(block.Block.Time)
			delegatedVesting := acc.GetDelegatedVesting()

			// When total vesting amount is greater than or equal to delegated vesting amount, then
			// there is still a room to delegate. Otherwise, vesting should be zero.
			if len(vestingCoins) > 0 {
				if vestingCoins.IsAllGTE(delegatedVesting) {
					vestingCoins = vestingCoins.Sub(delegatedVesting...)
					for _, vc := range vestingCoins {
						if vc.Denom == denom {
							vesting = vesting.Add(vc)
						}
					}
				}
			}

			if len(vestedCoins) > 0 {
				for _, vc := range vestedCoins {
					if vc.Denom == denom {
						vested = vested.Add(vc)
					}
				}
			}

			total := sdk.NewCoin(denom, sdk.NewInt(0))

			// Sum up all coins that exist in an account.
			total = total.Add(available).
				Add(delegated).
				Add(undelegated).
				Add(rewards).
				Add(commission).
				Add(vesting)

			acct := mdschema.AccountCoin{
				// ChainID:           chainID,
				Address: acc.Address,
				// AccountNumber:     acc.AccountNumber,
				// AccountType:       types.DelayedVestingAccount,
				Denom:        denom,
				Total:        total.Amount.String(),
				Available:    available.Amount.String(),
				Rewards:      rewards.Amount.String(),
				Commission:   commission.Amount.String(),
				Delegated:    delegated.Amount.String(),
				Undelegated:  undelegated.Amount.String(),
				Vested:       "0",
				Vesting:      "0",
				FailedVested: "0",
				LastTx:       txHashStr,
				LastTxTime:   txTime,
				// CreationTime: block.Block.Time.String(),
			}

			accounts = append(accounts, acct)

		default:
			return []mdschema.AccountCoin{}, fmt.Errorf("unrecognized account type: %T", account)
		}
	}

	return accounts, nil
}
