/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
)

// NextCmd represents the next command.
func NextCmd(aerospaceClient *aerospace.AeroSpaceClient) *cobra.Command {
	nextCmd := &cobra.Command{
		Use:   "next",
		Short: "Cycles through scratchpad windows",
		Long: `Cycles through scratchpad windows on the selected monitor.

Scratchpad windows are ordered by window ID. Starting after the focused window, next prefers the
first window outside the current workspace to avoid repeating a no-op. If all candidates are already
there, it uses normal successor-and-wrap order, starting at the first window without a focus cursor.
The selected window moves to the current workspace and becomes the cursor for the next invocation.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			outputFormat, err := cmd.Flags().GetString("output")
			if err != nil {
				stderr.Println("Error: unable to get output format")
				return
			}
			formatter, err := cli.NewOutputFormatter(os.Stdout, outputFormat)
			if err != nil {
				stderr.Println("Error: unsupported output format")
				return
			}

			monitorID, err := parseMonitorFlag(cmd)
			if err != nil {
				stderr.Printf("Error: %v\n", err)
				return
			}

			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				stderr.Println(
					"Error: unable to get focused workspace\n%s",
					err,
				)
				return
			}

			querier := aerospace.NewAerospaceQuerier(aerospaceClient.GetUnderlyingClient())
			mover := aerospace.NewAeroSpaceMover(aerospaceClient)

			window, err := querier.GetNextScratchpadWindowForMonitor(
				monitorID,
				focusedWorkspace.Workspace,
			)
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

			if printErr := formatter.Print(cli.OutputEvent{
				Command:         commandNext,
				Action:          actionToWorkspace,
				WindowID:        window.WindowID,
				AppName:         window.AppName,
				Workspace:       window.Workspace,
				TargetWorkspace: focusedWorkspace.Workspace,
				Result:          "ok",
			}); printErr != nil {
				stderr.Println("Error: %v", printErr)
			}
		},
	}

	return nextCmd
}
