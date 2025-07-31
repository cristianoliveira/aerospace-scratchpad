/*
Copyright © 2025 Cristian Oliveira licence@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"time"

	aerospaceipc "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/spf13/cobra"
)

func WorkspaceHandlerCmd(
	aerospaceClient aerospaceipc.AeroSpaceClient,
) *cobra.Command {
	var wsHandlerCmd = &cobra.Command{
		Use:   "workspace-handler <workspace>",
		Short: "This command handles when a window in scratchpad is focused (which shouldn't happen)",
		Long: `This command handles when a window within the scratchpad workspace is focused. It'll move the window to the last focused workspace and take the window too, behaving more like "summoning the window".
Add this snippet in your ~/aerospace.toml config:
'''toml
exec-on-workspace-change = ['/bin/bash', '-c',
    'aerospace-scratchpad workspace-handler $AEROSPACE_FOCUSED_WORKSPACE'
]
'''
`,
    Aliases: []string{"wsh"},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			client := aerospaceClient.Connection()

			focusedWorkspace := args[0]
			if focusedWorkspace != constants.DefaultScratchpadWorkspaceName {
				logger.LogDebug(
					"WSH: focused workspace is not scratchpad",
					"workspace", focusedWorkspace,
				)
				return
			}

			logger.LogInfo("WSH: focused workspace is scratchpad")

			focusedWindow, err := aerospaceClient.GetFocusedWindow()
			if err != nil {
				logger.LogError("WS: unable to get focused window", "error", err)
				fmt.Println("Error: unable to get focused window")
				return
			}

			logger.LogInfo("WSH: focused window", "window", focusedWindow)

			if focusedWindow.Workspace == constants.DefaultScratchpadWorkspaceName {
				_, err := client.SendCommand("workspace-back-and-forth", nil)
				if err != nil {
					logger.LogError("WSH: unable to get focused window", "error", err)
					fmt.Println("Error: unable to get focused window")
					return
				}

				newFocusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
				if err != nil {
					logger.LogError("WSH: unable to get focused workspace after moving window", "error", err)
					fmt.Println("Error: unable to get focused workspace after moving window", err)
					return
				}

				logger.LogInfo(
					"WSH: new focused workspace after moving window",
					"workspace", newFocusedWorkspace.Workspace,
				)

				for {
					if newFocusedWorkspace.Workspace != constants.DefaultScratchpadWorkspaceName {
						break
					}

					logger.LogInfo("Focused workspace is still scratchpad, waiting for it to change...")
					time.Sleep(100 * time.Microsecond) // Sleep for 5 seconds before the next iteration
					newFocusedWorkspace, err = aerospaceClient.GetFocusedWorkspace()
					if err != nil {
						logger.LogError("WSH: unable to get focused workspace after moving window", "error", err)
						fmt.Println("Error: unable to get focused workspace after moving window", err)
					}
				}

				_, err = client.SendCommand(
					"move-node-to-workspace",
					[]string{
						newFocusedWorkspace.Workspace,
						"--window-id", fmt.Sprintf("%d", focusedWindow.WindowID),
						"--focus-follows-window",
					},
				)
				if err != nil {
					fmt.Printf(
						"Error: unable to move window '%+v' to workspace '%s': %v\n",
						focusedWindow,
						focusedWorkspace,
						err,
					)
				}
			}
		},
	}

	return wsHandlerCmd
}
