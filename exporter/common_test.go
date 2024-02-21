package exporter

import (
	"os"
	"testing"

	"github.com/cosmostation/cosmostation-coreum/app"
)

var (
	ex *Exporter
)

func TestMain(m *testing.M) {
	chainEx := app.NewApp("chain-exporter")
	ex = NewExporter(chainEx)

	os.Exit(m.Run())
}
