package main

import (
	"flag"
	"log"
	"os"

	"github.com/cosmostation/cosmostation-coreum/app"
	"github.com/cosmostation/cosmostation-coreum/exporter"
	"go.uber.org/zap"
)

func main() {
	mode := flag.String("mode", "basic", "chain-exporter mode \n  - basic : default, will store current chain status\n  - raw : will only store jsonRawMessage of block and transaction to database\n  - refine : refine new data from database the legacy chain stored\n  - genesis : extract genesis state from the given file")
	initialHeight := flag.Int64("initial-height", 0, "initial height of chain-exporter to sync")
	genesisFilePath := flag.String("genesis-file-path", "", "absolute path of genesis.json")
	flag.Parse()

	log.Println("mode : ", *mode)
	log.Println("genesis-file-path : ", *genesisFilePath)
	log.Println("initial-height :", *initialHeight)

	fileBaseName := "chain-exporter"
	cApp := app.NewApp(fileBaseName)

	exporter.SetInitialHeight(*initialHeight)
	ex := exporter.NewExporter(cApp)
	ex.SetChainID()

	switch *mode {
	case "basic": //기본 동작
		ex.Start(exporter.BASIC_MODE)
	case "raw":
		ex.Start(exporter.RAW_MODE)
	case "refine":
		if err := ex.Refine(exporter.REFINE_MODE); err != nil {
			zap.S().Error(err)
		}
		zap.S().Info("refine successfully complete")
	case "genesis":
		if err := ex.GetGenesisStateFromGenesisFile(*genesisFilePath); err != nil {
			zap.S().Error(err)
			os.Exit(1)
		}
		zap.S().Info("genesis file parsing complete")
	default:
		log.Println("Unknow operator type :", *mode)
	}

}
