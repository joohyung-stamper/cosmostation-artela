package model

import (
	"encoding/json"
	"net/http"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	mdschema "github.com/cosmostation/mintscan-database/schema"
)

type ResultAllBalances struct {
	Balance []Balance `json:"balances"`
}

type Balance struct {
	Denom       string `json:"denom"`
	Total       string `json:"total"`
	Available   string `json:"available"`
	Delegated   string `json:"delegated"`
	Undelegated string `json:"undelegated"`
	Rewards     string `json:"rewards"`
	Commission  string `json:"commission"`
	Vesting     string `json:"vesting"`
	Vested      string `json:"vested"`
}

// ResultTotalBalance defines the structure for total kava balance of a delegator.
type ResultTotalBalance struct {
	Total       sdktypes.Coin `json:"total"`
	Available   sdktypes.Coin `json:"available"`
	Delegated   sdktypes.Coin `json:"delegated"`
	Undelegated sdktypes.Coin `json:"undelegated"`
	Rewards     sdktypes.Coin `json:"rewards"`
	Commission  sdktypes.Coin `json:"commission"`
	Vesting     sdktypes.Coin `json:"vesting"`
	Vested      sdktypes.Coin `json:"vested"`
	// FailedVested sdk.Coin `json:"failed_vested"`
	// Incentive    sdk.Coin `json:"incentive"`
	// Deposited    sdk.Coin `json:"deposited"`
}

// ResultBlock defines the structure for block result response.
type ResultBlock struct {
	ID                     int64       `json:"id,omitempty"`
	ChainID                string      `json:"chainid"`
	Height                 int64       `json:"height"`
	Proposer               string      `json:"proposer"`
	OperatorAddress        string      `json:"operator_address,omitempty"`
	Moniker                string      `json:"moniker"`
	BlockHash              string      `json:"block_hash"`
	Identity               string      `json:"identity,omitempty"`
	NumSignatures          int64       `json:"num_signatures,omitempty" sql:",notnull"`
	NumTxs                 int64       `json:"num_txs"`
	TotalNumProposerBlocks int         `json:"total_num_proposer_blocks,omitempty"`
	Txs                    []*ResultTx `json:"txs"`
	Timestamp              time.Time   `json:"timestamp"`
}

// ResultDelegations defines the structure for delegations result response.
// account 상세 Delegations 카드에 나오는 데이터
type ResultDelegations struct {
	DelegatorAddress string `json:"delegator_address"`
	ValidatorAddress string `json:"validator_address"`
	Moniker          string `json:"moniker"`
	Shares           string `json:"shares"`
	// Balance          string `json:"balance"`
	// Balance          Coin   `json:"balance"`
	Amount string `json:"amount"`
	// Rewards []Coin `json:"delegator_rewards"`
	Rewards sdktypes.DecCoins `json:"delegator_rewards"`
}

