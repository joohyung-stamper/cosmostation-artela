package model

import "time"

const (
	// YES is one of the voting options that agrees to the proposal.
	YES = "Yes"

	// NO is one of the voting options that disagree with the proposal.
	NO = "No"

	// NOWITHVETO is one of the voting options that strongly disagree with the proposal.
	NOWITHVETO = "NoWithVeto"

	// ABSTAIN is the one of the voting options that gives up his/her voting right.
	ABSTAIN = "Abstain"
)

// Votes defines the structure for proposal votes.
type Votes struct {
	Voter   string    `json:"voter"`
	Moniker string    `json:"moniker" sql:"default:null"`
	Option  string    `json:"option"`
	TxHash  string    `json:"tx_hash"`
	Time    time.Time `json:"time"`
}
