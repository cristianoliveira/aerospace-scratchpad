/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/spf13/cobra"
)

// moveCmd represents the move command
func MoveCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "move <pattern>",
		Short: "Move a window to scratchpad",
		Long: `Move a window to the scratchpad.

This command moves a window to the scratchpad using a regex to match the app name.
If no pattern is provided, it moves the currently focused window.
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

			if windowNamePattern == "" {
				focusedWindow, err := aerospaceClient.GetFocusedWindow()
				logger.LogDebug(
					"MOVE: retrieving focused window",
					"focusedWindow", focusedWindow,
					"error", err,
				)
				if err != nil {
					stderr.Println("Error: unable to get focused window: %v", err)
					return
				}
				if focusedWindow == nil {
					stderr.Println("Error: no focused window found")
					return
				}
				windowNamePattern = fmt.Sprintf("^%s$", focusedWindow.AppName)
				logger.LogDebug(
					"MOVE: using focused window app name as pattern",
					"windowNamePattern", windowNamePattern,
				)
			}

			// Parse filter flags (matches show command behavior)
			filterFlags, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				logger.LogError("MOVE: unable to get filter flags", "error", err)
				stderr.Println("Error: unable to get filter flags")
				return
			}

			// Query windows matching pattern and filters
			querier := aerospace.NewAerospaceQuerier(aerospaceClient)
			mover := aerospace.NewAeroSpaceMover(aerospaceClient)

			windows, err := querier.GetFilteredWindows(windowNamePattern, filterFlags)
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

			for _, window := range windows {
				// Move the window to the scratchpad
				fmt.Printf("Moving window %+v to scratchpad\n", window)

				err := mover.MoveWindowToScratchpad(window)
				if err != nil {
					if strings.Contains(err.Error(), "already belongs to workspace") {
						continue
					}

					stderr.Println("Error: %v", err)
					return
				}
			}
		},
	}

	// Filter flags --filter
	command.Flags().StringArrayP("filter", "F", []string{}, "Filter windows by a specific property (e.g., app-name, window-title). Can be used multiple times.")

	return command
}
