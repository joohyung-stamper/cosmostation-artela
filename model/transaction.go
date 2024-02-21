package model

import (
	"encoding/json"
	"time"

	"github.com/cosmostation/cosmostation-coreum/app"
	mdschema "github.com/cosmostation/mintscan-database/schema"
)

// TxData defines the structure for transction data list.
// type TxData struct {
// 	Txs []json.RawMessage `json:"txs"`
// }

// TxList defines the structure for transaction list.
type TxList struct {
	TxHash []string `json:"tx_list"`
}

// Message defines the structure for transaction message.
// type Message struct {
// 	Type  string          `json:"type"`
// 	Value json.RawMessage `json:"value"`
// }

// Fee defines the structure for transaction fee.
// type Fee struct {
// 	Gas    string `json:"gas,omitempty"`
// 	Amount []struct {
// 		Amount string `json:"amount,omitempty"`
// 		Denom  string `json:"denom,omitempty"`
// 	} `json:"amount,omitempty"`
// }

// Event defines the structure for transaction event.
// type Event struct {
// 	Type       string `json:"type"`
// 	Attributes []struct {
// 		Key   string `json:"key"`
// 		Value string `json:"value"`
// 	} `json:"attributes"`
// }

// Log defines the structure for transaction log.
// type Log struct {
// 	MsgIndex int     `json:"msg_index"`
// 	Log      string  `json:"log"`
// 	Events   []Event `json:"events"`
// }

// ParseTransaction receives single transaction from database and return it after unmarshal them.
func ParseTransaction(a *app.App, tx mdschema.Transaction) (result *ResultTx) {
	var jsonRaws json.RawMessage

	jsonRaws = tx.Chunk

	if tx.ID != 0 {
		header := ResultTxHeader{
			ID:        tx.ID,
			ChainID:   a.ChainNumMap[tx.ChainInfoID],
			BlockID:   tx.BlockID,
			Timestamp: tx.Timestamp.Format(time.RFC3339),
		}

		result = &ResultTx{
			ResultTxHeader: header,
			Data:           jsonRaws,
		}
	}

	return result
}

// ParseTransactions receives result transactions from database and return them after unmarshal them.
func ParseTransactions(a *app.App, txs []mdschema.Transaction) (results []*ResultTx) {
	for i := range txs {
		results = append(results, ParseTransaction(a, txs[i]))
	}
	return results
}
