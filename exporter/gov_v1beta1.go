package exporter

import (
	"context"
	"fmt"

	"github.com/cosmostation/cosmostation-coreum/custom"
	govutil "github.com/cosmostation/mintscan-backend-library/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"

	// cosmos-sdk
	"github.com/cosmos/cosmos-sdk/codec"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func (ex *Exporter) GetAllProposals_v1beta1() (result []mdschema.Proposal, err error) {
	keyExists := true
	var nextKey []byte
	for keyExists {
		res, err := ex.Client.GRPC.GetProposals_v1beta1(context.Background(), nextKey)
		if err != nil {
			return []mdschema.Proposal{}, fmt.Errorf("failed to request gov proposals: %s", err)
		}

		nextKey = res.Pagination.GetNextKey()
		keyExists = len(nextKey) > 0

		if len(res.Proposals) <= 0 {
			return []mdschema.Proposal{}, nil
		}

		for _, p := range res.Proposals {

			prop, err := ex.transformProposal_v1beta1(custom.AppCodec, &p)
			if err != nil {
				return []mdschema.Proposal{}, fmt.Errorf("failed to make proposal : %s", err)
			}

			result = append(result, *prop)
		}
	}

	return result, nil
}

// GetProposal_v1beta1은 특정 프로포절 정보를 GRPC 얻어온다.
func (ex *Exporter) GetProposal_v1beta1(id uint64) (result *mdschema.Proposal, err error) {
	p, err := ex.Client.GRPC.GetProposal_v1beta1(context.Background(), id)
	if err != nil {
		return result, fmt.Errorf("failed to request gov proposals: %s", err)
	}

	return ex.transformProposal_v1beta1(custom.AppCodec, p)
}

func (ex *Exporter) transformProposal_v1beta1(cdc codec.Codec, p *v1beta1.Proposal) (result *mdschema.Proposal, err error) {
	chunk, err := cdc.MarshalJSON(p)
	if err != nil {
		return result, fmt.Errorf("failed to marshal proposal: %s", err)
	}

	// total depoist : amount / denom
	tda, tdd := govutil.CoinsToString_v1beta1(p.TotalDeposit)

	tally, err := ex.Client.GRPC.GetProposalTallyResult_v1beta1(context.Background(), p.ProposalId)
	if err != nil {
		return result, fmt.Errorf("failed to request gov proposals: %s", err)
	}

	title, pType, err := govutil.ExportProposalAttribute_v1beta1(chunk)
	if err != nil {
		return result, fmt.Errorf("failed to get proposal details : %s", err)
	}

	tempTitle, desc, err := govutil.ExtractMetadata_v1beta1("")
	if err == nil || tempTitle != "" {
		title = tempTitle
	}
	result = &mdschema.Proposal{
		ID:                 p.ProposalId,
		Title:              title,
		Description:        desc,
		ProposalType:       pType,
		ProposalStatus:     p.Status.String(),
		Yes:                tally.Yes.String(),
		Abstain:            tally.Abstain.String(),
		No:                 tally.No.String(),
		NoWithVeto:         tally.NoWithVeto.String(),
		SubmitTime:         p.SubmitTime,
		DepositEndTime:     p.DepositEndTime,
		TotalDepositAmount: tda,
		TotalDepositDenom:  tdd,
		VotingStartTime:    p.VotingStartTime,
		VotingEndTime:      p.VotingEndTime,
		Metadata:           "",
		MetadataChunk:      []byte{},
		GovRestPath:        "v1beta1",
		Chunk:              chunk,
	}
	return result, nil
}
