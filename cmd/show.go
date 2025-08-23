/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"regexp"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/spf13/cobra"
)

// Filter represents a filter with property and regex pattern
type Filter struct {
	Property string
	Pattern  *regexp.Regexp
}

// showCmd represents the show command
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
				logger.LogError("SHOW: unable to get filter flags", "error", err)
				stderr.Println("Error: unable to get filter flags")
				return
			}

			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				logger.LogError("SHOW: unable to get focused workspace", "error", err)
				stderr.Println("Error: unable to get focused workspace")
				return
			}
			logger.LogDebug("SHOW: retrieved focused workspace", "workspace", focusedWorkspace)

			querier := aerospace.NewAerospaceQuerier(aerospaceClient)

			windows, err := querier.GetFilteredWindows(
				windowNamePattern,
				filterFlags,
			)

			// instantiate the regex
			windowPattern, err := regexp.Compile(windowNamePattern)
			if err != nil {
				logger.LogError(
					"SHOW: unable to compile window pattern",
					"pattern",
					windowNamePattern,
					"error",
					err,
				)
				stderr.Println("Error: invalid window-name-pattern")
				return
			}
			logger.LogDebug("SHOW: compiled window pattern", "pattern", windowPattern)

			var windowsOutsideView []aerospacecli.Window
			var windowsInFocusedWorkspace []aerospacecli.Window
			var hasAtLeastOneWindowFocused bool
			for _, window := range windows {
				if !windowPattern.MatchString(window.AppName) {
					continue
				}

				var isWindowInFocusedWorkspace bool
				if window.Workspace == "" {
					isWindowInFocusedWorkspace, err = querier.IsWindowInWorkspace(
						window.WindowID,
						focusedWorkspace.Workspace,
					)
					if err != nil {
						stderr.Printf("Error: unable to check if window '%+v' is in workspace '%s'\n", window, focusedWorkspace.Workspace)
						return
					}

				} else {
					isWindowInFocusedWorkspace = window.Workspace == focusedWorkspace.Workspace

				}
				if isWindowInFocusedWorkspace {
					windowsInFocusedWorkspace = append(windowsInFocusedWorkspace, window)

					isWindowFocused, err := querier.IsWindowFocused(window.WindowID)
					if err != nil {
						stderr.Printf("Error: unable to check if window '%+v' is focused\n", window)
						return
					}

					// Make sure that once hasAtLeastOneWindowFocused is true, it will remain true
					hasAtLeastOneWindowFocused = hasAtLeastOneWindowFocused || isWindowFocused
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
				err := sendToFocusedWorkspace(
					aerospaceClient,
					window,
					focusedWorkspace,
					!hasAtLeastOneWindowFocused,
				)
				if err != nil {
					stderr.Printf(
						"Error: unable to move window '%+v' to scratchpad\n%s",
						window,
						err,
					)
					return
				}
			}

			if len(windowsInFocusedWorkspace) == 0 && len(windowsOutsideView) == 0 {
				if len(filterFlags) > 0 {
					stderr.Println("Error: no windows matched the pattern and filters")
				} else {
					stderr.Println("Error: no windows matched the pattern '%s'", windowNamePattern)
				}
				return
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
					logger.LogDebug("SHOW: set focus to window", "window", window)
					fmt.Printf("Window '%+v' is focused\n", window)
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
					if err = sendToScratchpad(aerospaceClient, window); err != nil {
						logger.LogDebug(
							"Error: unable to move window '%+v' to scratchpad\n%s",
							"window", window,
							"error", err,
						)
						continue
					}
				} else {
					err = aerospaceClient.SetFocusByWindowID(window.WindowID)
					if err != nil {
						stderr.Printf(
							"Error: unable to set focus to window '%+v'\n%s",
							window,
							err,
						)
						return
					}
					fmt.Printf("Window '%+v' is focused\n", window)
				}
			}
		},
	}

	// Filter flags --filter
	command.Flags().StringArrayP("filter", "F", []string{}, "Filter windows by a specific property (e.g., app-name, window-title). Can be used multiple times.")

	return command
}

func sendToScratchpad(
	aerospaceClient aerospacecli.AeroSpaceClient,
	window aerospacecli.Window,
) error {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("SHOW: sendToScratchpad ", "window", window)

	err := aerospaceClient.MoveWindowToWorkspace(
		window.WindowID,
		constants.DefaultScratchpadWorkspaceName,
	)
	logger.LogDebug(
		"SHOW: after aerospaceClient.MoveWindowToWorkspace",
		"window", window,
		"to-workspace", constants.DefaultScratchpadWorkspaceName,
		"error", err,
	)
	if err != nil {
		return err
	}

	err = aerospaceClient.SetLayout(
		window.WindowID,
		"floating",
	)
	if err != nil {
		fmt.Printf(
			"Warn: unable to set layout for window '%+v' to floating\n%s",
			window,
			err,
		)
	}

	fmt.Printf("Window '%+v' hidden to scratchpad\n", window)
	return nil
}

func sendToFocusedWorkspace(
	aerospaceClient aerospacecli.AeroSpaceClient,
	window aerospacecli.Window,
	focusedWorkspace *aerospacecli.Workspace,
	shouldSetFocus bool,
) error {
	if focusedWorkspace == nil {
		return fmt.Errorf("focused workspace is nil")
	}

	if err := aerospaceClient.MoveWindowToWorkspace(
		window.WindowID,
		focusedWorkspace.Workspace,
	); err != nil {
		return fmt.Errorf("unable to move window '%+v' to workspace '%s': %w", window, focusedWorkspace.Workspace, err)
	}

	if shouldSetFocus {
		if err := aerospaceClient.SetFocusByWindowID(window.WindowID); err != nil {
			return fmt.Errorf("unable to set focus to window '%+v': %w", window, err)
		}
	}

	fmt.Printf("Window '%+v' is summoned\n", window)
	return nil
}
