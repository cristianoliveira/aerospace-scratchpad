/*
Copyright © 2025 Cristian Oliveira licence@cristianoliveira.dev
*/
package cmd

import (
	"os"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
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
	}

	// Commands
	rootCmd.AddCommand(MoveCmd(aerospaceClient))
	rootCmd.AddCommand(ShowCmd(aerospaceClient))
	rootCmd.AddCommand(SummonCmd(aerospaceClient))

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
// NOTE: to update VERSION to empty string 
// and then run scripts/validate-version.sh
// var VERSION = "v0.0.1-20250518-16d72bb"
