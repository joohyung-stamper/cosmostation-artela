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
)

type Txs struct {
	Txs []json.RawMessage `json:"txs"`
}

func GetTxs(a *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		heightStr := vars["height"]

		height, err := strconv.ParseInt(heightStr, 10, 64)
		if err != nil {
			zap.S().Debug("failed to parse int ", zap.Error(err))
			errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
			return
		}

		t := Txs{}
		// get transactions
		_, txs, err := a.Client.RPC.GetBlockAndTxsFromNode(custom.AppCodec, height)
		if err != nil {
			zap.S().Debug("failed to parse HTTP args ", zap.Error(err))
			errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
			return
		}

		for i := range txs {
			raw, err := custom.AppCodec.MarshalJSON(txs[i])
			if err != nil {
				zap.S().Debugf("failed to marshal tx hash = %s, err = %s\n", txs[i].TxHash, zap.Error(err))
				errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
				return
			}
			t.Txs = append(t.Txs, raw)
		}

		respond(rw, t)
		return
	}
}
