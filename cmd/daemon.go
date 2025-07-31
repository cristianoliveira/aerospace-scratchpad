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

func DaemonCmd(
	aerospaceClient aerospaceipc.AeroSpaceClient,
) *cobra.Command {
	var daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Daemon to manage focus on scratchpad workspace",
		Long: `This command runs a daemon that monitors the focused workspace and moves windows
from the scratchpad workspace to the focused workspace when necessary.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			client := aerospaceClient.Connection()

			// loop to keep the daemon running
			for {
				focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
				if err != nil {
					logger.LogError("SHOW: unable to get focused workspace", "error", err)
					fmt.Println("Error: unable to get focused workspace")
					continue
				}

				// Check if is .scratchpad workspace
				if focusedWorkspace.Workspace == constants.DefaultScratchpadWorkspaceName {
					logger.LogInfo("DAEMON: focused workspace is scratchpad")

					focusedWindow, err := aerospaceClient.GetFocusedWindow()
					if err != nil {
						logger.LogError("SHOW: unable to get focused window", "error", err)
						fmt.Println("Error: unable to get focused window")
						continue
					}

					logger.LogInfo(
						"DAEMON: focused window",
						"workspace", focusedWindow.Workspace,
						"window", focusedWindow,
					)

					if focusedWindow.Workspace == focusedWorkspace.Workspace {
						_, err := client.SendCommand("workspace-back-and-forth", nil)
						if err != nil {
							logger.LogError("SHOW: unable to get focused window", "error", err)
							fmt.Println("Error: unable to get focused window")
							continue
						}

						newFocusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
						if err != nil {
							logger.LogError("SHOW: unable to get focused workspace after moving window", "error", err)
							fmt.Println("Error: unable to get focused workspace after moving window", err)
						}

						logger.LogInfo(
							"DAEMON: new focused workspace after moving window",
							"workspace", newFocusedWorkspace.Workspace,
						)

						for {
							if newFocusedWorkspace.Workspace != constants.DefaultScratchpadWorkspaceName {
								break
							}

							fmt.Println("Focused workspace is still scratchpad, waiting for it to change...")
							time.Sleep(200 * time.Microsecond) // Sleep for 5 seconds before the next iteration
							newFocusedWorkspace, err = aerospaceClient.GetFocusedWorkspace()
							if err != nil {
								logger.LogError("SHOW: unable to get focused workspace after moving window", "error", err)
								fmt.Println("Error: unable to get focused workspace after moving window", err)
							}
						}

						res, err := client.SendCommand(
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
								focusedWorkspace.Workspace,
								err,
							)
						}

						// Ensure the window moved takes the focus
						for {
							time.Sleep(200 * time.Microsecond) // Sleep for 5 seconds before the next iteration
							newFocusedWindows, err := aerospaceClient.GetFocusedWindow()
							if err != nil {
								logger.LogError("SHOW: unable to get focused window after moving", "error", err)
								fmt.Println("Error: unable to get focused window after moving", err)
								break
							}

							if newFocusedWindows.WindowID == focusedWindow.WindowID {
								logger.LogInfo(
									"DAEMON: focused window after moving",
									"window", focusedWindow,
								)
								break
							}

							err = aerospaceClient.SetFocusByWindowID(focusedWindow.WindowID)
							if err != nil {
								logger.LogError("DAEMON: unable to set focus by window ID", "error", err)
								fmt.Println("Error: unable to set focus by window ID", err)
								break
							}
						}
							
						fmt.Printf("Response: %+v\n", res)
					}
				}

				time.Sleep(200 * time.Microsecond) // Sleep for 5 seconds before the next iteration
			}
		},
	}

	return daemonCmd
}
