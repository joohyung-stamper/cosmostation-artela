package exporter

import (
	"fmt"

	tmjson "github.com/cometbft/cometbft/libs/json"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"
)

func (ex *Exporter) getBlock(block *tmctypes.ResultBlock) (*mdschema.Block, error) {
	// chunk, err := json.Marshal(block)
	// if err != nil {
	// 	return &mdschema.Block{}, fmt.Errorf("failed to marshal block : %s", err)
	// }
	ph := block.Block.Header.LastBlockID.Hash.String()
	ns := int64(0)
	for i := range block.Block.LastCommit.Signatures {
		if block.Block.LastCommit.Signatures[i].Signature != nil {
			ns++
		}
	}
	if block.Block.Height == 1 || block.Block.Height == initialHeight {
		ph = "genesis"
	}
	b := &mdschema.Block{
		ChainInfoID:   ex.ChainIDMap[block.Block.Header.ChainID],
		Height:        block.Block.Height,
		Proposer:      block.Block.ProposerAddress.String(),
		Hash:          block.BlockID.Hash.String(),
		ParentHash:    ph,
		NumSignatures: ns,
		NumTxs:        int64(len(block.Block.Data.Txs)),
		// Chunk:         chunk,
		Timestamp: block.Block.Time,
	}

	return b, nil
}

// getBlock exports block information
func (ex *Exporter) getBlockFromDB(rawBlocks []mdschema.RawBlock) (blocks []mdschema.Block, err error) {

	/*
		block unmarshal
		참고 : tmjson "github.com/cometbft/cometbft/libs/json"
	*/

	for i := range rawBlocks {
		var block tmctypes.ResultBlock
		err := tmjson.Unmarshal(rawBlocks[i].Chunk, &block)
		if err != nil {
			return blocks, fmt.Errorf("failed to marshal block : %s", err)
		}
		ph := block.Block.Header.LastBlockID.Hash.String()
		ns := int64(0)
		for i := range block.Block.LastCommit.Signatures {
			if block.Block.LastCommit.Signatures[i].Signature != nil {
				ns++
			}
		}
		if block.Block.Height == 1 || block.Block.Height == initialHeight {
			ph = "genesis"
		}
		b := mdschema.Block{
			ChainInfoID:   ex.ChainIDMap[block.Block.Header.ChainID],
			Height:        block.Block.Height,
			Proposer:      block.Block.ProposerAddress.String(),
			Hash:          block.BlockID.Hash.String(),
			ParentHash:    ph,
			NumSignatures: ns,
			NumTxs:        int64(len(block.Block.Data.Txs)),
			// Chunk:         rawBlocks[i].Chunk,
			Timestamp: block.Block.Time,
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

// getRawBlock decodes transactions in a block and return a format of database transaction.
func (ex *Exporter) getRawBlock(block *tmctypes.ResultBlock) (*mdschema.RawBlock, error) {
	b := new(mdschema.RawBlock)
	chunk, err := tmjson.Marshal(block)
	if err != nil {
		return &mdschema.RawBlock{}, fmt.Errorf("failed to marshal block : %s", err)
	}
	b.ChainID = block.Block.ChainID
	b.Height = block.Block.Height
	b.BlockHash = block.BlockID.Hash.String()
	b.NumTxs = int64(len(block.Block.Data.Txs))
	b.Chunk = chunk

	return b, nil
}
