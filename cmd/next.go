/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"github.com/spf13/cobra"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
)

// NextCmd represents the next command.
func NextCmd(aerospaceClient aerospacecli.AeroSpaceClient) *cobra.Command {
	nextCmd := &cobra.Command{
		Use:   "next",
		Short: "Shows the next scratchpad window",
		Long: `Shows the next scratchpad window in the current workspace.

This command cycles through the scratchpad windows, displaying them in the current workspace.
It does not send the windows back to the scratchpad, but rather focuses the next available scratchpad window.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				stderr.Println(
					"Error: unable to get focused workspace\n%s",
					err,
				)
				return
			}

			querier := aerospace.NewAerospaceQuerier(aerospaceClient)
			mover := aerospace.NewAeroSpaceMover(aerospaceClient)

			window, err := querier.GetNextScratchpadWindow()
			if err != nil {
				stderr.Println("Error: %v", err)
				return
			}

			setFocus := true
			if moveErr := mover.MoveWindowToWorkspace(
				window,
				focusedWorkspace,
				setFocus,
			); moveErr != nil {
				stderr.Println("Error: %v", moveErr)
				return
			}
		},
	}

	return nextCmd
}
