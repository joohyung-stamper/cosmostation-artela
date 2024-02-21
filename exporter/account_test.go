package exporter

import (
	"context"
	"log"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmostation/cosmostation-coreum/custom"
)

func TestAccounts(t *testing.T) {

	if ex == nil {
		log.Println("ex nil")
		return
	}

	address := "cosmos1x5wgh6vwye60wv3dtshs9dmqggwfx2ldnqvev0"

	queryClient := authtypes.NewQueryClient(ex.Client.GetGRPCClient())

	accountResp, err := queryClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		log.Println(err)
	}

	acc := accountResp.GetAccount()
	var ai authtypes.AccountI
	custom.AppCodec.UnpackAny(acc, &ai)

	switch ai := ai.(type) {
	case *authtypes.ModuleAccount:
		log.Printf("type : %T", ai)
	case *authtypes.BaseAccount:
		log.Printf("type : %T", ai)
	case *authvestingtypes.DelayedVestingAccount:
		log.Printf("type : %T", ai)
	case *authvestingtypes.ContinuousVestingAccount:
		log.Printf("type : %T", ai)
	case *authvestingtypes.PeriodicVestingAccount:
		log.Printf("type : %T", ai)
	default:
		log.Printf("default type : %T", ai)
	}

}
