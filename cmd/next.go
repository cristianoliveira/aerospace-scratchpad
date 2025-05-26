/*
Copyright © 2025 Cristian Oliveira licence@cristianoliveira.dev
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

		This command will cycle through the scratchpad windows showing in them in current workspace.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()

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
			return
		},
	}

	return nextCmd
}
