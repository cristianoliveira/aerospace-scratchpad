package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	windowsipc "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
)

func PinCmd(aerospaceClient *aerospace.AeroSpaceClient) *cobra.Command {
	command := &cobra.Command{
		Use:   "pin [pattern]",
		Short: "Pin a scratchpad window to its current monitor",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			formatter, err := outputFormatterFrom(cmd)
			if err != nil {
				stderr.Println("Error: %v", err)
				return
			}

			windows, err := windowsForOptionalPattern(cmd, args, aerospaceClient)
			if err != nil {
				stderr.Println("Error: %v", err)
				return
			}

			filterFlags, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				stderr.Println("Error: %v", err)
				return
			}
			pattern := optionalPattern(args)

			for _, window := range windows {
				monitorID, resolveErr := aerospace.ResolveWindowMonitorID(
					aerospaceClient.GetUnderlyingClient(),
					window,
				)
				if resolveErr != nil {
					stderr.Println("Error: %v", resolveErr)
					return
				}
				pinErr := pinWindowOrRule(pattern, filterFlags, window.WindowID, monitorID)
				if pinErr != nil {
					stderr.Println("Error: %v", pinErr)
					return
				}
				if printErr := formatter.Print(cli.OutputEvent{
					Command:   commandPin,
					Action:    "pin",
					WindowID:  window.WindowID,
					AppName:   window.AppName,
					Workspace: window.Workspace,
					Result:    "ok",
					Message:   "pinned to monitor " + strconv.Itoa(monitorID),
				}); printErr != nil {
					logger.LogError("PIN: unable to write output", "error", printErr)
				}
			}
		},
	}
	return command
}

func UnpinCmd(aerospaceClient *aerospace.AeroSpaceClient) *cobra.Command {
	command := &cobra.Command{
		Use:   "unpin [pattern]",
		Short: "Unpin a scratchpad window",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.GetDefaultLogger()
			formatter, err := outputFormatterFrom(cmd)
			if err != nil {
				stderr.Println("Error: %v", err)
				return
			}

			windows, err := windowsForOptionalPattern(cmd, args, aerospaceClient)
			if err != nil {
				stderr.Println("Error: %v", err)
				return
			}

			filterFlags, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				stderr.Println("Error: %v", err)
				return
			}
			pattern := optionalPattern(args)
			if pattern != "" {
				unpinErr := aerospace.UnpinRuleForPattern(pattern, filterFlags)
				if unpinErr != nil {
					stderr.Println("Error: %v", unpinErr)
					return
				}
			}

			for _, window := range windows {
				unpinErr := aerospace.UnpinWindow(window.WindowID)
				if unpinErr != nil {
					stderr.Println("Error: %v", unpinErr)
					return
				}
				if printErr := formatter.Print(cli.OutputEvent{
					Command:   commandUnpin,
					Action:    "unpin",
					WindowID:  window.WindowID,
					AppName:   window.AppName,
					Workspace: window.Workspace,
					Result:    "ok",
				}); printErr != nil {
					logger.LogError("UNPIN: unable to write output", "error", printErr)
				}
			}
		},
	}
	return command
}

func outputFormatterFrom(cmd *cobra.Command) (*cli.OutputFormatter, error) {
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return nil, err
	}
	return cli.NewOutputFormatter(os.Stdout, outputFormat)
}

func pinWindowOrRule(pattern string, filters []string, windowID int, monitorID int) error {
	if pattern != "" {
		return aerospace.PinRuleForPattern(pattern, filters, monitorID)
	}
	return aerospace.PinWindow(windowID, monitorID)
}

func optionalPattern(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return strings.TrimSpace(args[0])
}

func windowsForOptionalPattern(
	cmd *cobra.Command,
	args []string,
	aerospaceClient *aerospace.AeroSpaceClient,
) ([]windowsipc.Window, error) {
	if optionalPattern(args) == "" {
		window, err := aerospaceClient.GetFocusedWindow()
		if err != nil {
			return nil, err
		}
		return []windowsipc.Window{*window}, nil
	}

	filterFlags, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return nil, err
	}
	querier := aerospace.NewAerospaceQuerier(aerospaceClient.GetUnderlyingClient())
	return querier.GetFilteredWindows(optionalPattern(args), filterFlags)
}
