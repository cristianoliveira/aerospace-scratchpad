/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
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

			querier := aerospace.NewAerospaceQuerier(aerospaceClient)
			mover := aerospace.NewAeroSpaceMover(aerospaceClient)

			windows, err := querier.GetFilteredWindows(
				windowNamePattern,
				filterFlags,
			)
			if err != nil {
				stderr.Printf("Error: %v\n", err)
				return
			}

			var windowsOutsideView []aerospacecli.Window
			var windowsInFocusedWorkspace []aerospacecli.Window
			var hasAtLeastOneWindowFocused bool
			for _, window := range windows {
				var isWindowInFocusedWorkspace bool
				if window.Workspace == "" {
					isWindowInFocusedWorkspace, err = querier.IsWindowInWorkspace(
						window.WindowID,
						focusedWorkspace.Workspace,
					)
					if err != nil {
						stderr.Printf(
							"Error: unable to check if window '%+v' is in workspace '%s'\n",
							window,
							focusedWorkspace.Workspace,
						)
						return
					}
				} else {
					isWindowInFocusedWorkspace = window.Workspace == focusedWorkspace.Workspace
				}
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
					fmt.Fprintf(os.Stdout, "Window '%+v' is focused\n", window)
				}

				return
			}

			for _, window := range windowsInFocusedWorkspace {
				logger.LogDebug(
					"SHOW: processing window in focused workspace",
					"window", window,
					"hasAtLeastOneWindowFocused", hasAtLeastOneWindowFocused,
				)
				if hasAtLeastOneWindowFocused {
					if i == 0 {
						if err = aerospaceClient.SetFocusBackAndForth(); err != nil {
							// No need to exit here, just log the error and continue
							logger.LogError(
								"SHOW: unable to switch focus back and forth",
								"error",
								err,
							)
						}
					}

					if err = mover.MoveWindowToScratchpad(window); err != nil {
						logger.LogDebug(
							"Error: unable to move window '%+v' to scratchpad\n%s",
							"window",
							window,
							"error",
							err,
						)
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
				fmt.Fprintf(os.Stdout, "Window '%+v' is focused\n", window)
			}
		},
	}
	return command
}
