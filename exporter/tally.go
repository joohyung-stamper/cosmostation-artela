package exporter

import (
	"context"

	mdschema "github.com/cosmostation/mintscan-database/schema"
	"go.uber.org/zap"
)

// tally update

// updateProposal update proposal which is passed voting end time
func (ex *Exporter) UpdateTally(id uint64) error {
	tally, err := ex.Client.GRPC.GetProposalTallyResult_v1(context.Background(), id)
	if err != nil {
		zap.S().Errorf("failed to get tally on prop[%d]: %s", id, err)
		return err
	}

	p := &mdschema.Proposal{
		ID:         id,
		Yes:        tally.GetYesCount(),
		Abstain:    tally.GetAbstainCount(),
		No:         tally.GetNoCount(),
		NoWithVeto: tally.GetNoWithVetoCount(),
	}

	return ex.DB.InsertOrUpdateTally(p)
}
