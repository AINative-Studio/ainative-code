package main

import (
	"os"

	"github.com/AINative-studio/ainative-code/internal/cmd"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

func main() {
	// Initialize global logger
	logger.Init()

	// Execute root command
	if err := cmd.Execute(); err != nil {
		logger.ErrorEvent().Err(err).Msg("Failed to execute command")
		os.Exit(1)
	}
}
