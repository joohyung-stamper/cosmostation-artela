package exporter

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cosmostation/cosmostation-coreum/app"
	"github.com/cosmostation/cosmostation-coreum/custom"
	"go.uber.org/zap"

	// mbl
	mdschema "github.com/cosmostation/mintscan-database/schema"

	// sdk
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

var (
	// Version is a project's version string.
	Version = "Development"

	// Commit is commit hash of this project.
	Commit = ""

	controler     = make(chan struct{}, 60)
	wg            = new(sync.WaitGroup)
	initialHeight = int64(0)
)

// Exporter is
type Exporter struct {
	*app.App
}

// NewExporter returns new Exporter instance
func NewExporter(a *app.App) *Exporter {
	return &Exporter{a}
}

// preProcess 는 실제 프로세스 수행 전, 필요한 설정 환경 등을 동적으로 설정
func SetInitialHeight(height int64) {
	initialHeight = height
	zap.S().Debugf("InitialHeight : %d\n", initialHeight)
}

// Start starts to synchronize blockchain data
func (ex *Exporter) Start(op int) {
	zap.S().Info("Starting Chain Exporter...")
	zap.S().Infof("Version: %s | Commit: %s", Version, Commit)
	zap.S().Infof("Schema Info : %s, %s\n", mdschema.GetCommonSchema(), mdschema.GetChainSchema())

	tick10Sec := time.NewTicker(time.Second * 10)
	tick20Min := time.NewTicker(time.Minute * 20)

	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			zap.S().Info("start - sync blockchain")
			err := ex.sync(op)
			if err != nil {
				zap.S().Infof("error - sync blockchain: %s\n", err)
			}
			zap.S().Info("finish - sync blockchain")

			time.Sleep(time.Second)
		}
	}()
	// go ex.runFeeMaker()
	// app init 시 최초 전체 프로포절 업데이트
	ex.saveAllProposals()
	go ex.watchLiveProposals()
	go ex.updateProposals()

	if op == BASIC_MODE {
		go func() {
			for {
				select {
				case <-tick10Sec.C:
					zap.S().Info("start sync validators")
					ex.saveValidators()
					// ex.saveLiveProposals()
					zap.S().Info("finish sync validators")
				case <-tick20Min.C:
					zap.S().Info("start sync validators keybase identities")
					ex.saveValidatorsIdentities()
					zap.S().Info("finish sync validators keybase identities")
				case <-done:
					return
				}
			}
		}()
	}

	<-done // implement gracefully shutdown when signal received
	zap.S().Infof("shutdown signal received")
}

// sync compares block height between the height saved in your database and
// the latest block height on the active chain and calls process to start ingesting data.
func (ex *Exporter) sync(op int) error {
	// Query latest block height saved in database
	dbHeight, err := ex.DB.GetLatestBlockHeight(ex.ChainIDMap[ex.Config.Chain.ChainID])
	if dbHeight == -1 {
		return fmt.Errorf("unexpected error in database: %s", err)
	}
	rawDBHeight, err := ex.RawDB.GetLatestBlockHeight()
	if rawDBHeight == -1 {
		return fmt.Errorf("unexpected error in database: %s", err)
	}

	// Query latest block height on the active network
	latestBlockHeight, err := ex.Client.RPC.GetLatestBlockHeight()
	if latestBlockHeight == -1 {
		return fmt.Errorf("failed to query the latest block height on the active network: %s", err)
	}

	if dbHeight == 0 && initialHeight != 0 {
		dbHeight = initialHeight - 1
		rawDBHeight = initialHeight - 1
		zap.S().Info("initial Height set : ", initialHeight)
	}

	beginHeight := dbHeight
	if dbHeight > rawDBHeight || op == RAW_MODE {
		beginHeight = rawDBHeight
	}
	zap.S().Infof("dbHeight %d, rawHeight %d \n", dbHeight, rawDBHeight)

	for h := beginHeight + 1; h <= latestBlockHeight; h++ {
		// block, txs, err := ex.Client.RPC.GetBlockAndTxsFromNode(custom.EncodingConfig.Marshaler, h)
		block, txs, err := ex.Client.RPC.GetBlockAndTxsFromNode(custom.EncodingConfig.Codec, h)
		if err != nil {
			return fmt.Errorf("failed to get block and txs : %s", err)
		}

		switch op {
		case BASIC_MODE:
			if h > dbHeight {
				err = ex.process(block, txs, op)
				if err != nil {
					return err
				}
			}
			fallthrough //continue to case RAW_MODE
		case RAW_MODE:
			if h > rawDBHeight {
				err = ex.rawProcess(block, txs)
				if err != nil {
					return err
				}
			}
		case REFINE_MODE:
		default:
			zap.S().Info("unknown mode = ", op)
			os.Exit(1)
		}
		zap.S().Infof("synced block %d/%d", h, latestBlockHeight)
	}
	return nil
}

