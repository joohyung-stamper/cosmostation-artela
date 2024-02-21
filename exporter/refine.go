package exporter

import (
	"fmt"
	"time"

	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmostation/cosmostation-coreum/custom"
	mdschema "github.com/cosmostation/mintscan-database/schema"
	"go.uber.org/zap"
)

func (ex *Exporter) Refine(op int) error {
	var chainID string
	aminoUnmarshal := custom.AppCodec.UnmarshalJSON

	// 프로그램이 기동 되고, rawdb로부터 동기화 할 목표 높이
	srcRawBlockHeight, err := ex.RawDB.GetLatestBlockHeight()
	if err != nil {
		zap.S().Info("raw_block have no blocks to refine")
		return err
	}

	// db에 저장된 블록 높이 chain_info_id = x order by id desc limit 1
	dstCurrentBlockHeight, err := ex.DB.GetLatestBlockHeight(ex.ChainIDMap[ex.Config.Chain.ChainID])
	if err != nil {
		zap.S().Info("failed to get latest block height")
		return err
	}

	// db에 저장된 마지막 transaction
	dstCurrentTransaction, err := ex.DB.GetLatestTransaction(ex.ChainIDMap[ex.Config.Chain.ChainID])
	if err != nil {
		zap.S().Info("failed to get latest transaction")
		return err
	}

	if srcRawBlockHeight > dstCurrentBlockHeight {
		beginRawBlock, err := ex.RawDB.GetBlockByHeight(dstCurrentBlockHeight)
		beginRawBlockID := beginRawBlock.ID + 1
		rawBlockIDMax, err := ex.RawDB.GetBlockIDMax(srcRawBlockHeight)
		if rawBlockIDMax == -1 {
			return fmt.Errorf("fail to get block max(id) in database: %s", err)
		}

		zap.S().Infof("will be refine blocks based on id from %d to %d\n", beginRawBlockID, rawBlockIDMax)
		for i := beginRawBlockID; i <= rawBlockIDMax; i++ {
			zap.S().Info("block working id : ", i)
			rb, err := ex.RawDB.GetBlockByID(i, srcRawBlockHeight)
			if err != nil {
				return err
			}

			if rb == nil {
				return nil
			}
			i = rb[len(rb)-1].ID // 최종적으로 받아온 id 값으로 다음 loop를 시작한다.

			if err := ex.refineRawBlocks(rb); err != nil {
				return err
			}
		}
	}

	if srcRawBlockHeight > dstCurrentTransaction.Height {
		beginRawTransaction, err := ex.RawDB.GetTransactionByHash(dstCurrentTransaction.Height, dstCurrentTransaction.Hash)
		beginRawTransactionID := beginRawTransaction.ID + 1
		rawTxIDMax, err := ex.RawDB.GetTxIDMax(srcRawBlockHeight)
		if rawTxIDMax == -1 {
			return fmt.Errorf("fail to get transaction max(id) in database: %s", err)
		}
		zap.S().Infof("will be refine transactions based on id from %d to %d\n", beginRawTransactionID, rawTxIDMax)
		for i := beginRawTransactionID; i <= rawTxIDMax; i++ {
			zap.S().Info("tx working id : ", i)

			ts, err := ex.RawDB.GetTransactionsByID(i, srcRawBlockHeight)
			if err != nil {
				return err
			}
			if ts == nil {
				return nil
			}
			txs := make([]*sdktypes.TxResponse, len(ts))
			for j, t := range ts {
				tx := new(sdktypes.TxResponse)
				if err := aminoUnmarshal(t.Chunk, tx); err != nil {
					return err
				}
				txs[j] = tx
				chainID = t.ChainID
				i = t.ID // 최종적으로 받아온 id 값으로 다음 loop를 시작한다.
			}

			if err := ex.refineRawTransactions(chainID, txs); err != nil {
				return err
			}
		}
	}

	// 실시간 refine
	for {
		if err := ex.refineSync(); err != nil {
			zap.S().Infof("error - sync blockchain: %s\n", err)
		}
		time.Sleep(2 * time.Second)
	}
}
func (ex *Exporter) refineRawTransactions(chainID string, txs []*sdktypes.TxResponse) (err error) {
	refineData := new(mdschema.RefineData)

	//cosmoshub-1에서는 txResponse에 timestamp를 가지고 있지 않기 때문에, 블록으로부터 가져온다.
	// 시작
	var begin, end int64
	if len(txs) > 0 {
		begin = txs[0].Height
		end = txs[len(txs)-1].Height

	}
	blockList, err := ex.getBlocksHasTxs(chainID, begin, end)
	if err != nil {
		return fmt.Errorf("failed to get block list: %s", err)
	}
	// 추가 로직 종료

	refineData.Transactions, err = ex.getTxs(chainID, blockList, txs, true)
	if err != nil {
		return fmt.Errorf("failed to get txs: %s", err)
	}

	refineData.TMAs = ex.disassembleTransaction(txs)

	if true {
		return ex.DB.InsertRefineData(refineData)
	}
	return fmt.Errorf("currently, disabled to store data into database\n")

}