// ResultProposal defines the structure for proposal result response.
type ResultProposal struct {
	ProposalID           uint64    `json:"proposal_id"`
	TxHash               string    `json:"tx_hash"`
	Proposer             string    `json:"proposer" sql:"default:null"`
	Moniker              string    `json:"moniker" sql:"default:null"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	ProposalType         string    `json:"proposal_type"`
	ProposalStatus       string    `json:"proposal_status"`
	Yes                  string    `json:"yes"`
	Abstain              string    `json:"abstain"`
	No                   string    `json:"no"`
	NoWithVeto           string    `json:"no_with_veto"`
	InitialDepositAmount string    `json:"initial_deposit_amount" sql:"default:null"`
	InitialDepositDenom  string    `json:"initial_deposit_denom" sql:"default:null"`
	TotalDepositAmount   []string  `json:"total_deposit_amount"`
	TotalDepositDenom    []string  `json:"total_deposit_denom"`
	SubmitTime           time.Time `json:"submit_time"`
	DepositEndtime       time.Time `json:"deposit_end_time" sql:"deposit_end_time"`
	VotingStartTime      time.Time `json:"voting_start_time"`
	VotingEndTime        time.Time `json:"voting_end_time"`
}

// ResultInflation defines the structure for inflation result response
type ResultInflation struct {
	Inflation float64 `json:"inflation"`
}

// ResultVote defines the structure for vote information result response.
type ResultVote struct {
	Tally *ResultTally `json:"tally"`
	Votes []*Votes     `json:"votes"`
}

// ResultProposalDetail defines the structure for deposit detail information result response.
type ResultProposalDetail struct {
	ProposalID         int64            `json:"proposal_id"`
	TotalVotesNum      int              `json:"total_votes_num"`
	TotalDepositAmount float64          `json:"total_deposit_amount"`
	ResultVoteInfo     ResultVote       `json:"vote_info"`
	DepositInfo        mdschema.Deposit `json:"deposit_info"`
}

// ResultStatus defines the structure for status result response.
type ResultStatus struct {
	ChainID                string                             `json:"chain_id"`
	BlockHeight            int64                              `json:"block_height"`
	BlockTime              float64                            `json:"block_time"`
	TotalTxsNum            int                                `json:"total_txs_num"`
	TotalValidatorNum      int                                `json:"total_validator_num"`
	UnjailedValidatorNum   int                                `json:"unjailed_validator_num"`
	JailedValidatorNum     int                                `json:"jailed_validator_num"`
	TotalSupplyTokens      banktypes.QueryTotalSupplyResponse `json:"total_supply_tokens"`
	TotalCirculatingTokens banktypes.QueryTotalSupplyResponse `json:"total_circulating_tokens"`
	BondedTokens           float64                            `json:"bonded_tokens"`
	NotBondedTokens        float64                            `json:"not_bonded_tokens"`
	Inflation              sdktypes.Dec                       `json:"inflation"`
	// CommunityPool          *distrtypes.QueryCommunityPoolResponse `json:"community_pool"`
	CommunityPool sdktypes.DecCoins `json:"community_pool"`
	Timestamp     time.Time         `json:"timestamp"`
}

// ResultValidator defines the structure for validator result response.
type ResultValidator struct {
	Rank                 int       `json:"rank"`
	AccountAddress       string    `json:"account_address"`
	OperatorAddress      string    `json:"operator_address"`
	ConsensusPubkey      string    `json:"consensus_pubkey"`
	Jailed               bool      `json:"jailed"`
	Status               int       `json:"status"`
	Tokens               string    `json:"tokens"`
	DelegatorShares      string    `json:"delegator_shares"`
	Moniker              string    `json:"moniker"`
	Identity             string    `json:"identity"`
	Website              string    `json:"website"`
	Details              string    `json:"details"`
	UnbondingHeight      string    `json:"unbonding_height"`
	UnbondingTime        time.Time `json:"unbonding_time"`
	CommissionRate       string    `json:"rate"`
	CommissionMaxRate    string    `json:"max_rate"`
	CommissionChangeRate string    `json:"max_change_rate"`
	UpdateTime           time.Time `json:"update_time"`
	Uptime               *Uptime   `json:"uptime"`
	MinSelfDelegation    string    `json:"min_self_delegation"`
	KeybaseURL           string    `json:"keybase_url"`
}

// ResultValidatorDetail defines the structure for validator detail result response.
type ResultValidatorDetail struct {
	Rank                 int       `json:"rank"`
	AccountAddress       string    `json:"account_address"`
	OperatorAddress      string    `json:"operator_address"`
	ConsensusPubkey      string    `json:"consensus_pubkey"`
	BondedHeight         int64     `json:"bonded_height"`
	BondedTime           time.Time `json:"bonded_time"`
	Jailed               bool      `json:"jailed"`
	Status               int       `json:"status"`
	Tokens               string    `json:"tokens"`
	DelegatorShares      string    `json:"delegator_shares"`
	Moniker              string    `json:"moniker"`
	Identity             string    `json:"identity"`
	Website              string    `json:"website"`
	Details              string    `json:"details"`
	UnbondingHeight      string    `json:"unbonding_height"`
	UnbondingTime        time.Time `json:"unbonding_time"`
	CommissionRate       string    `json:"rate"`
	CommissionMaxRate    string    `json:"max_rate"`
	CommissionChangeRate string    `json:"max_change_rate"`
	UpdateTime           time.Time `json:"update_time"`
	Uptime               *Uptime   `json:"uptime"`
	MinSelfDelegation    string    `json:"min_self_delegation"`
	KeybaseURL           string    `json:"keybase_url"`
}

// ResultMisses defines the structure for validator miss blocks result response.
type ResultMisses struct {
	StartHeight  int64     `json:"start_height"`
	EndHeight    int64     `json:"end_height"`
	MissingCount int64     `json:"missing_count"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

type (
	// ResultMissesDetail defines the structure for validator miss block detail result response.
	ResultMissesDetail struct {
		ID           int64          `json:"id"`
		LatestHeight int64          `json:"latest_height"`
		ResultUptime []ResultUptime `json:"uptime"`
	}

	// ResultUptime defines the structure for validator's uptime (last 100 blocks from the latest block).
	ResultUptime struct {
		ID        int64     `json:"id"`
		Height    int64     `json:"height"`
		Timestamp time.Time `json:"timestamp"`
	}
)

// ResultPowerEventHistory defines the structure for validator voting power history result response.
type ResultPowerEventHistory struct {
	ID             int64     `json:"id"`
	Height         int64     `json:"height"`
	MsgType        string    `json:"msg_type"`
	VotingPower    float64   `json:"voting_power"`
	NewVotingPower float64   `json:"new_voting_power"`
	TxHash         string    `json:"tx_hash"`
	Timestamp      time.Time `json:"timestamp"`
}

// ResultVotingPowerHistoryCount wraps count for validator's power event history.
type ResultVotingPowerHistoryCount struct {
	Moniker         string `json:"moniker"`
	OperatorAddress string `json:"operator_address"`
	Count           int    `json:"count"`
}

// ResultValidatorDelegations defines the structure for validator delegations result response.
type ResultValidatorDelegations struct {
	TotalDelegatorNum     int                     `json:"total_delegator_num"`
	DelegatorNumChange24H int                     `json:"delegator_num_change_24h"`
	ValidatorDelegations  []*ValidatorDelegations `json:"delegations"`
}

// ResultRewards defines the structure for rewards result response.
// type ResultRewards struct {
// 	Rewards []Rewards `json:"rewards"`
// 	Total   []Coin    `json:"total"`
// }

// ResultTally defines the structure for tally result response.
type ResultTally struct {
	YesAmount        string `json:"yes_amount"`
	AbstainAmount    string `json:"abstain_amount"`
	NoAmount         string `json:"no_amount"`
	NoWithVetoAmount string `json:"no_with_veto_amount"`
	YesNum           int    `json:"yes_num"`
	AbstainNum       int    `json:"abstain_num"`
	NoNum            int    `json:"no_num"`
	NoWithVetoNum    int    `json:"no_with_veto_num"`
}

// ResultDeposit defines the structure for deposit result response.
type ResultDeposit struct {
	Depositor     string    `json:"depositor"`
	Moniker       string    `json:"moniker" sql:"default:null"`
	DepositAmount string    `json:"deposit_amount"`
	DepositDenom  string    `json:"deposit_denom"`
	Height        int64     `json:"height"`
	TxHash        string    `json:"tx_hash"`
	Timestamp     time.Time `json:"timestamp"`
}

// ResultMarket defines the structure for market result response.
type ResultMarket struct {
	Price             float64   `json:"price"`
	Currency          string    `json:"currency"`
	MarketCapRank     uint      `json:"market_cap_rank"`
	PercentChange1H   float64   `json:"percent_change_1h"`
	PercentChange24H  float64   `json:"percent_change_24h"`
	PercentChange7D   float64   `json:"percent_change_7d"`
	PercentChange30D  float64   `json:"percent_change_30d"`
	TotalVolume       float64   `json:"total_volume"`
	CirculatingSupply float64   `json:"circulating_supply"`
	LastUpdated       time.Time `json:"last_updated"`
}

type ResultBlockHeader struct {
	ID        int64  `json:"id"`
	ChainID   string `json:"chain_id"`
	Timestamp string `json:"timestamp"`
}
type ResultTxHeader struct {
	ID        int64  `json:"id"`
	ChainID   string `json:"chain_id"`
	BlockID   int64  `json:"block_id"`
	Timestamp string `json:"timestamp"`
}

// ResultTx defines the structure for txs result response.
type ResultTx struct {
	ResultTxHeader `json:"header"`
	Data           json.RawMessage `json:"data"`
}

// Respond responds result of any data type.
func Respond(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	switch data := data.(type) {
	case []byte:
		// fmt.Printf("type : %T\n", data)
		w.Write(data)
	default:
		// fmt.Printf("type : %T\n", data)
		json.NewEncoder(w).Encode(data)
	}
}
