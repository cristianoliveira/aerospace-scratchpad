/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
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
		Long: `Summon a window from scratchpad on the current workspace.

Different from 'show' commadn, this command will not toggle the window.
`,

		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			cli.ValidateAllNonEmpty,
		),

		Run: func(cmd *cobra.Command, args []string) {
			windowNamePattern :=  strings.TrimSpace(args[0])

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
				stderr.Println("Error: invalid window-name-pattern")
				return
			}

			for _, window := range windows {
				if !windowPattern.MatchString(window.AppName) {
					continue
				}

				aerospaceClient.MoveWindowToWorkspace(
					window.WindowID,
					focusedWorkspace.Workspace,
				);

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
