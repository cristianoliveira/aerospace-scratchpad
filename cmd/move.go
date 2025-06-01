/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	_ "net/http/pprof"
	"regexp"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var windowNamePattern string
			if len(args) == 0 {
				windowNamePattern = ""
			} else {
				windowNamePattern = strings.TrimSpace(args[0])
			}

			if windowNamePattern == "" {
				focusedWindow, err := aerospaceClient.GetFocusedWindow()
				if err != nil {
					return fmt.Errorf("unable to get focused window: %v", err)
				}
				if focusedWindow == nil {
					return fmt.Errorf("no focused window found")
				}
				windowNamePattern = fmt.Sprintf("^%s$", focusedWindow.AppName)
			}

			// instantiate the regex
			regex, err := regexp.Compile(windowNamePattern)
			if err != nil {
				fmt.Println("invalid window-name-pattern")
			}

			// Get all windows
			windows, err := aerospaceClient.GetAllWindows()
			if err != nil {
				return fmt.Errorf("unable to get windows")
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
				if err != nil {
					if strings.Contains(err.Error(), "already belongs to workspace") {
						return fmt.Errorf("window '%+v' already belongs to scratchpad", window)
					}

					return fmt.Errorf("unable to move window '%+v' to scratchpad", window)
				}

				err = aerospaceClient.SetLayout(
					window.WindowID,
					"floating",
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
				fmt.Println("No windows matched the pattern")
			}

			return nil
		},
	}

	return moveCmd
}
