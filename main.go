/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package main

import (
	"log"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

func main() {
	defaultLogger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Error: creating logger\n%v", err)
		return
	}
	defer func() {
		if closeErr := defaultLogger.Close(); closeErr != nil {
			log.Fatalf("Error: closing logger\n%v", closeErr)
		}
	}()
	logger.SetDefaultLogger(defaultLogger)
	defaultLogger.LogInfo("Executing Aerospace Scratchpad CLI")

	aerospaceMarkClient, err := aerospacecli.NewAeroSpaceClient()
	if err != nil {
		log.Printf("Error creating Aerospace client: %v", err)
	}

	cmd.Execute(aerospaceMarkClient)
}
