package cmd

import (
	"log/slog"

	"github.com/Consoleaf/kopl/utils"
)

var logger *slog.Logger

func init() {
	logger = slog.New(utils.NewSimpleCLIHandler())
}
