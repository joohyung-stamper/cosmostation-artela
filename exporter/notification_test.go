package exporter

import (
	"testing"
	"time"

	"github.com/cosmostation/cosmostation-coreum/app"
)

var (
	// Bank module staking hashes
	SampleMsgSendTxHash      = "A80ADDA7929801AF3B1E6957BE9C63C30B5A0B9F903E760C555CAC19D2FC0DFC"
	SampleMsgMultiSendTxHash = ""
)

// TODO: no available tx hash in mainnet

func TestProposalAlarm(t *testing.T) {
	chainEx := app.NewApp("chain-exporter")
	ex = NewExporter(chainEx)
	go ex.ProposalNotificationToSlack(6)
	time.Sleep(5 * time.Second)
}
