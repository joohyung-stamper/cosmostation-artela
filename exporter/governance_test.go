package exporter

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmostation/cosmostation-coreum/custom"
	govutil "github.com/cosmostation/mintscan-backend-library/types"

	mdschema "github.com/cosmostation/mintscan-database/schema"

	"github.com/stretchr/testify/require"
)

func TestGetProposals(t *testing.T) {
	resp, err := ex.GetAllProposals_v1()
	require.NoError(t, err)

	for i := range resp {
		t.Log("id :", resp[i].ID)
		t.Log("metadata :", resp[i].Metadata)
		t.Log("title :", resp[i].Title)
		t.Log("desc :", resp[i].Description)
	}
}
func TestSaveAllProposals(t *testing.T) {
	ex.saveAllProposals()
}

func TestGetRecoveryVote(t *testing.T) {
	// beginTxID := int64(35778600)
	beginTxID := int64(62636577)
	msgType := "gov/vote"
	limit := 50
	for {
		txs, err := ex.DB.GetTransactionsByMsgType(beginTxID, msgType, limit)
		require.NoError(t, err)

		t.Log(len(txs))
		if len(txs) == 0 {
			break
		}

		txResps := make([]*sdktypes.TxResponse, 0)
		for i := range txs {
			txResp := &sdktypes.TxResponse{}
			custom.AppCodec.UnmarshalJSON(txs[i].Chunk, txResp)
			txResps = append(txResps, txResp)
			beginTxID = txs[i].ID
		}
		// p, d, v, err := ex.getGovernance(nil, txResps)
		// require.NoError(t, err)
		// _, _ = p, d
		// t.Log("votes := ", len(v))
		// for i := range v {
		// 	t.Log(v[i].ProposalID, v[i].Voter, v[i].Option, v[i].Timestamp)
		// }
		beginTxID++
		t.Log(beginTxID)
		basic := new(mdschema.BasicData)
		basic.Proposals, basic.Deposits, basic.Votes, err = ex.getGovernance(nil, txResps)
		require.NoError(t, err)
		ex.DB.InsertExportedData(basic)

	}
}

func TestUpdateProposal(t *testing.T) {
	// 기존 프로포절에 IPFS를 강제로 삽입하여, IPFS 로직이 정상동작하는지 테스트
	p, err := ex.GetProposal_v1(15)
	require.NoError(t, err)
	meta := "ipfs://bafkreie7tvsxl7fxsubn5joenzqdcvelry6orszk3wp3iksaspzxwcko34"

	p.ID = 999
	p.Metadata = meta

	var metadataChunk []byte
	title := p.Title
	desc := p.Description
	var tempTitle string
	if p.Metadata != "" {
		tempTitle, desc, err = govutil.ExtractMetadata_v1(p.Metadata)
		if err != nil {
			hash := govutil.ExtractURLFromMetadata(p.Metadata)
			metadataChunk, tempTitle, desc, _ = ex.Client.GetMetadataFromIPFS(hash)
			// if err != nil {
			// ipfs 관련 오류 무시
			// return result, fmt.Errorf("failed to get proposal metadata : %s", err)
			// }
		}
	}
	if tempTitle != "" {
		title = tempTitle
	}
	p.Title = title
	p.MetadataChunk = metadataChunk
	p.MetadataChunk = []byte{}
	p.Description = desc

	t.Log(metadataChunk)
	ex.DB.InsertOrUpdateProposal(p)
}

func TestGetProposalUpdateNeeded(t *testing.T) {
	ps, err := ex.DB.GetProposalUpdateNeeded()
	require.NoError(t, err)

	t.Log(ps)

}

func TestGetProposals_v1(t *testing.T) {
	keyExists := true
	nextKey := make([]byte, 8)
	binary.BigEndian.PutUint64(nextKey, 0)
	for keyExists {

		res, err := ex.Client.GRPC.GetProposals_v1(context.Background(), nextKey)
		require.NoError(t, err)

		nextKey = res.Pagination.GetNextKey()
		keyExists = len(nextKey) > 0

		for _, p := range res.Proposals {

			chunk, err := custom.AppCodec.MarshalJSON(p)
			require.NoError(t, err)
			tda, tdd := govutil.CoinsToString_v1(p.TotalDeposit)

			title, desc, pType, err := govutil.ExportProposalAttribute_v1(chunk)
			require.NoError(t, err)

			tempTitle, desc, err := govutil.ExtractMetadata_v1(p.Metadata)
			if err == nil || tempTitle != "" {
				title = tempTitle
			}

			tally, err := ex.Client.GRPC.GetProposalTallyResult_v1(context.Background(), p.Id)
			if err != nil {
				return
			}

			if p.VotingStartTime == nil {
				t.Log("voting time is nil")
				t.Log(p.Status.String())
				var ut time.Time
				p.VotingStartTime = &ut
				p.VotingEndTime = &ut
				t.Log(ut)
			}

			mdProp := &mdschema.Proposal{
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
				TotalDepositAmount: tda, // totalDepositAmount,
				TotalDepositDenom:  tdd,
				VotingStartTime:    *p.VotingStartTime,
				VotingEndTime:      *p.VotingEndTime,
				Metadata:           p.Metadata,
				GovRestPath:        "v1",
				Chunk:              chunk,
			}

			if p.Status < v1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
				fmt.Println(mdProp)
			}
		}
	}
}

func TestSaveAllProposalsWithoutCondition(t *testing.T) {
	proposals, err := ex.GetAllProposals_v1()
	require.NoError(t, err)

	if len(proposals) <= 0 {
		t.Log("found empty proposals")
		return
	}

	err = ex.DB.InsertOrUpdateProposals(proposals)
	require.NoError(t, err)
}
