package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/CoreumFoundation/coreum/v3/app"
	"github.com/cosmostation/cosmostation-coreum/custom"
	mbltypes "github.com/cosmostation/mintscan-backend-library/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"
	"go.uber.org/zap"

	//cosmos-sdk
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	//tendermint
	tmconfig "github.com/cometbft/cometbft/config"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
)

const (
	// startingHeight is used to extract genesis accounts and parse their assets.
	startingHeight = int64(1)
)

// GetGenesisStateFromGenesisFile get the genesis account information from genesis state ({NODE_HOME}/config/Genesis.json)
func (ex *Exporter) GetGenesisStateFromGenesisFile(genesisPath string) (err error) {
	// genesisFile := os.Getenv("PWD") + "/genesis.json"
	baseConfig := tmconfig.DefaultBaseConfig()
	genesisFile := filepath.Join(app.DefaultNodeHome, baseConfig.Genesis)

	if genesisPath == "" {
		genesisPath = genesisFile
	}
	// genesisFile := "/Users/jeonghwan/dev/cosmostation/cosmostation-coreum/genesis.json"
	genDoc, err := tmtypes.GenesisDocFromFile(genesisPath)
	if err != nil {
		log.Println(err, "failed to read genesis doc file %s", genesisPath)
		return
	}

	var genesisState map[string]json.RawMessage
	if err = json.Unmarshal(genDoc.AppState, &genesisState); err != nil {
		log.Println(err, "failed to unmarshal genesis state")
		return
	}
	// a := genesisState[authtypes.ModuleName]
	// log.Println(string(a)) //print message that key is auth {...}
	authGenesisState := authtypes.GetGenesisStateFromAppState(custom.AppCodec, genesisState)
	stakingGenesisState := stakingtypes.GetGenesisStateFromAppState(custom.AppCodec, genesisState)
	bondDenom := stakingGenesisState.GetParams().BondDenom

	authAccs := authGenesisState.GetAccounts()
	NumberOfTotalAccounts := len(authAccs)
	accountMapper := make(map[string]*mdschema.AccountCoin, NumberOfTotalAccounts)
	for _, authAcc := range authAccs {
		var ga authtypes.GenesisAccount
		custom.AppCodec.UnpackAny(authAcc, &ga)
		switch ga := ga.(type) {
		case *authtypes.BaseAccount:
		case *authvestingtypes.DelayedVestingAccount:
			log.Println("DelayedVestingAccount", ga.String())
			log.Println("delegated Free :", ga.GetDelegatedFree())
			log.Println("delegated vesting :", ga.GetDelegatedVesting())
			log.Println("vested coins:", ga.GetVestedCoins(time.Now()))
			log.Println("vesting coins :", ga.GetVestingCoins(time.Now()))
			log.Println("original vesting :", ga.GetOriginalVesting())
		case *authvestingtypes.ContinuousVestingAccount:
			log.Println("ContinuousVestingAccount", ga.String())
		case *authvestingtypes.PeriodicVestingAccount:
			log.Println("PeriodicVestingAccount", ga.String())
		}
		sAcc := mdschema.AccountCoin{
			// ChainID:           genDoc.ChainID,
			Address: ga.GetAddress().String(),
			// AccountNumber:  ga.GetAccountNumber(), //account number is set by specified order in genesis file
			// AccountType:    authAcc.GetTypeUrl(),  //type 변경
			Total:        "0",
			Available:    "0",
			Delegated:    "0",
			Rewards:      "0",
			Commission:   "0",
			Undelegated:  "0",
			FailedVested: "0",
			Vested:       "0",
			Vesting:      "0",
			// CreationTime: genDoc.GenesisTime.String(),
		}
		accountMapper[ga.GetAddress().String()] = &sAcc
	}

	balIter := banktypes.GenesisBalancesIterator{}
	balIter.IterateGenesisBalances(custom.AppCodec, genesisState,
		func(bal bankexported.GenesisBalance) (stop bool) {
			accAddress := bal.GetAddress()
			accCoins := bal.GetCoins()

			// accountMapper[accAddress.String()].CoinsSpendable = *accCoins.AmountOf(bondDenom).BigInt()
			accountMapper[accAddress.String()].Available = accCoins.AmountOf(bondDenom).String()
			return false
		},
	)

	var accounts []mdschema.AccountCoin
	for _, acc := range accountMapper {
		accounts = append(accounts, *acc)
		log.Println(acc)
	}

	ex.DB.InsertGenesisAccount(accounts)

	return
}

