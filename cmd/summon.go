/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"
	"regexp"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/spf13/cobra"
)

func SummonCmd(
	aerospaceClient aerospacecli.AeroSpaceClient,
) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "summon <pattern>",
		Short: "Summon a window from scratchpad",
		Long: `Summon a window from the scratchpad to the current workspace.

This command brings a window from the scratchpad to the current workspace using a regex to match the window name or title.
If no pattern is provided, it summons the first window in the scratchpad.
`,

		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			cli.ValidateAllNonEmpty,
		),

		Run: func(cmd *cobra.Command, args []string) {
			windowNamePattern := strings.TrimSpace(args[0])

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

			// instantiate the regex
			windowPattern, err := regexp.Compile(windowNamePattern)
			if err != nil {
				stderr.Println("Error: invalid app-name-pattern")
				return
			}

			for _, window := range windows {
				if !windowPattern.MatchString(window.AppName) {
					continue
				}

				err := aerospaceClient.MoveWindowToWorkspace(
					window.WindowID,
					focusedWorkspace.Workspace,
				)
				if err != nil {
					stderr.Println(
						"Error: unable to move window '%+v' to workspace '%s': %v\n",
						window,
						focusedWorkspace.Workspace,
						err,
					)
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
