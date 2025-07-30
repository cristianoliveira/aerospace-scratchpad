/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
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

// daemonCmd represents the daemon command
func DaemonCmd(
	aerospaceClient aerospaceipc.AeroSpaceClient,
) *cobra.Command {
	var daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
		and usage of using your command. For example:

		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
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
					logger.LogInfo("Focused workspace is not scratchpad, skipping", "focusedWorkspace", focusedWorkspace)
					fmt.Println("Focused Workspace:", focusedWorkspace)

					focusedWindow, err := aerospaceClient.GetFocusedWindow()
					if err != nil {
						logger.LogError("SHOW: unable to get focused window", "error", err)
						fmt.Println("Error: unable to get focused window")
						continue
					}

					if focusedWindow.Workspace == focusedWorkspace.Workspace {
						fmt.Println("Focused window is already in scratchpad workspace, skipping")
						_, err := client.SendCommand("workspace-back-and-forth", nil)
						if err != nil {
							logger.LogError("SHOW: unable to get focused window", "error", err)
							fmt.Println("Error: unable to get focused window")
							continue
						}

						newFocusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
						for {
							if newFocusedWorkspace.Workspace != constants.DefaultScratchpadWorkspaceName {
								break
							}

							fmt.Println("Focused workspace is still scratchpad, waiting for it to change...")
							newFocusedWorkspace, err = aerospaceClient.GetFocusedWorkspace()
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
								focusedWorkspace.Workspace,
								err,
							)
							continue
						}
					}
				}

				time.Sleep(200 * time.Microsecond) // Sleep for 5 seconds before the next iteration
			}
		},
	}

	return daemonCmd
}
