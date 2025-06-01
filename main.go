/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package main

import (
	"fmt"
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
		if err := defaultLogger.Close(); err != nil {
			log.Fatalf("Error: closing logger\n%v", err)
		}
	}()
	logger.SetDefaultLogger(defaultLogger)
	defaultLogger.LogInfo("Executing Aerospace Scratchpad CLI")

	aerospaceMarkClient, err := aerospacecli.NewAeroSpaceConnection()
	if err != nil {
		fmt.Println("Error creating Aerospace client:", err)
	}

	cmd.Execute(aerospaceMarkClient)
}
