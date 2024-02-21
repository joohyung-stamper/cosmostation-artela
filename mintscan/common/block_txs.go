package common

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cosmostation/cosmostation-coreum/app"
	"github.com/cosmostation/cosmostation-coreum/custom"
	"github.com/cosmostation/cosmostation-coreum/errors"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	tmjson "github.com/cometbft/cometbft/libs/json"
)

type BlockTxs struct {
	Block json.RawMessage   `json:"block"`
	Txs   []json.RawMessage `json:"txs"`
}

func GetBlockTxs(a *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		heightStr := vars["height"]

		height, err := strconv.ParseInt(heightStr, 10, 64)
		if err != nil {
			zap.S().Debug("failed to parse int ", zap.Error(err))
			errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
			return
		}

		bt := BlockTxs{}
		// get block and transactions
		block, txs, err := a.Client.RPC.GetBlockAndTxsFromNode(custom.AppCodec, height)
		if err != nil {
			zap.S().Debug("failed to parse HTTP args ", zap.Error(err))
			errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
			return
		}

		// marshal
		blockChunk, err := tmjson.Marshal(block)
		if err != nil {
			zap.S().Debugf("failed to marshal block height = %d, err = %s\n", block.Block.Height, zap.Error(err))
			errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
			return
		}

		bt.Block = blockChunk

		for i := range txs {
			raw, err := custom.AppCodec.MarshalJSON(txs[i])
			if err != nil {
				zap.S().Debugf("failed to marshal tx hash = %s, err = %s\n", txs[i].TxHash, zap.Error(err))
				errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
				return
			}
			bt.Txs = append(bt.Txs, raw)
		}

		respond(rw, bt)
		return
	}
}
