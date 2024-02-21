package model

import (
	"time"
)

const (
	// Cosmos is the coin id of Cosmos Network for CoinGecko API.
	Cosmos = "cosmos"
)

// NetworkInfo defines the structure for chain's network information.
type NetworkInfo struct {
	BondendTokensPercentChange24H float64              `json:"bonded_tokens_percent_change_24h"`
	BondedTokensStats             []*BondedTokensStats `json:"bonded_tokens_stats"`
}

// BondedTokensStats defines the structure for bonded tokens statistics.
type BondedTokensStats struct {
	BondedTokens float64   `json:"bonded_tokens"`
	BondedRatio  float64   `json:"bonded_ratio"`
	LastUpdated  time.Time `json:"last_updated"`
}
