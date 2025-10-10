/*
Copyright © 2025 Cristian Oliveira licence@cristianoliveira.dev
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
on-focused-monitor-changed = [
  "exec-and-forget aerospace-scratchpad wsh bring-scratchpad-to-monitor"
]
exec-on-workspace-change = ["/bin/bash", "-c",
  "aerospace-scratchpad wsh bring-window-to-workspace $AEROSPACE_PREV_WORKSPACE $AEROSPACE_FOCUSED_WORKSPACE"
]
'''
`,
		Aliases: []string{"wsh"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			client := aerospaceClient.Connection()

			logger.LogInfo("WSH: [started] workspace handler", "args", args)

			command := args[0]
			if command == "bring-scratchpad-to-monitor" {
				moveScratchpadToCurrentMonitor(aerospaceClient)
				return
			}

			if len(args) < 3 {
				logger.LogError("WSH: not enough arguments")
				// Fail and return error 1
				log.Fatal("Error: not enough arguments missing <previous-workspace> <focused-workspace>")
				return
			}

			prevWorkspace, focusedWorkspace := args[1], args[2]
			if prevWorkspace == constants.DefaultScratchpadWorkspaceName {
				logger.LogDebug(
					"WSH: previous workspace is scratchpad, nothing to do",
					"workspace", prevWorkspace,
				)
				return
			}

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

				_, err := client.SendCommand("workspace", []string{prevWorkspace})
				if err != nil {
					logger.LogError("WSH: unable to get focused window", "error", err)
					fmt.Println("Error: unable to get focused window")
					return
				}

				// if tmp file exists, return
				if _, err := os.Stat(constants.TempScratchpadMovingFile); err == nil {
					logger.LogInfo("WSH: temp file exists, returning")
					err := os.Remove(constants.TempScratchpadMovingFile)
					if err != nil {
						logger.LogError("WSH: unable to remove temp file", "error", err)
						fmt.Println("Error: unable to remove temp file", err)
					}
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

				attempts := 0
				for {
					if newFocusedWorkspace.Workspace != constants.DefaultScratchpadWorkspaceName {
						break
					}
					if attempts >= 5 {
						logger.LogError("WSH: focused workspace is still scratchpad after 5 attempts, giving up")
						break
					}

					logger.LogInfo(
						"Focused workspace is still scratchpad, retrying...",
						"attempt:", attempts,
					)
					newFocusedWorkspace, err = aerospaceClient.GetFocusedWorkspace()
					if err != nil {
						logger.LogError("WSH: unable to get focused workspace after moving window", "error", err)
						fmt.Println("Error: unable to get focused workspace after moving window", err)
					}

					time.Sleep(100 * time.Millisecond) // Sleep for 100 milliseconds to ensure the workspace change is processed
					attempts++
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

				logger.LogInfo(
					"WSH: [final] moved window to new focused workspace",
					"workspace", newFocusedWorkspace.Workspace,
					"window", focusedWindow,
				)
			}
		},
	}

	return wsHandlerCmd
}

type MoveScratchpadResult struct {
	Workspace string `json:"workspace"`
	MonitorID int    `json:"monitor-id"`
}

func moveScratchpadToCurrentMonitor(aerospaceClient aerospaceipc.AeroSpaceClient) {
	logger := logger.GetDefaultLogger()
	client := aerospaceClient.Connection()
	logger.LogDebug("WSH: moving scratchpad to current monitor")

	if _, err := os.Create(constants.TempScratchpadMovingFile); err != nil {
		logger.LogError("WSH: unable to create temp file", "error", err)
		fmt.Println("Error: unable to create temp file", err)
		return
	}

	// Check if the .scratchpad ws is already on the current monitor
	jsonAllWorkspaces, err := client.SendCommand(
		"list-workspaces",
		[]string{"--monitor",
			"focused",
			"--json",
			"--format",
			"%{workspace} %{monitor-id}",
		})

	if err != nil {
		logger.LogError("WSH: unable to list workspaces in focused monitor", "error", err)
		fmt.Println("Error: unable to list workspaces in focused monitor", err)
		return
	}

	var workspacesInWorkspace []MoveScratchpadResult
	err = json.Unmarshal([]byte(jsonAllWorkspaces.StdOut), &workspacesInWorkspace)
	if err != nil {
		logger.LogError("WSH: unable to unmarshal workspaces in focused monitor", "error", err)
		fmt.Println("Error: unable to unmarshal workspaces in focused monitor", err)
		return
	}

	_, err = client.SendCommand(
		"summon-workspace",
		[]string{
			constants.DefaultScratchpadWorkspaceName,
		},
	)
	if err != nil {
		logger.LogError("WSH: unable to move scratchpad to current monitor", "error", err)
		fmt.Println("Error: unable to move scratchpad to current monitor", err)
	}
}
