package db

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	mdschema "github.com/cosmostation/mintscan-database/schema"
	pg "github.com/go-pg/pg/v10"
)

// QueryVoteOptions returns all voting option counts for the proposal
// Voting options are Yes, No, NoWithVeto, and Abstain
func (db *Database) QueryVoteOptions(proposalID uint64) (yes, no, noWithVeto, abstain int, err error) {
	var optionCount []struct {
		Option string
		Count  int
	}

	err = db.Model(&mdschema.Vote{}).
		Column("option").
		ColumnExpr("count(option) AS count").
		Where("proposal_id = ?", proposalID).
		Group("option").
		Select(&optionCount)

	if err != nil {
		if err == pg.ErrNoRows {
			return yes, no, noWithVeto, abstain, nil
		}
		return yes, no, noWithVeto, abstain, err
	}

	for _, oc := range optionCount {
		switch oc.Option {
		case govtypes.VoteOption_name[int32(govtypes.OptionYes)]:
			yes = oc.Count
		case govtypes.VoteOption_name[int32(govtypes.OptionNo)]:
			no = oc.Count
		case govtypes.VoteOption_name[int32(govtypes.OptionNoWithVeto)]:
			noWithVeto = oc.Count
		case govtypes.VoteOption_name[int32(govtypes.OptionAbstain)]:
			abstain = oc.Count
			// case types.YES:
			// 	yes = oc.Count
			// case types.NO:
			// 	no = oc.Count
			// case types.NOWITHVETO:
			// 	noWithVeto = oc.Count
			// case types.ABSTAIN:
			// 	abstain = oc.Count
		}
	}

	return yes, no, noWithVeto, abstain, nil
}

func (db *Database) QueryGetLiveProposal() ([]mdschema.Proposal, error) {

	props := make([]mdschema.Proposal, 0)

	err := db.Model(&props).
		Where("proposal_status = 'PROPOSAL_STATUS_VOTING_PERIOD' and voting_end_time < now()").
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return props, nil
		}
		return nil, err
	}

	return props, nil
}
