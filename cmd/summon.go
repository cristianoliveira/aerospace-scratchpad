/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
)

// SummonCmd represents the summon command.
//
//nolint:funlen,gocognit,nestif // command wiring and validation keep this function long/branchy
func SummonCmd(
	aerospaceClient *aerospace.AeroSpaceClient,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "summon <pattern>",
		Short: "Summon a matching window to the current workspace",
		Long: `Summon a matching window to the current workspace.

This command brings windows matching the regex pattern to the current workspace and focuses them.
Use "next" to cycle through scratchpad windows without specifying a pattern.
`,

		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			cli.ValidateAllNonEmpty,
		),

		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			windowNamePattern := strings.TrimSpace(args[0])

			outputFormat, err := cmd.Flags().GetString("output")
			if err != nil {
				logger.LogError("SUMMON: unable to get output flag", "error", err)
				stderr.Println("Error: unable to get output format")
				return
			}
			formatter, err := cli.NewOutputFormatter(os.Stdout, outputFormat)
			if err != nil {
				logger.LogError("SUMMON: invalid output format", "error", err)
				stderr.Println("Error: unsupported output format")
				return
			}

			focusedWorkspace, err := aerospaceClient.GetFocusedWorkspace()
			if err != nil {
				logger.LogError(
					"SUMMON: unable to get focused workspace",
					"error",
					err,
				)
				stderr.Println("Error: unable to get focused workspace")
				return
			}

			// Parse filter flags
			filterFlags, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				logger.LogError(
					"SUMMON: unable to get filter flags",
					"error",
					err,
				)
				stderr.Println("Error: unable to get filter flags")
				return
			}

			// Filter windows using the shared querier
			querier := aerospace.NewAerospaceQuerier(aerospaceClient.GetUnderlyingClient())
			mover := aerospace.NewAeroSpaceMover(aerospaceClient)

			windows, err := querier.GetFilteredWindows(
				windowNamePattern,
				filterFlags,
			)
			if err != nil {
				logger.LogError(
					"SUMMON: unable to get filtered windows",
					"error",
					err,
				)
				stderr.Println("Error: %v", err)
				return
			}

			for _, window := range windows {
				setFocus := true
				moveErr := mover.MoveWindowToWorkspace(
					&window,
					focusedWorkspace,
					setFocus,
				)
				if moveErr != nil {
					if strings.Contains(
						moveErr.Error(),
						"already belongs to workspace",
					) {
						logger.LogDebug(
							"SUMMON: window already belongs to workspace",
							"window",
							window,
							"workspace",
							focusedWorkspace,
							"error",
							moveErr,
						)
						focusErr := aerospaceClient.SetFocusByWindowID(window.WindowID)
						if focusErr != nil {
							logger.LogError(
								"SUMMON: unable to set focus to window",
								"window",
								window,
								"error",
								focusErr,
							)
							stderr.Printf(
								"Error: unable to set focus to window '%+v'\n%s",
								window,
								focusErr,
							)
							return
						}

						if printErr := formatter.Print(cli.OutputEvent{
							Command:         commandSummon,
							Action:          actionToWorkspace,
							WindowID:        window.WindowID,
							AppName:         window.AppName,
							Workspace:       window.Workspace,
							TargetWorkspace: focusedWorkspace.Workspace,
							Result:          "skipped",
							Message:         "already in target workspace",
						}); printErr != nil {
							logger.LogError("SUMMON: unable to write output", "error", printErr)
						}
						continue
					}

					logger.LogDebug(
						"SUMMON: unable to move window to workspace",
						"window",
						window,
						"workspace",
						focusedWorkspace,
						"error",
						moveErr,
					)
					stderr.Println("Error: %v", moveErr)
					return
				}

				if printErr := formatter.Print(cli.OutputEvent{
					Command:         commandSummon,
					Action:          actionToWorkspace,
					WindowID:        window.WindowID,
					AppName:         window.AppName,
					Workspace:       window.Workspace,
					TargetWorkspace: focusedWorkspace.Workspace,
					Result:          "ok",
				}); printErr != nil {
					logger.LogError("SUMMON: unable to write output", "error", printErr)
				}
			}
		},
	}
	return command
}
