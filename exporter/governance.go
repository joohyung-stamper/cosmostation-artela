package exporter

import (
	"strconv"
	"sync"
	"time"

	"github.com/cosmostation/cosmostation-coreum/custom"
	govutil "github.com/cosmostation/mintscan-backend-library/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"

	//cosmos-sdk
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"go.uber.org/zap"
)

var muProp sync.RWMutex

const (
	SUBMIT_PROPOSAL = iota
	DEPOSIT         // update total deposit & proposal status
	VOTED           // update tally & proposal status
)

type propFlag struct {
	TxHash string
	Flag   int // deposit
}

var propList = make(map[uint64]struct{})

func (ex *Exporter) watchLiveProposals() {
	for {
		p, err := ex.DB.GetLiveProposalIDs()
		if err != nil {
			zap.S().Info("failed to get live proposals")
			time.Sleep(2 * time.Second)
			continue
		}
		muProp.Lock()
		for i := range p {
			// _ := govtypesv1.ProposalStatus_value[p[i].ProposalStatus]
			_, ok := propList[p[i].ID]
			if !ok {
				propList[p[i].ID] = struct{}{}
			}
		}
		muProp.Unlock()
		zap.S().Info("proposal list updated")
		time.Sleep(6 * time.Second)
	}
}

func (ex *Exporter) updateProposals() {
	for {
		if !ex.App.CatchingUp {
			zap.S().Info("start updating proposals : ", propList)
			muProp.RLock()
			for id := range propList {
				if err := ex.updateProposal(id); err != nil {
					continue
				}
				delete(propList, id)
			}
			muProp.RUnlock()
			zap.S().Info("finish update proposals : ", propList)
		} else {
			zap.S().Info("pending update proposals, app is catching up")
		}
		time.Sleep(10 * time.Second)
	}
}

// getGovernance returns governance by decoding governance related transactions in a block.
func (ex *Exporter) getGovernance(blockTimeStamp *time.Time, txResp []*sdktypes.TxResponse) ([]mdschema.Proposal, []mdschema.Deposit, []mdschema.Vote, error) {
	proposals := make([]mdschema.Proposal, 0)
	deposits := make([]mdschema.Deposit, 0)
	votes := make([]mdschema.Vote, 0)
	distinctVotes := make(map[string]map[uint64][]mdschema.Vote, 0)

	if len(txResp) <= 0 {
		return proposals, deposits, votes, nil
	}

	for _, tx := range txResp {
		// code == 0 이면, 오류 트랜잭션이다.
		if tx.Code != 0 {
			continue
		}

		ts := blockTimeStamp
		//blockTimeStamp가 nil 인 경우 각 tx의 timestamp로 처리한다.
		if ts == nil {
			t, err := time.Parse(time.RFC3339, tx.Timestamp)
			if err != nil {
				return proposals, deposits, votes, err
			}
			ts = &t
		}

		msgs := tx.GetTx().GetMsgs()

		for i, msg := range msgs {
			votesInMsg := make([]mdschema.Vote, 0)
			switch m := msg.(type) {
			case *authztypes.MsgExec:
				for j := range m.Msgs {
					var msgExecAuthorized sdktypes.Msg
					custom.AppCodec.UnpackAny(m.Msgs[j], &msgExecAuthorized)
					ex.govRoute(&proposals, &deposits, &votesInMsg, ts, msgExecAuthorized, i, tx)
				}

			default:
				ex.govRoute(&proposals, &deposits, &votesInMsg, ts, msg, i, tx)
			}

			// votes 결과를 덮어쓴다.(최신 보트만 유지)
			for i := range votesInMsg {
				voterMap, ok := distinctVotes[votesInMsg[i].Voter]
				if !ok {
					voterMap = make(map[uint64][]mdschema.Vote)
					distinctVotes[votesInMsg[i].Voter] = voterMap
				}
				voterMap[votesInMsg[i].ProposalID] = votesInMsg
			}
		}
		// get effective votes : map[address]map[id]vote
		// 같은 tx 내 메세지에 대해 index 순으로 덮어쓰고, 그 다음 tx에 대해 덮어 쓴다.
		// vote -> vote
		// vote -> weighted_vote -> vote
	}
	// 최종 보트
	for _, voterMap := range distinctVotes {
		for _, vote := range voterMap {
			votes = append(votes, vote...)
		}
	}

	return proposals, deposits, votes, nil
}

