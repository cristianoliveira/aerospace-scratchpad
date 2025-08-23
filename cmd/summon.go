/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/spf13/cobra"
)

func SummonCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "summon <pattern>",
		Short: "Summon a window from scratchpad",
		Long: `Summon a window from the scratchpad to the current workspace.

This command brings a window from the scratchpad to the current workspace using a regex to match the window name or title.
If no pattern is provided, it summons the first window in the scratchpad.
`,

		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			cli.ValidateAllNonEmpty,
		),

		Run: func(cmd *cobra.Command, args []string) {
			windowNamePattern := strings.TrimSpace(args[0])

			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				stderr.Println("Error: unable to get focused workspace")
				return
			}

			// Parse filter flags
			filterFlags, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				stderr.Println("Error: unable to get filter flags")
				return
			}

			// Filter windows using the shared querier
			querier := aerospace.NewAerospaceQuerier(aerospaceClient)
			mover := aerospace.NewAeroSpaceMover(aerospaceClient)

			windows, err := querier.GetFilteredWindows(windowNamePattern, filterFlags)
			if err != nil {
				stderr.Println("Error: %v", err)
				return
			}

			for _, window := range windows {
				setFocus := true
				err := mover.MoveWindowToWorkspace(
					&window,
					focusedWorkspace,
					setFocus,
				)
				if err != nil {
					stderr.Println("Error: %v", err)
					return
				}
			}
		},
	}

	// Filter flags --filter
	command.Flags().StringArrayP("filter", "F", []string{}, "Filter windows by a specific property (e.g., app-name, window-title). Can be used multiple times.")

	return command
}