// getBlockHasTxs()는 database로부터 chainID, 시작 블록 높이, 종료 블록 높이의 범위 내에서 []블록 슬라이스를 리턴(id, height, timestamp)를 리턴한다.
// schema.Block의 나머지 구조체는 nil 임에 주의한다.
func (ex *Exporter) getBlocksHasTxs(chainID string, begin, end int64) (list map[int64]*mdschema.Block, err error) {
	blocks, err := ex.DB.GetBlockHasTxs(ex.ChainIDMap[chainID], begin, end)
	if err != nil {
		return list, err
	}

	list = map[int64]*mdschema.Block{} // 초기화

	for i := range blocks {
		list[blocks[i].Height] = &blocks[i]
	}
	return list, nil
}

func (ex *Exporter) refineRawBlocks(block []mdschema.RawBlock) (err error) {
	refineData := new(mdschema.RefineData)

	refineData.Blocks, err = ex.getBlockFromDB(block)
	if err != nil {
		return fmt.Errorf("failed to get block: %s", err)
	}

	if true {
		return ex.DB.InsertRefineData(refineData)
	}
	return fmt.Errorf("currently do not store any data\n")

}

func (ex *Exporter) refineSync() error {
	// Query latest block height saved in database
	dbHeight, err := ex.DB.GetLatestBlockHeight(ex.ChainIDMap[ex.Config.Chain.ChainID])
	if dbHeight == -1 {
		return fmt.Errorf("unexpected error in database: %s", err)
	}

	// Query latest block height on the active network
	latestBlockHeight, err := ex.Client.RPC.GetLatestBlockHeight()
	if latestBlockHeight == -1 {
		return fmt.Errorf("failed to query the latest block height on the active network: %s", err)
	}

	if dbHeight == 0 && initialHeight != 0 {
		dbHeight = initialHeight - 1
		zap.S().Info("initial Height set : ", initialHeight)
	}

	beginHeight := dbHeight

	zap.S().Infof("dbHeight %d\n", dbHeight)

	for i := beginHeight + 1; i <= latestBlockHeight; i++ {
		block, err := ex.Client.RPC.GetBlock(i)
		if err != nil {
			return fmt.Errorf("failed to query block: %s", err)
		}
		retryFlag := false
		zap.S().Infof("number of Transactions : %d", len(block.Block.Txs))
		txList := block.Block.Txs
		txs := make([]*sdktypes.TxResponse, len(block.Block.Txs))

		for idx, tx := range txList {
			hex := fmt.Sprintf("%X", tx.Hash())
			controler <- struct{}{}
			wg.Add(1)
			go func(i int, hex string) {
				zap.S().Info(i, hex)
				defer func() {
					<-controler
					wg.Done()
				}()

				txs[i], err = ex.Client.CliCtx.GetTx(hex)
				if err != nil {
					zap.S().Error("Error while getting tx ", hex)
					retryFlag = true
					return
				}
			}(idx, hex)
		}
		wg.Wait()

		if retryFlag {
			zap.S().Error("can not get all of tx, retry get tx in block height = ", i)
			i--
			time.Sleep(1 * time.Second)
			continue
		}

		if i > dbHeight {
			err = ex.refineRealTimeprocess(block, txs)
			if err != nil {
				return err
			}
		}
		zap.S().Infof("synced block %d/%d", i, latestBlockHeight)
	}
	return nil
}
func (ex *Exporter) refineRealTimeprocess(block *tmctypes.ResultBlock, txs []*sdktypes.TxResponse) (err error) {
	basic := new(mdschema.BasicData)

	basic.Block, err = ex.getBlock(block)
	if err != nil {
		return fmt.Errorf("failed to get block: %s", err)
	}

	if basic.Block.NumTxs > 0 {
		// 시작
		// block-id 추출을 위해 사용
		list := make(map[int64]*mdschema.Block)
		list[block.Block.Height] = basic.Block
		// 종료

		basic.Transactions, err = ex.getTxs(block.Block.ChainID, list, txs, false)
		if err != nil {
			return fmt.Errorf("failed to get txs: %s", err)
		}
		basic.TMAs = ex.disassembleTransaction(txs)
	}

	return ex.DB.InsertRefineRealTimeData(basic)
}
