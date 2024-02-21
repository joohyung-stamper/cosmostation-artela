package exporter

import (
	"fmt"
	"log"
	"time"

	// internal
	"github.com/cosmostation/cosmostation-coreum/custom"

	// core
	mbltypes "github.com/cosmostation/mintscan-backend-library/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"

	// sdk
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// getTxs decodes transactions in a block and return a format of database transaction.
func (ex *Exporter) getTxs(chainID string, list map[int64]*mdschema.Block, txResp []*sdktypes.TxResponse, useList bool) ([]mdschema.Transaction, error) {
	txs := make([]mdschema.Transaction, 0)

	if len(txResp) <= 0 {
		return txs, nil
	}

	codec := custom.AppCodec

	for i := range txResp {

		chunk, err := codec.MarshalJSON(txResp[i])
		if err != nil {
			return txs, fmt.Errorf("failed to marshal tx : %s", err)
		}

		ts, err := time.Parse(time.RFC3339Nano, txResp[i].Timestamp)
		if err != nil {
			return txs, fmt.Errorf("failed to parse timestamp : %s", err)
		}

		var blockID int64
		if useList {
			blockID = list[txResp[i].Height].ID
		}

		t := mdschema.Transaction{
			ChainInfoID: ex.ChainIDMap[chainID],
			BlockID:     blockID,
			Height:      txResp[i].Height,
			Code:        txResp[i].Code,
			Hash:        txResp[i].TxHash,
			Chunk:       chunk,
			Timestamp:   ts,
		}

		txs = append(txs, t)
	}

	return txs, nil
}

// getTxsChunk decodes transactions in a block and return a format of database transaction.
func (ex *Exporter) getRawTransactions(block *tmctypes.ResultBlock, txResps []*sdktypes.TxResponse) ([]mdschema.RawTransaction, error) {
	txChunk := make([]mdschema.RawTransaction, len(txResps), len(txResps))
	if len(txResps) <= 0 {
		return txChunk, nil
	}

	for i, txResp := range txResps {
		chunk, err := custom.AppCodec.MarshalJSON(txResp)
		if err != nil {
			log.Println(err)
			return txChunk, fmt.Errorf("failed to marshal tx : %s", err)
		}
		txChunk[i].ChainID = block.Block.ChainID
		txChunk[i].Height = txResp.Height
		txChunk[i].TxHash = txResp.TxHash
		txChunk[i].Chunk = chunk
	}

	return txChunk, nil
}

func (ex *Exporter) disassembleTransaction(txResps []*sdktypes.TxResponse) (uniqTransactionMessageAccounts []mdschema.TMA) {
	if len(txResps) <= 0 {
		return nil
	}

	for _, txResp := range txResps {
		msgs := txResp.GetTx().GetMsgs()

		txHash := txResp.TxHash

		uniqueMsgAccount := make(map[string]map[string]struct{}) // tx 내 동일 메세지에 대한 유일한 어카운트 저장

		for _, msg := range msgs {

			msgType, accounts := mbltypes.AccountExporterFromCosmosTxMsg(&msg)
			// 어떤 msg 타입에 대해서도 signer를 이용해 accounts를 확보하면, 모든 메세지를 파싱할 수 있다.
			signers := getSignerAddress(msg.GetSigners())
			accounts = append(accounts, signers...)

			for _, txParser := range custom.CustomTxParsers {
				if msgType != "" {
					break
				}
				customMsgType, account := txParser(&msg, txHash)
				msgType = customMsgType
				accounts = append(accounts, account...)
			}

			if msgType == "" {
				// msgType 이 없을 경우, 해당 건은 수집하지 않는다.
				continue
			}
			for i := range accounts {
				ma, ok := uniqueMsgAccount[msgType]
				if !ok {
					ma = make(map[string]struct{})
					uniqueMsgAccount[msgType] = ma
				}
				ma[accounts[i]] = struct{}{}
			}
		} // end msgs for loop

		// msg 별 유일 어카운트 수집
		tma := parseTransactionMessageAccount(txHash, uniqueMsgAccount, txResp.Height)
		uniqTransactionMessageAccounts = append(uniqTransactionMessageAccounts, tma...)
	} // 모든 tx 완료

	return uniqTransactionMessageAccounts
}

// msg - account 매핑 unique
func parseTransactionMessageAccount(txHash string, msgAccount map[string]map[string]struct{}, height int64) []mdschema.TMA {
	tma := make([]mdschema.TMA, 0)
	for msg := range msgAccount {
		for acc := range msgAccount[msg] {
			ta := mdschema.TMA{
				TxHash:         txHash,
				MsgType:        msg,
				AccountAddress: acc,
				Height:         height,
			}
			tma = append(tma, ta)
		}
	}
	return tma
}

func getSignerAddress(accAddrs []sdktypes.AccAddress) (address []string) {
	for _, addr := range accAddrs {
		if addr.String() != "" {
			address = append(address, addr.String())
		}
	}

	return address
}
