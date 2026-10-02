package cmd

import (
	"strconv"

	windowsipc "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

func focusPinnedWindowOnOtherMonitor(
	commandName string,
	aerospaceClient *aerospace.AeroSpaceClient,
	formatter *cli.OutputFormatter,
	window windowsipc.Window,
	currentMonitorID int,
) (bool, error) {
	pinnedMonitorID, pinned, err := aerospace.PinnedMonitorIDForWindow(window)
	if err != nil || !pinned || pinnedMonitorID == currentMonitorID {
		return false, err
	}

	focusErr := aerospaceClient.SetFocusByWindowID(window.WindowID)
	if focusErr != nil {
		return true, focusErr
	}

	if printErr := formatter.Print(cli.OutputEvent{
		Command:   commandName,
		Action:    actionFocus,
		WindowID:  window.WindowID,
		AppName:   window.AppName,
		Workspace: window.Workspace,
		Result:    "ok",
		Message:   "pinned to monitor " + strconv.Itoa(pinnedMonitorID),
	}); printErr != nil {
		logger.GetDefaultLogger().LogError("PINNED: unable to write output", "error", printErr)
	}
	return true, nil
}