func (ex *Exporter) govRoute(ps *[]mdschema.Proposal, ds *[]mdschema.Deposit, vs *[]mdschema.Vote, ts *time.Time, msg sdktypes.Msg, logIndex int, tx *sdktypes.TxResponse) {

	switch m := msg.(type) {
	case *govtypesv1beta1.MsgSubmitProposal:
		zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)

		// Get proposal id for this proposal.
		// Handle case of multiple messages which has multiple events and attributes.
		var proposalID uint64
		for _, event := range tx.Logs[logIndex].Events {
			if event.Type == "submit_proposal" {
				for _, attribute := range event.Attributes {
					if attribute.Key == "proposal_id" {
						proposalID, _ = strconv.ParseUint(attribute.Value, 10, 64)
					}
				}
			}
		}

		var initialDepositAmount string
		var initialDepositDenom string

		if len(m.InitialDeposit) > 0 {
			initialDepositAmount = m.InitialDeposit[0].Amount.String()
			initialDepositDenom = m.InitialDeposit[0].Denom
		}

		p := mdschema.Proposal{
			ID:                   proposalID,
			TxHash:               tx.TxHash,
			Proposer:             m.Proposer,
			InitialDepositAmount: initialDepositAmount,
			InitialDepositDenom:  initialDepositDenom,
		}

		*ps = append(*ps, p)

		d := mdschema.Deposit{
			Height:     tx.Height,
			ProposalID: proposalID,
			Depositor:  m.Proposer,
			Amount:     initialDepositAmount,
			Denom:      initialDepositDenom,
			TxHash:     tx.TxHash,
			GasWanted:  tx.GasWanted,
			GasUsed:    tx.GasUsed,
			Timestamp:  *ts,
		}
		*ds = append(*ds, d)

		go ex.ProposalNotificationToSlack(p.ID)

	case *govtypesv1beta1.MsgDeposit:
		zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)

		var amount string
		var denom string

		if len(m.Amount) > 0 {
			amount = m.Amount[0].Amount.String()
			denom = m.Amount[0].Denom
		}

		d := mdschema.Deposit{
			Height:     tx.Height,
			ProposalID: m.ProposalId,
			Depositor:  m.Depositor,
			Amount:     amount,
			Denom:      denom,
			TxHash:     tx.TxHash,
			GasWanted:  tx.GasWanted,
			GasUsed:    tx.GasUsed,
			Timestamp:  *ts,
		}

		*ds = append(*ds, d)
		go ex.ProposalNotificationToSlack(d.ProposalID)

	case *govtypesv1beta1.MsgVote:
		zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)
		v := mdschema.Vote{
			Height:     tx.Height,
			ProposalID: m.ProposalId,
			Voter:      m.Voter,
			Option:     m.Option.String(),
			Weight:     sdktypes.OneDec().String(),
			TxHash:     tx.TxHash,
			GasWanted:  tx.GasWanted,
			GasUsed:    tx.GasUsed,
			Timestamp:  *ts,
		}

		*vs = append(*vs, v)

	case *govtypesv1beta1.MsgVoteWeighted:
		zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)
		for i := range m.Options {
			v := mdschema.Vote{
				Height:     tx.Height,
				ProposalID: m.ProposalId,
				Voter:      m.Voter,
				Option:     m.Options[i].Option.String(),
				Weight:     m.Options[i].Weight.String(),
				TxHash:     tx.TxHash,
				GasWanted:  tx.GasWanted,
				GasUsed:    tx.GasUsed,
				Timestamp:  *ts,
			}
			*vs = append(*vs, v)
		}

	case *govtypesv1.MsgSubmitProposal:
		zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)

		// Get proposal id for this proposal.
		// Handle case of multiple messages which has multiple events and attributes.
		var proposalID uint64
		for _, event := range tx.Logs[logIndex].Events {
			if event.Type == "submit_proposal" {
				for _, attribute := range event.Attributes {
					if attribute.Key == "proposal_id" {
						proposalID, _ = strconv.ParseUint(attribute.Value, 10, 64)
					}
				}
			}
		}

		var initialDepositAmount string
		var initialDepositDenom string

		if len(m.InitialDeposit) > 0 {
			initialDepositAmount = m.InitialDeposit[0].Amount.String()
			initialDepositDenom = m.InitialDeposit[0].Denom
		}
	RETRY:
		chunk, err := custom.AppCodec.MarshalJSON(m)
		if err != nil {
			time.Sleep(2 * time.Second)
			zap.S().Error("failed to marshal submit proposal... retry")
			goto RETRY
		}

		title, desc, pType, err := govutil.ExportProposalAttribute_v1(chunk)
		if err != nil {
			zap.S().Error("failed to get proposal details : %s", err)
			goto RETRY
		}

		var metadataChunk []byte
		var tempTitle, tempDesc string
		if m.Metadata != "" {
			tempTitle, tempDesc, err = govutil.ExtractMetadata_v1(m.Metadata)
			if err != nil {
				hash := govutil.ExtractURLFromMetadata(m.Metadata)
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

		p := mdschema.Proposal{
			ID:                   proposalID,
			TxHash:               tx.TxHash,
			ProposalType:         pType,
			Title:                title,
			Description:          desc,
			Proposer:             m.Proposer,
			Metadata:             m.Metadata,
			MetadataChunk:        metadataChunk,
			InitialDepositAmount: initialDepositAmount,
			InitialDepositDenom:  initialDepositDenom,
		}

		*ps = append(*ps, p)

		d := mdschema.Deposit{
			Height:     tx.Height,
			ProposalID: proposalID,
			Depositor:  m.Proposer,
			Amount:     initialDepositAmount,
			Denom:      initialDepositDenom,
			TxHash:     tx.TxHash,
			GasWanted:  tx.GasWanted,
			GasUsed:    tx.GasUsed,
			Timestamp:  *ts,
		}
		*ds = append(*ds, d)

		go ex.ProposalNotificationToSlack(p.ID)

	case *govtypesv1.MsgDeposit:
		zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)

		var amount string
		var denom string

		if len(m.Amount) > 0 {
			amount = m.Amount[0].Amount.String()
			denom = m.Amount[0].Denom
		}

		d := mdschema.Deposit{
			Height:     tx.Height,
			ProposalID: m.ProposalId,
			Depositor:  m.Depositor,
			Amount:     amount,
			Denom:      denom,
			TxHash:     tx.TxHash,
			GasWanted:  tx.GasWanted,
			GasUsed:    tx.GasUsed,
			Timestamp:  *ts,
		}

		*ds = append(*ds, d)
		go ex.ProposalNotificationToSlack(d.ProposalID)

	case *govtypesv1.MsgVote:
		zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)

		v := mdschema.Vote{
			Height:     tx.Height,
			ProposalID: m.ProposalId,
			Voter:      m.Voter,
			Option:     m.Option.String(),
			Weight:     sdktypes.OneDec().String(),
			Metadata:   m.Metadata,
			TxHash:     tx.TxHash,
			GasWanted:  tx.GasWanted,
			GasUsed:    tx.GasUsed,
			Timestamp:  *ts,
		}

		*vs = append(*vs, v)

	case *govtypesv1.MsgVoteWeighted:
		zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)
		for i := range m.Options {
			v := mdschema.Vote{
				Height:     tx.Height,
				ProposalID: m.ProposalId,
				Voter:      m.Voter,
				Option:     m.Options[i].Option.String(),
				Weight:     m.Options[i].Weight,
				Metadata:   m.Metadata,
				TxHash:     tx.TxHash,
				GasWanted:  tx.GasWanted,
				GasUsed:    tx.GasUsed,
				Timestamp:  *ts,
			}
			*vs = append(*vs, v)
		}

	}
}

