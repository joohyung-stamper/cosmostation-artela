package exporter

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	//internal
	"github.com/cosmostation/cosmostation-coreum/custom"

	//cosmos-sdk
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdktypestx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/require"
)

// TestGetTxsChunk decodes transactions in a block and return a format of database transaction.
func TestGetTxsChunk(t *testing.T) {
	require.NotNil(t, ex.Client)
	// 13030, 272247
	// 122499 (multi msg type)
	block, txResps, err := ex.Client.RPC.GetBlockAndTxsFromNode(custom.AppCodec, 3316106)
	require.NoError(t, err)
	_ = block
	tma := ex.disassembleTransaction(txResps)
	log.Println(tma)

	return
}

func TestMsgParsing(t *testing.T) {
	require.NotNil(t, ex.Client)

	block, txResps, err := ex.Client.RPC.GetBlockAndTxsFromNode(custom.AppCodec, 3344967)
	// block, txResps, err := ex.Client.RPC.GetBlockAndTxsFromNode(custom.AppCodec, 3354129)
	require.NoError(t, err)
	_ = block
	tma := ex.disassembleTransaction(txResps)
	log.Println(tma)
}

func InsertJSONStringToDB(txResps []*sdktypes.TxResponse) ([]string, error) {
	jsonString := make([]string, len(txResps), len(txResps))
	for i, txResp := range txResps {
		chunk, err := custom.AppCodec.MarshalJSON(txResp)
		if err != nil {
			log.Println(err)
		}
		jsonString[i] = string(chunk)
		// show result
		fmt.Println(jsonString[i])
	}

	return jsonString, nil
}

func JSONStringUnmarshal(jsonString []string) error {
	txResps := make([]sdktypes.TxResponse, len(jsonString), len(jsonString))
	for i, js := range jsonString {
		err := custom.AppCodec.UnmarshalJSON([]byte(js), &txResps[i])
		if err != nil {
			log.Println(err)
			return err
		}
		// show result
		fmt.Println("decode:", txResps[i].String())
	}

	return nil
}

func TestGetMessage(t *testing.T) {
	// 13030, 272247
	// 122499 (multi msg type)
	block, err := ex.Client.RPC.GetBlock(970957)
	if err != nil {
		t.Log(err)
	}
	txResps, err := ex.Client.CliCtx.GetTxs(block)
	if err != nil {
		t.Log(err)
	}

	for _, txResp := range txResps {
		txI := txResp.GetTx()
		tx, ok := txI.(*sdktypestx.Tx)
		if !ok {
			return
		}
		getMessages := tx.GetBody().GetMessages()
		msgjson := make([]json.RawMessage, len(getMessages), len(getMessages))
		var err error
		for i, msg := range getMessages {
			msgjson[i], err = custom.AppCodec.MarshalJSON(msg)
			if err != nil {
				t.Log(err)
				return
			}
		}
		jsonraws, err := json.Marshal(msgjson)
		t.Log(string(jsonraws))
	}

	return
}

func TestUnmarshalMessageString(t *testing.T) {
	msgStr := "[{\"@type\": \"/cosmos.staking.v1beta1.MsgDelegate\", \"amount\": {\"denom\": \"umuon\", \"amount\": \"18044801\"}, \"delegator_address\": \"cosmos10fyfu7fl78f88a7zhcwu72wk3hjlzdm83yr09k\", \"validator_address\": \"cosmosvaloper10fyfu7fl78f88a7zhcwu72wk3hjlzdm85sh6f9\"}]"

	var jsonRaws []json.RawMessage
	json.Unmarshal([]byte(msgStr), &jsonRaws)

	for _, raw := range jsonRaws {
		t.Log(string(raw))
		var any codectypes.Any
		custom.AppCodec.UnmarshalJSON(raw, &any)
		t.Log(any.TypeUrl)
		// any.GetCachedValue().(type)
		t.Log(any.GetCachedValue())
		b, err := json.Marshal(any)
		require.NoError(t, err)
		t.Log(string(any.Value))

		t.Log(string(b))
	}

}

// func TestMap(t *testing.T) {
// 	m := make(map[string]struct{})

// 	key1 := ""
// 	key2 := "abcd"

// 	m[key1] = struct{}{}
// 	m[key2] = struct{}{}

// 	for k, v := range m {
// 		t.Log("key :", k, " value :", v)
// 	}
// }

func TestGetBlockandTx(t *testing.T) {
	h := int64(696591)
	// b, txs, err := ex.Client.RPC.GetBlockAndTxsFromNode(custom.EncodingConfig.Marshaler, h)
	b, txs, err := ex.Client.RPC.GetBlockAndTxsFromNode(custom.EncodingConfig.Codec, h)
	require.NoError(t, err)

	t.Log("height =", b.Block.Height)
	for i := range txs {
		t.Log(txs[i].TxHash)
	}

}

func TestICA(t *testing.T) {
	msgType := "ibcchannel/recv_packet"
	beginID := int64(0)
	// singleTx, err := ex.DB.GetTransactionByHash("EB5CCDC5CF23595FB62D2D4904A3EAF19814FF36C4164FD9CCDD4FF94EDA56A5")
	// require.NoError(t, err)
	// t.Log(singleTx.Hash)
	// singleTxResp := sdktypes.TxResponse{}
	// err = custom.AppCodec.UnmarshalJSON(singleTx.Chunk, &singleTxResp)
	// require.NoError(t, err)
	// msgs := singleTxResp.GetTx().GetMsgs()
	// for j := range msgs {
	// 	custom.AccountExporterFromIBCMsg(&msgs[j], singleTx.Hash)
	// }
	// os.Exit(0)

	for {
		txResps, err := ex.DB.GetTransactionsByMsgType(beginID, msgType, 50)
		require.NoError(t, err)

		if len(txResps) == 0 {
			break
		}

		for i := range txResps {
			txResp := sdktypes.TxResponse{}
			custom.AppCodec.UnmarshalJSON(txResps[i].Chunk, &txResp)
			msgs := txResp.GetTx().GetMsgs()
			for j := range msgs {
				custom.AccountExporterFromIBCMsg(&msgs[j], txResps[i].Hash)
			}

			beginID = txResps[i].ID + 1
		}
		t.Log("Next begin : ", beginID)
	}
}
