package common

import (
	"github.com/cosmostation/cosmostation-coreum/app"
	"github.com/gorilla/mux"
)

// RegisterHandlers registers all common query HTTP REST handlers on the provided mux router
func RegisterHandlers(a *app.App, r *mux.Router) {
	r.HandleFunc("/blocks/latest", GetBlocksLatest(a)).Methods("GET")
	r.HandleFunc("/block_txs/{height}", GetBlockTxs(a)).Methods("GET")
	r.HandleFunc("/block/{height}", GetBlock(a)).Methods("GET")
	r.HandleFunc("/txs/{height}", GetTxs(a)).Methods("GET")
	r.HandleFunc("/block_results/{height}", GetBlockResults(a)).Methods("GET")
}
