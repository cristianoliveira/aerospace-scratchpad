/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"log"
	_ "net/http/pprof"
	"regexp"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/spf13/cobra"
)

// moveCmd represents the move command
func MoveCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	moveCmd := &cobra.Command{
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

			// instantiate the regex
			regex, err := regexp.Compile(windowNamePattern)
			if err != nil {
				logger.LogError(
					"MOVE: error compiling regex",
					"windowNamePattern", windowNamePattern,
					"error", err,
				)
				log.Fatalf("Error compiling regex '%s': %v", windowNamePattern, err)
			}

			// Get all windows
			windows, err := aerospaceClient.GetAllWindows()
			logger.LogDebug(
				"MOVE: retrieved all windows",
				"windows", windows,
				"error", err,
			)
			if err != nil {
				stderr.Println("Error: unable to get windows")
				return
			}

			var movedCount int
			for _, window := range windows {
				if !regex.MatchString(window.AppName) {
					continue
				}

				// Move the window to the scratchpad
				fmt.Printf("Moving window %+v to scratchpad\n", window)
				err := aerospaceClient.MoveWindowToWorkspace(
					window.WindowID,
					constants.DefaultScratchpadWorkspaceName,
				)
				logger.LogDebug(
					"MOVE: moving window to scratchpad",
					"window", window,
					"workspace", constants.DefaultScratchpadWorkspaceName,
					"error", err,
				)
				if err != nil {
					if strings.Contains(err.Error(), "already belongs to workspace") {
						continue
					}

					stderr.Println("Error: unable to move window '%+v' to scratchpad", window)
					// exit loop
					return
				}

				err = aerospaceClient.SetLayout(
					window.WindowID,
					"floating",
				)
				logger.LogDebug(
					"MOVE: setting layout to floating",
					"window", window,
					"layout", "floating",
					"error", err,
				)
				if err != nil {
					fmt.Printf(
						"warn: unable to set layout for window '%+v' to floating\n%s",
						window,
						err,
					)
				}

				movedCount++
			}

			if movedCount == 0 {
				logger.LogDebug(
					"MOVE: no windows matched the pattern",
					"pattern", windowNamePattern,
				)

				stderr.Println(
					"Error: no windows matched the pattern '%s'",
					windowNamePattern,
				)
				return
			}
		},
	}

	return moveCmd
}
