/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package main

import (
	"fmt"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
)

func main() {
	aerospaceMarkClient, err := aerospacecli.NewAeroSpaceConnection()
	if err != nil {
		fmt.Println("Error creating Aerospace client:", err)
	}

	cmd.Execute(aerospaceMarkClient)
}