func (ex *Exporter) rawProcess(block *tmctypes.ResultBlock, txs []*sdktypes.TxResponse) (err error) {
	rawData := new(mdschema.RawData)

	rawData.Block, err = ex.getRawBlock(block)
	if err != nil {
		return fmt.Errorf("failed to get block: %s", err)
	}
	// rawData.Transactions, err = ex.getRawTransactions(block, txs)
	// if err != nil {
	// 	return fmt.Errorf("failed to get txs: %s", err)
	// }
	return ex.RawDB.InsertExportedData(rawData)
}

// process ingests chain data, such as block, transaction, validator, evidence information and
// save them in database.
func (ex *Exporter) process(block *tmctypes.ResultBlock, txs []*sdktypes.TxResponse, op int) (err error) {
	basic := new(mdschema.BasicData)

	basic.Block, err = ex.getBlock(block)
	if err != nil {
		return fmt.Errorf("failed to get block: %s", err)
	}

	if time.Since(basic.Block.Timestamp.UTC()).Seconds() > 60 {
		ex.App.CatchingUp = true
	} else {
		ex.App.CatchingUp = false
	}

	basic.Evidence, err = ex.getEvidence(block)
	if err != nil {
		return fmt.Errorf("failed to get evidence: %s", err)
	}

	if block.Block.LastCommit.Height != 0 {
		prevBlock, err := ex.Client.RPC.GetBlock(block.Block.LastCommit.Height)
		if err != nil {
			return fmt.Errorf("failed to query previous block: %s", err)
		}

		vals, err := ex.Client.RPC.GetValidatorsInHeight(block.Block.LastCommit.Height)
		if err != nil {
			return fmt.Errorf("failed to query validators: %s", err)
		}

		basic.GenesisValidatorsSet, err = ex.getGenesisValidatorsSet(block, vals)
		if err != nil {
			return fmt.Errorf("failed to get genesis validator set: %s", err)
		}
		basic.MissBlocks, basic.AccumulatedMissBlocks, basic.MissDetailBlocks, err = ex.getValidatorsUptime(prevBlock, block, vals)
		if err != nil {
			return fmt.Errorf("failed to get missing blocks: %s", err)
		}
	}

	if basic.Block.NumTxs > 0 {
		basic.ChainInfo, err = ex.DB.GetCurrentChainInfo(ex.Config.Chain.ChainID)
		if err != nil {
			return fmt.Errorf("failed to get current chaininfo: %s", err)
		}
		basic.ChainInfo.NumberOfTxs += basic.Block.NumTxs

		basic.Proposals, basic.Deposits, basic.Votes, err = ex.getGovernance(&block.Block.Header.Time, txs)
		if err != nil {
			return fmt.Errorf("failed to get governance: %s", err)
		}
		// exportData.ValidatorsPowerEventHistory, err = ex.getPowerEventHistory(block, txs)
		basic.ValidatorsPowerEventHistory, err = ex.getPowerEventHistoryNew(txs)
		if err != nil {
			return fmt.Errorf("failed to get transactions: %s", err)
		}

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

	// TODO: is this right place to be?
	if ex.Config.Alarm.Switch {
		ex.handlePushNotification(block, txs)
	}

	return ex.DB.InsertExportedData(basic)
}
