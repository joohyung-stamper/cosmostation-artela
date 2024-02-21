package common

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cosmostation/cosmostation-coreum/app"
	"github.com/cosmostation/cosmostation-coreum/errors"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	tmjson "github.com/cometbft/cometbft/libs/json"
	// tmctype "github.com/tendermint/tendermint/rpc/core/types"
)

func GetBlock(a *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		heightStr := vars["height"]

		height, err := strconv.ParseInt(heightStr, 10, 64)
		if err != nil {
			zap.S().Debug("failed to parse int ", zap.Error(err))
			errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
			return
		}

		// get block
		block, err := a.Client.RPC.Block(context.Background(), &height)
		if err != nil {
			zap.S().Debug("failed to parse HTTP args ", zap.Error(err))
			errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
		}

		// marshal
		blockChunk, err := tmjson.Marshal(block)
		if err != nil {
			zap.S().Debugf("failed to marshal block, height = %s, err = %s\n", heightStr, zap.Error(err))
			errors.ErrServerUnavailable(rw, http.StatusInternalServerError)
			return
		}

		respond(rw, blockChunk)
		return
	}
}
