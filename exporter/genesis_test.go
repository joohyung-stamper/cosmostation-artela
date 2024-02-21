package exporter

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	//internal
	"github.com/CoreumFoundation/coreum/v3/app"
	"github.com/cosmostation/cosmostation-coreum/custom"
	mdschema "github.com/cosmostation/mintscan-database/schema"

	tmconfig "github.com/cometbft/cometbft/config"
	tmtypes "github.com/cometbft/cometbft/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestGetGenesisStateFromGenesisFile(t *testing.T) {
	var accounts []mdschema.AccountCoin
	// genesisFile := os.Getenv("PWD") + "/genesis.json"
	baseConfig := tmconfig.DefaultBaseConfig()
	genesisFile := filepath.Join(app.DefaultNodeHome, baseConfig.Genesis)
	// genesisFile := "/Users/jeonghwan/dev/cosmostation/cosmostation-coreum/genesis.json"
	log.Println("genesis file path :", genesisFile)
	genesisFile = "../ignoredir/cosmoshub-test-stargate-e.json"
	genDoc, err := tmtypes.GenesisDocFromFile(genesisFile)
	if err != nil {
		log.Println(err, "failed to read genesis doc file %s", genesisFile)
	}
	log.Println("genesis_time :", genDoc.GenesisTime)
	log.Println("chainid :", genDoc.ChainID)
	log.Println("initial_height :", genDoc.InitialHeight)

	var appState map[string]json.RawMessage
	if err = json.Unmarshal(genDoc.AppState, &appState); err != nil {
		log.Println(err, "failed to unmarshal genesis state")
	}
	// a := appState[authtypes.ModuleName]
	// log.Println(string(a)) //print message that key is auth {...}
	authGenesisState := authtypes.GetGenesisStateFromAppState(custom.AppCodec, appState)
	stakingGenesisState := stakingtypes.GetGenesisStateFromAppState(custom.AppCodec, appState)
	bondDenom := stakingGenesisState.Params.BondDenom
	lastValidatorPowers := stakingGenesisState.GetLastValidatorPowers()
	for _, val := range lastValidatorPowers {
		log.Println(val.Power)
	}
	os.Exit(0)
	var distributionGenesisState distributiontypes.GenesisState
	if appState[distributiontypes.ModuleName] != nil {
		custom.AppCodec.MustUnmarshalJSON(appState[distributiontypes.ModuleName], &distributionGenesisState)
		log.Println("abcded :", distributionGenesisState.DelegatorStartingInfos[0].DelegatorAddress)
		log.Println("counts :", len(distributionGenesisState.DelegatorStartingInfos))
	}
	// bondDenom := "uatom"

	authAccs := authGenesisState.GetAccounts()
	NumberOfTotalAccounts := len(authAccs)
	accountMapper := make(map[string]*mdschema.AccountCoin, NumberOfTotalAccounts)
	for _, authAcc := range authAccs {
		var ga authtypes.GenesisAccount
		custom.AppCodec.UnpackAny(authAcc, &ga)
		switch ga := ga.(type) {
		case *authtypes.BaseAccount:
		case *authvestingtypes.DelayedVestingAccount:
			/* Endtime 이 지난 vesting account 데이터는 의미 없다. */
			// ibc tokens은 delegate이 불가능하다 (stargate-5), bondDenom만 담는 것으로 하자.
			log.Printf("type %T\n", ga)
			log.Println("DelayedVestingAccount", ga.String())
			log.Println("delegated Free :", ga.GetDelegatedFree().AmountOf(bondDenom))
			log.Println("delegated vesting :", ga.GetDelegatedVesting().AmountOf(bondDenom))
			log.Println("vested coins:", ga.GetVestedCoins(time.Now()).AmountOf(bondDenom))    // 주어진 시간에 vesting이 풀린 코인
			log.Println("vesting coins :", ga.GetVestingCoins(time.Now()).AmountOf(bondDenom)) // 주어진 시간에 vesting 중인 코인
			log.Println("original vesting :", ga.GetOriginalVesting().AmountOf(bondDenom))
		case *authvestingtypes.ContinuousVestingAccount:
			log.Println("ContinuousVestingAccount", ga.String())
		case *authvestingtypes.PeriodicVestingAccount:
			log.Println("PeriodicVestingAccount", ga.String())
		}
		// log.Println(authAcc.GetTypeUrl())
		// log.Println(ga.GetAddress().String())
		// log.Println(ga.GetAccountNumber())
		sAcc := mdschema.AccountCoin{
			// ChainID:        genDoc.ChainID,
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
	balIter.IterateGenesisBalances(custom.AppCodec, appState,
		func(bal bankexported.GenesisBalance) (stop bool) {
			accAddress := bal.GetAddress()
			accCoins := bal.GetCoins()

			// accountMapper[accAddress.String()].CoinsSpendable = *accCoins.AmountOf(bondDenom).String()
			accountMapper[accAddress.String()].Available = accCoins.AmountOf(bondDenom).String()
			return false
		},
	)

	for _, acc := range accountMapper {
		accounts = append(accounts, *acc)
		// log.Println(acc)
	}

}

func TestExporterNil(t *testing.T) {
	s := new(mdschema.BasicData)
	log.Println(s)
	log.Println(s.Accounts)
	log.Println(len(s.Transactions))
	log.Println(s.Transactions)
}
