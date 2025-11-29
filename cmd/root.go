/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
)

// RootCmd represents the base command when called without any subcommands.
func RootCmd(
	aerospaceClient aerospace.AeroSpaceWMClient,
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

	// Global Flags
	rootCmd.PersistentFlags().
		BoolP("dry-run", "n", false, "Run the command without moving windows (dry run mode)")

	// Create custom client wrapper - now works with interface
	customClient := aerospace.NewAeroSpaceClient(aerospaceClient)
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		dry, _ := cmd.Flags().GetBool("dry-run")
		customClient.SetOptions(aerospace.ClientOpts{
			DryRun: dry,
		})
	}

	// Commands
	rootCmd.AddCommand(enableFilterFlag(MoveCmd(customClient)))
	rootCmd.AddCommand(enableFilterFlag(ShowCmd(customClient)))
	rootCmd.AddCommand(enableFilterFlag(SummonCmd(customClient)))
	rootCmd.AddCommand(NextCmd(customClient))
	rootCmd.AddCommand(InfoCmd(aerospaceClient))
	rootCmd.AddCommand(HookCmd(aerospaceClient))

	return rootCmd
}

func enableFilterFlag(command *cobra.Command) *cobra.Command {
	command.Flags().StringArrayP(
		"filter", "F", []string{},
		`Filter windows by a specific property (e.g. window-title=^foo).
Requires a key=value format. Can be used multiple times. `,
	)
	return command
}

func Execute(
	aerospaceClient aerospace.AeroSpaceWMClient,
) {
	rootCmd := RootCmd(aerospaceClient)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// VERSION The CLI current version
// THIS IS GENERATED DON'T EDIT
// NOTE: to update VERSION change it to an EMPTY STRING
// and then run scripts/validate-version.sh.
//
//nolint:gochecknoglobals // version is overridden via build flags
var VERSION = "v0.3.0"