// saveProposals saves all governance proposals
func (ex *Exporter) saveAllProposals() {
	NodePropCount, err := ex.Client.GRPC.GetNumberofProposals_v1()
	if err != nil {
		zap.S().Errorf("failed to get number of proposal from DB: %s", err)
		return
	}
	DBPropCount, err := ex.DB.GetNumberofValidProposal()
	if err != nil {
		zap.S().Errorf("failed to get number of proposal from Node: %s", err)
		return
	}
	// database에 저장된 프로포절의 수와 노드의 수가 같으면 업데이트 하지 않는다.
	if NodePropCount == uint64(DBPropCount) {
		zap.S().Info("skip saveAllProposals, all proposals have already been stored in database, count : ", NodePropCount)
		return
	}

	proposals, err := ex.GetAllProposals_v1()
	if err != nil {
		zap.S().Errorf("failed to get proposals: %s", err)
		return
	}

	if len(proposals) <= 0 {
		zap.S().Info("found empty proposals")
		return
	}

	err = ex.DB.InsertOrUpdateProposals(proposals)
	if err != nil {
		zap.S().Errorf("failed to insert or update proposal: %s", err)
		return
	}
}

// updateProposal update proposal which is passed voting end time
func (ex *Exporter) updateProposal(id uint64) error {
	p, err := ex.GetProposal_v1(id)
	if err != nil {
		zap.S().Errorf("failed to get proposal: %s", err)
		return err
	}

	return ex.DB.InsertOrUpdateProposal(p)
}