// deprecated
func (ex *Exporter) getGenesisAccounts(genesisAccts authtypes.GenesisAccounts) (accounts []mdschema.AccountCoin, err error) {
	// chainID, err := ex.Client.GetNetworkChainID()
	if err != nil {
		return []mdschema.AccountCoin{}, err
	}

	block, err := ex.Client.RPC.GetBlock(startingHeight)
	if err != nil {
		return []mdschema.AccountCoin{}, err
	}

	denom, err := ex.Client.GRPC.GetBondDenom(context.Background())
	if err != nil {
		return []mdschema.AccountCoin{}, err
	}

	for i, account := range genesisAccts {
		switch account := account.(type) {
		case *authtypes.BaseAccount:
			zap.S().Infof("Account type: %T | Synced account %d/%d", account, i, len(genesisAccts))

			// acc := account.(*authtypes.BaseAccount)

			spendable, rewards, commission, delegated, undelegated, err := ex.Client.GetBaseAccountTotalAsset(account.GetAddress().String())
			if err != nil {
				return []mdschema.AccountCoin{}, err
			}

			total := sdktypes.NewCoin(denom, sdktypes.NewInt(0))

			// Sum up all coins that exist in an account.
			total = total.Add(spendable).
				Add(delegated).
				Add(undelegated).
				Add(rewards).
				Add(commission)

			acct := mdschema.AccountCoin{
				// ChainID:          chainID,
				Address: account.Address,
				// AccountNumber:    acc.AccountNumber,
				// AccountType:      types.BaseAccount,
				Total:       total.Amount.String(),
				Available:   spendable.Amount.String(),
				Rewards:     rewards.Amount.String(),
				Commission:  commission.Amount.String(),
				Delegated:   delegated.Amount.String(),
				Undelegated: undelegated.Amount.String(),
			}

			accounts = append(accounts, acct)

		case *authtypes.ModuleAccount:
			zap.S().Infof("Account type: %T | Synced account %d/%d", account, i, len(genesisAccts))

			// acc := account.(authtypes.ModuleAccountI)

			spendable, rewards, commission, delegated, undelegated, err := ex.Client.GetBaseAccountTotalAsset(account.GetAddress().String())
			if err != nil {
				return []mdschema.AccountCoin{}, err
			}

			total := sdktypes.NewCoin(denom, sdktypes.NewInt(0))

			// Sum up all coins that exist in an account.
			total = total.Add(spendable).
				Add(delegated).
				Add(undelegated).
				Add(rewards).
				Add(commission)

			acct := mdschema.AccountCoin{
				// ChainID:          chainID,
				Address: account.GetAddress().String(),
				// AccountNumber:    account.GetAccountNumber(),
				// AccountType:      types.ModuleAccount,
				Total:       total.Amount.String(),
				Available:   spendable.Amount.String(),
				Rewards:     rewards.Amount.String(),
				Commission:  commission.Amount.String(),
				Delegated:   delegated.Amount.String(),
				Undelegated: undelegated.Amount.String(),
			}

			accounts = append(accounts, acct)

		case *authvestingtypes.PeriodicVestingAccount:
			zap.S().Infof("Account type: %T | Synced account %d/%d", account, i, len(genesisAccts))

			// acc := account.(*authvestingtypes.PeriodicVestingAccount)

			spendable, rewards, commission, delegated, undelegated, err := ex.Client.GetBaseAccountTotalAsset(account.GetAddress().String())
			if err != nil {
				return []mdschema.AccountCoin{}, err
			}

			vesting := sdktypes.NewCoin(denom, sdktypes.NewInt(0))
			vested := sdktypes.NewCoin(denom, sdktypes.NewInt(0))

			vestingCoins := account.GetVestingCoins(block.Block.Time)
			vestedCoins := account.GetVestedCoins(block.Block.Time)
			delegatedVesting := account.GetDelegatedVesting()

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

			total := sdktypes.NewCoin(denom, sdktypes.NewInt(0))

			// Sum up all coins that exist in an account.
			total = total.Add(spendable).
				Add(delegated).
				Add(undelegated).
				Add(rewards).
				Add(commission).
				Add(vesting)

			acct := mdschema.AccountCoin{
				// ChainID:          chainID,
				Address: account.Address,
				// AccountNumber:    account.AccountNumber,
				// AccountType:      types.PeriodicVestingAccount,
				Total:       total.Amount.String(),
				Available:   spendable.Amount.String(),
				Rewards:     rewards.Amount.String(),
				Commission:  commission.Amount.String(),
				Delegated:   delegated.Amount.String(),
				Undelegated: undelegated.Amount.String(),
				// CreationTime:     block.Block.Time.String(),
			}

			accounts = append(accounts, acct)

		default:
			return []mdschema.AccountCoin{}, fmt.Errorf("unrecognized account type: %T", account)
		}
	}

	return accounts, nil
}

// getGenesisValidatorsSet returns validator set in genesis.
func (ex *Exporter) getGenesisValidatorsSet(block *tmctypes.ResultBlock, vals *tmctypes.ResultValidators) ([]mdschema.PowerEventHistory, error) {
	// Get genesis validator set (block height 1).
	if block.Block.Height != 1 {
		return []mdschema.PowerEventHistory{}, nil
	}

	denom, err := ex.Client.GRPC.GetBondDenom(context.Background())
	if err != nil {
		return []mdschema.PowerEventHistory{}, err
	}

	if vals == nil {
		return []mdschema.PowerEventHistory{}, nil
	}
	genesisValsSet := make([]mdschema.PowerEventHistory, 0)
	for i, val := range vals.Validators {
		gvs := mdschema.PowerEventHistory{
			IDValidator:          i + 1,
			Height:               block.Block.Height,
			Moniker:              "",
			OperatorAddress:      "",
			Proposer:             val.Address.String(),
			VotingPower:          float64(val.VotingPower),
			MsgType:              mbltypes.StakingMsgCreateValidator,
			NewVotingPowerAmount: float64(val.VotingPower),
			NewVotingPowerDenom:  denom,
			TxHash:               "",
			Timestamp:            block.Block.Header.Time,
		}

		genesisValsSet = append(genesisValsSet, gvs)
	}

	return genesisValsSet, nil
}
