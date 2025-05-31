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
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
func ShowCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show <pattern>",
		Short: "Show a window from scratchpad",
		Long: `Show a window from the scratchpad in the current workspace.
By default, it will set the window to floating and focus on it.

Similar to I3/Sway WM, it will toggle show/hide the window if called multiple times.
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			windowNamePattern := args[0]
			windowNamePattern = strings.TrimSpace(windowNamePattern)
			if windowNamePattern == "" {
				stderr.Println("Error: <pattern> cannot be empty")
				return
			}

			windows, err := aerospaceClient.GetAllWindows()
			if err != nil {
				stderr.Println("Error: unable to get windows")
				return
			}

			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				stderr.Println("Error: unable to get focused workspace")
				return
			}

			querier := aerospace.NewAerospaceQuerier(aerospaceClient)

			// instantiate the regex
			windowPattern, err := regexp.Compile(windowNamePattern)
			if err != nil {
				stderr.Println("Error: invalid window-name-pattern")
				return
			}

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
					isWindowFocused, err := querier.IsWindowFocused(window.WindowID)
					if err != nil {
						stderr.Printf("Error: unable to check if window '%+v' is focused\n", window)
						return
					}
					if isWindowFocused {
						if err = aerospaceClient.MoveWindowToWorkspace(
							window.WindowID,
							constants.DefaultScratchpadWorkspaceName,
						); err != nil {
							stderr.Printf("Error: unable to move window '%+v' to scratchpad\n", window)
							return
						}

						err = aerospaceClient.SetLayout(
							window.WindowID,
							"floating",
						)
						if err != nil {
							fmt.Printf(
								"Error: unable to set layout for window '%+v' to floating\n%s",
								window,
								err,
							)
							return
						}

						fmt.Printf("Window '%+v' hidden to scratchpad\n", window)
						return
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

					fmt.Printf("Window '%+v' is showed\n", window)
					return
				}

				if err = aerospaceClient.MoveWindowToWorkspace(
					window.WindowID,
					focusedWorkspace.Workspace,
				); err != nil {
					stderr.Printf("Error: unable to move window '%+v' to workspace '%s'\n", window, focusedWorkspace.Workspace)
					return
				}

				if err = aerospaceClient.SetFocusByWindowID(window.WindowID); err != nil {
					stderr.Printf("Error: unable to set focus to window '%+v'\n", window)
					return
				}

				fmt.Printf("Window '%+v' is summoned\n", window)
			}
		},
	}

	return showCmd
}
