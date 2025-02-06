package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/cmd"
)

func main() {
	consoleOutput := zerolog.ConsoleWriter{Out: os.Stderr}
	loggerOutput := consoleOutput
	logger := zerolog.New(loggerOutput)
	log.Logger = logger
	cmd.Execute()
}
