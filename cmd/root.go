/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"os"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
func RootCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "aerospace-scratchpad",
		Short: "Scratchpad for AeroSpace WM",
		Long: `Scratchpad for AeroSpace WM

Allows you manage your windows in a scratchpad-like manner.
This is heavily inspired by i3wm's scratchpad feature, but follows aerospace command line conventions.

See:
https://i3wm.org/docs/userguide.html#_scratchpad
`,
		Version: VERSION,
	}

	// Commands
	rootCmd.AddCommand(MoveCmd(aerospaceClient))
	rootCmd.AddCommand(ShowCmd(aerospaceClient))
	rootCmd.AddCommand(SummonCmd(aerospaceClient))
	rootCmd.AddCommand(NextCmd(aerospaceClient))
	rootCmd.AddCommand(InfoCmd(aerospaceClient))

	return rootCmd
}

func Execute(
	aerospaceClient aerospacecli.AeroSpaceClient,
) {
	rootCmd := RootCmd(aerospaceClient)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// THIS IS GENERATED DON'T EDIT
// NOTE: to update VERSION change it to an EMPTY STRING
// and then run scripts/validate-version.sh
var VERSION = "v0.0.2"
