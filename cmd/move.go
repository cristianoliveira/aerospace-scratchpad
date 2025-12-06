/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	windowsipc "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
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
To move all floating windows (scratchpad windows) to the scratchpad, use the --all-floating flag.
`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			logger.LogDebug("MOVE: start command", "args", args)

			// Get all-floating flag first to determine behavior
			allFloatingFlag, err := cmd.Flags().GetBool("all-floating")
			if err != nil {
				logger.LogError(
					"MOVE: unable to get all-floating flag",
					"error",
					err,
				)
				stderr.Println("Error: unable to get all-floating flag")
				return
			}

			var windowNamePattern string
			focusedWindowID := -1

			// Skip pattern logic when --all-floating is used
			if !allFloatingFlag {
				windowNamePattern, focusedWindowID, err = getWindowPattern(
					args,
					aerospaceClient,
					logger,
				)
				if err != nil {
					return
				}
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
			querier := aerospace.NewAerospaceQuerier(aerospaceClient.GetUnderlyingClient())
			mover := aerospace.NewAeroSpaceMover(aerospaceClient)

			var windows []windowsipc.Window
			if allFloatingFlag {
				// Get all floating windows when --all-floating is set
				logger.LogDebug("MOVE: using --all-floating flag, getting all floating windows")
				windows, err = querier.GetAllFloatingWindows()
				if err != nil {
					logger.LogError(
						"MOVE: error retrieving floating windows",
						"error", err,
					)
					stderr.Println("Error: %v", err)
					return
				}
			} else {
				// Normal pattern-based filtering
				windows, err = querier.GetFilteredWindows(
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

			// When using --all-floating, skip the focused window check
			if allFloatingFlag && len(windows) == 0 {
				fmt.Fprintln(os.Stdout, "No floating windows found")
				return
			}

			for _, window := range windows {
				// Skip non-focused windows unless the --all or --all-floating flag is provided
				if !allFloatingFlag && focusedWindowID != -1 &&
					window.WindowID != focusedWindowID &&
					!allFlag {
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

					logger.LogError(
						"MOVE: error moving window to scratchpad",
						"window", window,
						"error", moveErr,
					)
					stderr.Println("Error: %v", moveErr)
					// Continue with remaining windows instead of returning
					continue
				}
			}
		},
	}

	// Add the all flag
	command.Flags().
		BoolP("all", "a", false, "Move all windows that match the focused window's app name to the scratchpad")

	// Add the all-floating flag
	command.Flags().
		Bool("all-floating", false, "Move all floating windows (scratchpad windows) to the scratchpad")

	return command
}

// getWindowPattern determines the window pattern and focused window ID from args.
// Returns pattern, focusedWindowID, and error.
func getWindowPattern(
	args []string,
	aerospaceClient *aerospace.AeroSpaceClient,
	log logger.Logger,
) (string, int, error) {
	var windowNamePattern string
	focusedWindowID := -1

	if len(args) == 0 {
		windowNamePattern = ""
	} else {
		windowNamePattern = strings.TrimSpace(args[0])
	}

	if windowNamePattern == "" {
		focusedWindow, err := aerospaceClient.GetFocusedWindow()
		log.LogDebug(
			"MOVE: retrieving focused window",
			"focusedWindow", focusedWindow,
			"error", err,
		)
		if err != nil {
			stderr.Println(
				"Error: unable to get focused window: %v",
				err,
			)
			return "", -1, err
		}
		if focusedWindow == nil {
			stderr.Println("Error: no focused window found")
			return "", -1, errors.New("no focused window found")
		}
		focusedWindowID = focusedWindow.WindowID
		windowNamePattern = fmt.Sprintf("^%s$", focusedWindow.AppName)
		log.LogDebug(
			"MOVE: using focused window app name as pattern",
			"windowNamePattern", windowNamePattern,
			"focusedWindowId", focusedWindowID,
		)
	}

	return windowNamePattern, focusedWindowID, nil
}
