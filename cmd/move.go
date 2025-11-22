/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
)

// MoveCmd represents the move command.
//
//nolint:funlen,gocognit
func MoveCmd(aerospaceClient *aerospace.AeroSpaceClient) *cobra.Command {
	command := &cobra.Command{
		Use:   "move <pattern>",
		Short: "Move a window to scratchpad",
		Long: `Move a window to the scratchpad.

This command moves a window to the scratchpad using a regex to match the app name.
If no pattern is provided, it moves the currently focused window.

To move all windows that match the focused window's app name to the scratchpad, use the --all flag.
`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			logger.LogDebug("MOVE: start command", "args", args)
			var windowNamePattern string
			if len(args) == 0 {
				windowNamePattern = ""
			} else {
				windowNamePattern = strings.TrimSpace(args[0])
			}

			focusedWindowID := -1
			if windowNamePattern == "" {
				focusedWindow, err := aerospaceClient.GetFocusedWindow()
				logger.LogDebug(
					"MOVE: retrieving focused window",
					"focusedWindow", focusedWindow,
					"error", err,
				)
				if err != nil {
					stderr.Println(
						"Error: unable to get focused window: %v",
						err,
					)
					return
				}
				if focusedWindow == nil {
					stderr.Println("Error: no focused window found")
					return
				}
				focusedWindowID = focusedWindow.WindowID
				windowNamePattern = fmt.Sprintf("^%s$", focusedWindow.AppName)
				logger.LogDebug(
					"MOVE: using focused window app name as pattern",
					"windowNamePattern", windowNamePattern,
					"focusedWindowId", focusedWindowID,
				)
			}

			// Parse filter flags (matches show command behavior)
			filterFlags, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				logger.LogError(
					"MOVE: unable to get filter flags",
					"error",
					err,
				)
				stderr.Println("Error: unable to get filter flags")
				return
			}

			// Get all flag
			allFlag, err := cmd.Flags().GetBool("all")
			if err != nil {
				logger.LogError(
					"MOVE: unable to get all flag",
					"error",
					err,
				)
				stderr.Println("Error: unable to get all flag")
				return
			}

			// Query windows matching pattern and filters
			querier := aerospace.NewAerospaceQuerier(aerospaceClient)
			mover := aerospace.NewAeroSpaceMover(aerospaceClient)

			windows, err := querier.GetFilteredWindows(
				windowNamePattern,
				filterFlags,
			)
			if err != nil {
				logger.LogError(
					"MOVE: error retrieving filtered windows",
					"error", err,
					"pattern", windowNamePattern,
					"filterFlags", filterFlags,
				)
				stderr.Println("Error: %v", err)
				return
			}

			logger.LogDebug(
				"MOVE: retrieved filtered windows",
				"windows", windows,
				"filterFlags", filterFlags,
			)

			logger.LogDebug(
				"SHOW: first window to hide, will focus next tiling window after hiding",
			)
			if err = aerospaceClient.FocusNextTilingWindow(); err != nil {
				// No need to exit here, just log the error and continue
				logger.LogError(
					"SHOW: unable to focus next tiling window",
					"error",
					err,
				)
			}

			for _, window := range windows {
				// Skip non-focused windows unless the --all flag is provided
				if focusedWindowID != -1 && window.WindowID != focusedWindowID && !allFlag {
					logger.LogDebug(
						"MOVE: skipping window, not focused and --all flag not provided",
						"window", window,
						"focusedWindowId", focusedWindowID,
					)
					continue
				}

				// Move the window to the scratchpad
				fmt.Fprintf(
					os.Stdout,
					"Moving window %+v to scratchpad\n",
					window,
				)

				moveErr := mover.MoveWindowToScratchpad(window)
				if moveErr != nil {
					if strings.Contains(
						moveErr.Error(),
						"already belongs to workspace",
					) {
						continue
					}

					stderr.Println("Error: %v", moveErr)
					return
				}
			}
		},
	}

	// Add the all flag
	command.Flags().
		BoolP("all", "a", false, "Move all windows that match the focused window's app name to the scratchpad")

	return command
}
