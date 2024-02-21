package custom

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	SetAppConfig()

	os.Exit(m.Run())
}
