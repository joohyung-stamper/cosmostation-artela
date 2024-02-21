package exporter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// Staking module transaction hashes
	SampleMsgCreateValidatorTxHash = "710845805280376737CC95F59AA236ADAAF879EA69F06C10116B2AF0E8C62730"
	SampleMsgDelegateTxHash        = "2AFD81EED8DA2D29C8B042D3456C5409A49038F26BA01321CC48ED3D94E86E66"
	SampleMsgUndelegateTxHash      = "E268909F8D90E3316EE1DC2CCA136B2DF095E4591D59F544BD44E465C2B1D568"
	SampleMsgBeginRedelegateTxHash = "48B4298BFB9F37F20C8521F3A62E9D2465E794E0639189D3C633DACEDA4AC1CE"
)

func TestValidatorRank(t *testing.T) {
	ex.saveValidators()
}

func TestGetPowerEventHistory(t *testing.T) {

	b, err := ex.Client.RPC.GetBlock(5103)
	require.NoError(t, err)

	stdTx, err := ex.Client.CliCtx.GetTxs(b)
	require.NoError(t, err)
	// b.Block.Data.Txs
	// _, stdTx, err := commonTxParser(SampleMsgCreateValidatorTxHash)
	// require.NoError(t, err)

	peh, err := ex.getPowerEventHistoryNew(stdTx)
	require.NoError(t, err)

	for _, p := range peh {
		t.Log("height:", p.Height)
		t.Log("moniker:", p.Moniker)
		t.Log("operaddr:", p.OperatorAddress)
		t.Log("proposer:", p.Proposer)
		t.Log("txhash:", p.TxHash)
	}
}
