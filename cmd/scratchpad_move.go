/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	windowsipc "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
)

// moveWindowToSourceMonitorScratchpad moves a window to the scratchpad
// attached to the window's own source monitor, resolved from the
// workspace-to-monitor mapping. This is the monitor invariant: a window is
// never moved to a scratchpad on another monitor; when the source monitor or
// a safe target cannot be established, it fails closed.
func moveWindowToSourceMonitorScratchpad(
	aerospaceClient *aerospace.AeroSpaceClient,
	mover *aerospace.MoverAeroSpace,
	window windowsipc.Window,
) (string, error) {
	sourceMonitorID, monitorErr := aerospace.ResolveSourceMonitorForWorkspace(
		aerospaceClient.GetUnderlyingClient(),
		window.Workspace,
	)
	if monitorErr != nil {
		return "", monitorErr
	}

	return mover.MoveWindowToScratchpadForMonitor(window, sourceMonitorID)
}
