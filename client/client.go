package client

import (

	// cosmos-sdk
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	//internal
	"github.com/cosmostation/cosmostation-coreum/custom"

	//mbl
	mblclient "github.com/cosmostation/mintscan-backend-library/client"
	mblconfig "github.com/cosmostation/mintscan-backend-library/config"
)

// Client implements a wrapper around both Tendermint RPC HTTP client and
// Cosmos SDK REST client that allow for essential data queries.
type Client struct {
	*mblclient.Client
}

// NewClient creates a new client with the given configuration and
// return Client struct. An error is returned if it fails.
func NewClient(cfg *mblconfig.ClientConfig) *Client {
	client := mblclient.NewClient(cfg)

	client.CliCtx.Context = client.CliCtx.Context.
		// WithCodec(custom.EncodingConfig.Marshaler).
		WithCodec(custom.EncodingConfig.Codec).
		WithLegacyAmino(custom.EncodingConfig.Amino).
		WithTxConfig(custom.EncodingConfig.TxConfig).
		WithInterfaceRegistry(custom.EncodingConfig.InterfaceRegistry).
		WithAccountRetriever(authtypes.AccountRetriever{})

	return &Client{client}
}
