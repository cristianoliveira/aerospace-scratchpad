/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/spf13/cobra"
)

// NextCmd represents the next command
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
			window, err := querier.GetNextScratchpadWindow()
			if err != nil {
				fmt.Println(err)
				stderr.Println("Error: unable to get next scratchpad window")
				return
			}

			if err := aerospaceClient.MoveWindowToWorkspace(
				window.WindowID,
				focusedWorkspace.Workspace,
			); err != nil {
				stderr.Printf("Error: unable to move window '%+v' to workspace '%s'\n", window, focusedWorkspace.Workspace)
				return
			}

			if err = aerospaceClient.SetFocusByWindowID(window.WindowID); err != nil {
				stderr.Printf("Error: unable to set focus to window '%+v'\n", window)
				return
			}

			fmt.Printf(
				"Next scratchpad window '%s' focused in workspace '%s'\n",
				window.AppName,
				focusedWorkspace.Workspace,
			)
		},
	}

	return nextCmd
}
