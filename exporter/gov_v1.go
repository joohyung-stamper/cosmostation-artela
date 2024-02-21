package exporter

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmostation/cosmostation-coreum/custom"
	govutil "github.com/cosmostation/mintscan-backend-library/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"

	// cosmos-sdk
	"github.com/cosmos/cosmos-sdk/codec"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (ex *Exporter) GetAllProposals_v1() (result []mdschema.Proposal, err error) {
	keyExists := true
	var nextKey []byte
	for keyExists {
		res, err := ex.Client.GRPC.GetProposals_v1(context.Background(), nextKey)
		if err != nil {
			return []mdschema.Proposal{}, fmt.Errorf("failed to request gov proposals: %s", err)
		}

		nextKey = res.Pagination.GetNextKey()
		keyExists = len(nextKey) > 0

		if len(res.Proposals) <= 0 {
			return []mdschema.Proposal{}, nil
		}

		for _, p := range res.Proposals {

			prop, err := ex.transformProposal_v1(custom.AppCodec, p)
			if err != nil {
				return []mdschema.Proposal{}, fmt.Errorf("failed to make proposal : %s", err)
			}

			result = append(result, *prop)
		}
	}

	return result, nil
}

// GetProposal_v1은 특정 프로포절 정보를 GRPC 얻어온다.
func (ex *Exporter) GetProposal_v1(id uint64) (result *mdschema.Proposal, err error) {
	p, err := ex.Client.GRPC.GetProposal_v1(context.Background(), id)
	if err != nil {
		return result, fmt.Errorf("failed to request gov proposals: %s", err)
	}

	return ex.transformProposal_v1(custom.AppCodec, p)
}

// metadata 업데이트가 필요한 경우 :
// select * from proposal where metadata is not NULL and metadata_chunk is NULL -> ipfs:// 스킴인데, db에 metadata_chunk가 저장되지 않은 경우

// metadata_chunk 업데이트가 필요 없는 경우 :
// len(metadata) == 0
// ipfs:// 스킴이 아님
// db에 metadata_chunk가 있음
func (ex *Exporter) transformProposal_v1(cdc codec.Codec, p *v1.Proposal) (result *mdschema.Proposal, err error) {
	chunk, err := cdc.MarshalJSON(p)
	if err != nil {
		return result, fmt.Errorf("failed to marshal proposal: %s", err)
	}

	// total depoist : amount / denom
	tda, tdd := govutil.CoinsToString_v1(p.TotalDeposit)

	tally, err := ex.Client.GRPC.GetProposalTallyResult_v1(context.Background(), p.Id)
	if err != nil {
		return result, fmt.Errorf("failed to request gov proposals: %s", err)
	}

	title, desc, pType, err := govutil.ExportProposalAttribute_v1(chunk)
	if err != nil {
		return result, fmt.Errorf("failed to get proposal details : %s", err)
	}

	var metadataChunk []byte
	var tempTitle, tempDesc string
	if p.Metadata != "" {
		tempTitle, tempDesc, err = govutil.ExtractMetadata_v1(p.Metadata)
		if err != nil {
			hash := govutil.ExtractURLFromMetadata(p.Metadata)
			metadataChunk, tempTitle, tempDesc, _ = ex.Client.GetMetadataFromIPFS(hash)
			// if err != nil {
			// ipfs 관련 오류 무시
			// return result, fmt.Errorf("failed to get proposal metadata : %s", err)
			// }
		}
	}
	if tempTitle != "" {
		title = tempTitle
	}
	if tempDesc != "" {
		desc = tempDesc
	}

	if p.VotingStartTime == nil {
		var ut time.Time
		p.VotingStartTime = &ut
		p.VotingEndTime = &ut
	}

	result = &mdschema.Proposal{
		ID:                 p.Id,
		Title:              title,
		Description:        desc,
		ProposalType:       pType,
		ProposalStatus:     p.Status.String(),
		Yes:                tally.GetYesCount(),
		Abstain:            tally.GetAbstainCount(),
		No:                 tally.GetNoCount(),
		NoWithVeto:         tally.GetNoWithVetoCount(),
		SubmitTime:         *p.SubmitTime,
		DepositEndTime:     *p.DepositEndTime,
		TotalDepositAmount: tda,
		TotalDepositDenom:  tdd,
		VotingStartTime:    *p.VotingStartTime,
		VotingEndTime:      *p.VotingEndTime,
		Metadata:           p.Metadata,
		MetadataChunk:      metadataChunk,
		GovRestPath:        "v1",
		Chunk:              chunk,
	}
	return result, nil
}
