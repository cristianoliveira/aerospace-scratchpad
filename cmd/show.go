/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	windowsipc "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
)

// ShowCmd represents the show command.
//
//nolint:funlen,gocognit
func ShowCmd(
	aerospaceClient *aerospace.AeroSpaceClient,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "show <pattern>",
		Short: "Show a window from scratchpad",
		Long: `Show a window from the scratchpad in the current workspace.
By default, it will set the window to floating and focus on it.

Similar to I3/Sway WM, it will toggle show/hide the window if called multiple times.
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			logger.LogDebug("SHOW: start command", "args", args)
			windowNamePattern := args[0]
			windowNamePattern = strings.TrimSpace(windowNamePattern)
			if windowNamePattern == "" {
				stderr.Println("Error: <pattern> cannot be empty")
				return
			}

			outputFormat, err := cmd.Flags().GetString("output")
			if err != nil {
				logger.LogError("SHOW: unable to get output flag", "error", err)
				stderr.Println("Error: unable to get output format")
				return
			}
			formatter, err := cli.NewOutputFormatter(os.Stdout, outputFormat)
			if err != nil {
				logger.LogError("SHOW: invalid output format", "error", err)
				stderr.Println("Error: unsupported output format")
				return
			}

			// Parse filter flags
			filterFlags, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				logger.LogError(
					"SHOW: unable to get filter flags",
					"error",
					err,
				)
				stderr.Println("Error: unable to get filter flags")
				return
			}

			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				logger.LogError(
					"SHOW: unable to get focused workspace",
					"error",
					err,
				)
				stderr.Println("Error: unable to get focused workspace")
				return
			}
			logger.LogDebug(
				"SHOW: retrieved focused workspace",
				"workspace",
				focusedWorkspace,
			)

			querier := aerospace.NewAerospaceQuerier(aerospaceClient.GetUnderlyingClient())
			mover := aerospace.NewAeroSpaceMover(aerospaceClient)

			windows, err := querier.GetFilteredWindows(
				windowNamePattern,
				filterFlags,
			)
			if err != nil {
				stderr.Printf("Error: %v\n", err)
				return
			}

			var windowsOutsideView []windowsipc.Window
			var windowsInFocusedWorkspace []windowsipc.Window
			var hasAtLeastOneWindowFocused bool
			for _, window := range windows {
				isWindowInFocusedWorkspace := window.Workspace == focusedWorkspace.Workspace
				if isWindowInFocusedWorkspace {
					windowsInFocusedWorkspace = append(
						windowsInFocusedWorkspace,
						window,
					)

					isWindowFocused, focusErr := querier.IsWindowFocused(
						window.WindowID,
					)
					if focusErr != nil {
						stderr.Printf(
							"Error: unable to check if window '%+v' is focused\n",
							window,
						)
						return
					}

					// Make sure that once hasAtLeastOneWindowFocused is true, it will remain true
					hasAtLeastOneWindowFocused = hasAtLeastOneWindowFocused ||
						isWindowFocused
				} else {
					windowsOutsideView = append(windowsOutsideView, window)
				}

				logger.LogDebug(
					"SHOW: loop",
					"windowsOutsideView", windowsOutsideView,
					"windowsInFocusedWorkspace", windowsInFocusedWorkspace,
					"hasAtLeastOneWindowFocused", hasAtLeastOneWindowFocused,
				)
			}

			logger.LogDebug(
				"SHOW: filtered windows",
				"windowsOutsideView", windowsOutsideView,
				"windowsInFocusedWorkspace", windowsInFocusedWorkspace,
				"hasAtLeastOneWindowFocused", hasAtLeastOneWindowFocused,
			)

			for _, window := range windowsOutsideView {
				moveErr := mover.MoveWindowToWorkspace(
					&window,
					focusedWorkspace,
					!hasAtLeastOneWindowFocused,
				)
				if moveErr != nil {
					stderr.Printf(
						"Error: unable to move window '%+v' to scratchpad\n%s",
						window,
						moveErr,
					)
					return
				}

				if printErr := formatter.Print(cli.OutputEvent{
					Command:         "show",
					Action:          "to-workspace",
					WindowID:        window.WindowID,
					AppName:         window.AppName,
					Workspace:       window.Workspace,
					TargetWorkspace: focusedWorkspace.Workspace,
					Result:          "ok",
				}); printErr != nil {
					logger.LogError("SHOW: unable to write output", "error", printErr)
				}
			}

			// NOTE: To avoid the ping pong of windows, so priority is
			// for bringing windows to the focused workspace
			if len(windowsOutsideView) > 0 {
				// Make sure to bring the remaining matched windows to the front
				for _, window := range windowsInFocusedWorkspace {
					err = aerospaceClient.SetFocusByWindowID(window.WindowID)
					if err != nil {
						stderr.Printf(
							"Error: unable to set focus to window '%+v'\n%s",
							window,
							err,
						)
						return
					}
					logger.LogDebug(
						"SHOW: set focus to window",
						"window",
						window,
					)
					if printErr := formatter.Print(cli.OutputEvent{
						Command:   "show",
						Action:    "focus",
						WindowID:  window.WindowID,
						AppName:   window.AppName,
						Workspace: window.Workspace,
						Result:    "ok",
					}); printErr != nil {
						logger.LogError("SHOW: unable to write output", "error", printErr)
					}
				}

				return
			}

			for i, window := range windowsInFocusedWorkspace {
				logger.LogDebug(
					"SHOW: processing window in focused workspace",
					"window", window,
					"hasAtLeastOneWindowFocused", hasAtLeastOneWindowFocused,
				)
				if hasAtLeastOneWindowFocused { //nolint:nestif // conditional flow mirrors show toggle behavior
					if i == 0 {
						logger.LogDebug(
							"SHOW: first window to hide, will focus next tiling window after hiding",
							"window",
							window,
						)
						if err = aerospaceClient.FocusNextTilingWindow(); err != nil {
							// No need to exit here, just log the error and continue
							logger.LogError(
								"SHOW: unable to focus next tiling window",
								"error",
								err,
							)
						}
					}

					moveErr := mover.MoveWindowToScratchpad(window)
					if moveErr != nil {
						logger.LogDebug(
							"Error: unable to move window '%+v' to scratchpad\n%s",
							"window",
							window,
							"error",
							moveErr,
						)
						if printErr := formatter.Print(cli.OutputEvent{
							Command:         "show",
							Action:          "to-scratchpad",
							WindowID:        window.WindowID,
							AppName:         window.AppName,
							Workspace:       window.Workspace,
							TargetWorkspace: constants.DefaultScratchpadWorkspaceName,
							Result:          "error",
							Message:         moveErr.Error(),
						}); printErr != nil {
							logger.LogError("SHOW: unable to write output", "error", printErr)
						}
						continue
					}

					if printErr := formatter.Print(cli.OutputEvent{
						Command:         "show",
						Action:          "to-scratchpad",
						WindowID:        window.WindowID,
						AppName:         window.AppName,
						Workspace:       window.Workspace,
						TargetWorkspace: constants.DefaultScratchpadWorkspaceName,
						Result:          "ok",
					}); printErr != nil {
						logger.LogError("SHOW: unable to write output", "error", printErr)
					}
					continue
				}

				err = aerospaceClient.SetFocusByWindowID(window.WindowID)
				if err != nil {
					stderr.Printf(
						"Error: unable to set focus to window '%+v'\n%s",
						window,
						err,
					)
					return
				}
				if printErr := formatter.Print(cli.OutputEvent{
					Command:   "show",
					Action:    "focus",
					WindowID:  window.WindowID,
					AppName:   window.AppName,
					Workspace: window.Workspace,
					Result:    "ok",
				}); printErr != nil {
					logger.LogError("SHOW: unable to write output", "error", printErr)
				}
			}
		},
	}
	return command
}
