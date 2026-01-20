/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	windowsipc "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
)

// parseMonitorFlag parses the --monitor flag and returns a monitor ID.
// Returns:
//
//	-1 for "all"
//	-2 for "current"
//	>=0 for specific monitor ID
func parseMonitorFlag(cmd *cobra.Command) (int, error) {
	monitorFlag, err := cmd.Flags().GetString("monitor")
	if err != nil {
		return -1, err
	}
	switch monitorFlag {
	case "all":
		return -1, nil
	case "current":
		return -2, nil
	default:
		// Try to parse as integer
		id, parseErr := strconv.Atoi(monitorFlag)
		if parseErr != nil {
			return -1, fmt.Errorf(
				"invalid monitor value: %s, expected 'current', 'all', or a monitor ID",
				monitorFlag,
			)
		}
		if id < 0 {
			return -1, errors.New("monitor ID cannot be negative")
		}
		return id, nil
	}
}

// ListCmd represents the list command.
func ListCmd(aerospaceClient *aerospace.AeroSpaceClient) *cobra.Command {
	command := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List scratchpad windows",
		Long: `List all scratchpad windows.

A scratchpad window is defined as:
- A window in a scratchpad workspace (.scratchpad or .scratchpad.<monitor-id>), OR
- A floating window (WindowLayout == "floating")

The output is scriptable and supports multiple formats (text, json, tsv, csv).
`,
		Run: func(cmd *cobra.Command, args []string) {
			runListCommand(cmd, args, aerospaceClient)
		},
	}

	return command
}

func runListCommand(cmd *cobra.Command, args []string, aerospaceClient *aerospace.AeroSpaceClient) {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("LIST: start command", "args", args)

	formatter, err := getOutputFormatter(cmd)
	if err != nil {
		return
	}

	filterFlags, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		logger.LogError("LIST: unable to get filter flags", "error", err)
		stderr.Println("Error: unable to get filter flags")
		return
	}

	monitorID, err := parseMonitorFlag(cmd)
	if err != nil {
		logger.LogError("LIST: invalid monitor flag", "error", err)
		stderr.Printf("Error: %v\n", err)
		return
	}

	querier := aerospace.NewAerospaceQuerier(aerospaceClient.GetUnderlyingClient())
	scratchpadWindows, err := querier.GetScratchpadWindowsForMonitor(monitorID)
	if err != nil {
		logger.LogError("LIST: unable to get scratchpad windows", "error", err)
		stderr.Printf("Error: %v\n", err)
		return
	}

	logger.LogDebug("LIST: retrieved scratchpad windows", "count", len(scratchpadWindows))

	filteredWindows := applyFiltersToList(scratchpadWindows, filterFlags)
	sortWindowsByAppName(filteredWindows)
	outputWindows(formatter, filteredWindows)
}

func getOutputFormatter(cmd *cobra.Command) (*cli.OutputFormatter, error) {
	logger := logger.GetDefaultLogger()
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		logger.LogError("LIST: unable to get output flag", "error", err)
		stderr.Println("Error: unable to get output format")
		return nil, err
	}

	formatter, err := cli.NewOutputFormatter(os.Stdout, outputFormat)
	if err != nil {
		logger.LogError("LIST: invalid output format", "error", err)
		stderr.Println("Error: unsupported output format")
		return nil, err
	}

	return formatter, nil
}

func applyFiltersToList(
	scratchpadWindows []windowsipc.Window,
	filterFlags []string,
) []windowsipc.Window {
	if len(filterFlags) == 0 {
		return scratchpadWindows
	}

	filters, err := aerospace.ParseFilters(filterFlags)
	if err != nil {
		stderr.Printf("Error: %v\n", err)
		return []windowsipc.Window{}
	}

	var filteredWindows []windowsipc.Window
	for _, window := range scratchpadWindows {
		matches, applyErr := aerospace.ApplyFilters(window, filters)
		if applyErr != nil {
			stderr.Printf("Error: %v\n", applyErr)
			return []windowsipc.Window{}
		}
		if matches {
			filteredWindows = append(filteredWindows, window)
		}
	}

	return filteredWindows
}

func sortWindowsByAppName(windows []windowsipc.Window) {
	sort.Slice(windows, func(i, j int) bool {
		// Sort by app name first
		if windows[i].AppName != windows[j].AppName {
			return windows[i].AppName < windows[j].AppName
		}
		// If app names are equal, sort by window ID for stable ordering
		return windows[i].WindowID < windows[j].WindowID
	})
}

func outputWindows(formatter *cli.OutputFormatter, windows []windowsipc.Window) {
	logger := logger.GetDefaultLogger()

	if len(windows) == 0 {
		if printErr := formatter.Print(cli.OutputEvent{
			Command:   "list",
			Action:    "list",
			Result:    "none",
			Message:   "no scratchpad windows found",
			Workspace: "",
		}); printErr != nil {
			logger.LogError("LIST: unable to write output", "error", printErr)
		}
		return
	}

	for _, window := range windows {
		if printErr := formatter.Print(cli.OutputEvent{
			Command:   "list",
			Action:    "list",
			WindowID:  window.WindowID,
			AppName:   window.AppName,
			Workspace: window.Workspace,
			Result:    "ok",
		}); printErr != nil {
			logger.LogError("LIST: unable to write output", "error", printErr)
		}
	}
}
