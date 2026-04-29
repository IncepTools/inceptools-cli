package ui_test

import (
	"github.com/IncepTools/inceptools-cli/src/ui"
	"testing"
)

func TestUIFunctions(t *testing.T) {
	// These functions just print to stdout/stderr.
	// We just call them to ensure no panics and for basic coverage.
	ui.Success("Test success")
	ui.Info("Test info")
	ui.Warn("Test warn")
	ui.Error("Test error")
	ui.Heading("Test heading")
	ui.Finished("Test finished")
}
