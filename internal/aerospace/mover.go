package aerospace

import (
	"errors"
	"fmt"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

type Mover interface {
	// MoveWindowToScratchpad sends a window to the scratchpad attached to the
	// window's own source monitor.
	MoveWindowToScratchpad(window windows.Window) (string, error)

	// MoveWindowToScratchpadForMonitor sends a window to a scratchpad workspace
	// attached to the given source monitor. The monitor invariant is enforced
	// best-effort (check-then-move over separate IPC calls, not race-free): if a
	// same-monitor target cannot be established (unknown monitor, target
	// attached elsewhere, or a missing target in a multi-monitor setup, where
	// safe provisioning cannot be proven), it returns an error and leaves the
	// window where it is. Missing targets are only provisioned on proven
	// single-monitor setups.
	MoveWindowToScratchpadForMonitor(window windows.Window, monitorID int) (string, error)

	// MoveWindowToWorkspace sends a window to a workspace and set focus
	MoveWindowToWorkspace(
		window *windows.Window,
		workspace *workspaces.Workspace,
		shouldSetFocus bool,
	) error
}

type MoverAeroSpace struct {
	aerospace AeroSpaceWMClient
}

func NewAeroSpaceMover(aerospace AeroSpaceWMClient) MoverAeroSpace {
	return MoverAeroSpace{
		aerospace: aerospace,
	}
}

// MoveWindowToScratchpad sends the window to the scratchpad attached to the
// window's own source monitor, resolved from the workspace-to-monitor mapping.
// It fails closed when the source monitor or a safe same-monitor target cannot
// be established.
func (a *MoverAeroSpace) MoveWindowToScratchpad(
	window windows.Window,
) (string, error) {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("MOVING: MoveWindowToScratchpad", "window", window)

	sourceMonitorID, err := ResolveSourceMonitorForWorkspace(
		a.aerospace,
		window.Workspace,
	)
	if err != nil {
		logger.LogError(
			"MOVER: unable to determine source monitor",
			"window",
			window,
			"error",
			err,
		)
		return "", fmt.Errorf(
			"unable to determine source monitor for window '%+v': %w",
			window,
			err,
		)
	}

	return a.MoveWindowToScratchpadForMonitor(window, sourceMonitorID)
}

func (a *MoverAeroSpace) MoveWindowToScratchpadForMonitor(
	window windows.Window,
	monitorID int,
) (string, error) {
	logger := logger.GetDefaultLogger()
	logger.LogDebug(
		"MOVING: MoveWindowToScratchpadForMonitor",
		"window",
		window,
		"monitorID",
		monitorID,
	)

	if monitorID <= 0 {
		return "", fmt.Errorf(
			"source monitor for window '%+v' is unknown; refusing to move to scratchpad",
			window,
		)
	}

	targetWorkspace, resolveErr := ResolveScratchpadWorkspaceNameForMonitor(
		a.aerospace,
		monitorID,
	)
	if resolveErr != nil {
		logger.LogError(
			"MOVER: unable to resolve scratchpad workspace for monitor",
			"monitorID",
			monitorID,
			"error",
			resolveErr,
		)
		return "", fmt.Errorf(
			"unable to resolve scratchpad workspace for monitor %d: %w",
			monitorID,
			resolveErr,
		)
	}

	if affinityErr := a.validateScratchpadMonitorAffinity(
		targetWorkspace,
		monitorID,
	); affinityErr != nil {
		logger.LogError(
			"MOVER: scratchpad target is not attached to the window's source monitor",
			"window",
			window,
			"targetWorkspace",
			targetWorkspace,
			"monitorID",
			monitorID,
			"error",
			affinityErr,
		)
		return targetWorkspace, affinityErr
	}

	if err := a.moveWindowToScratchpadWorkspace(window, targetWorkspace); err != nil {
		return targetWorkspace, err
	}
	return targetWorkspace, nil
}

// validateScratchpadMonitorAffinity guarantees the invariant: the window must
// never cross monitors. The target must be attached to the source monitor.
// A target that does not exist yet is only allowed on proven single-monitor
// setups, where no other monitor exists to cross. This is a check-then-move
// guard over separate IPC calls: AeroSpace state can change between
// validation and move, so the invariant is enforced best-effort at
// validation time, not race-free.
func (a *MoverAeroSpace) validateScratchpadMonitorAffinity(
	targetWorkspace string,
	sourceMonitorID int,
) error {
	workspaces, err := ListWorkspacesWithMonitors(a.aerospace)
	if err != nil {
		return fmt.Errorf(
			"unable to verify scratchpad target '%s' attachment: %w",
			targetWorkspace,
			err,
		)
	}

	for _, workspaceMonitor := range workspaces {
		if workspaceMonitor.Workspace != targetWorkspace {
			continue
		}
		if workspaceMonitor.MonitorID != sourceMonitorID {
			return fmt.Errorf(
				"scratchpad workspace '%s' is attached to monitor %d, but the window source monitor is %d; refusing cross-monitor move",
				targetWorkspace,
				workspaceMonitor.MonitorID,
				sourceMonitorID,
			)
		}
		return nil
	}

	// Target does not exist yet. AeroSpace would create it on the focused
	// monitor, and focus can change between this check and the move, so a
	// focused-matches-source match at validation time is not proof. Only a
	// proven single-monitor setup (exactly one monitor in the mapping) has no
	// other monitor to cross; otherwise fail closed.
	if countUniqueMonitors(workspaces) != 1 {
		return fmt.Errorf(
			"scratchpad workspace '%s' does not exist and cannot be safely provisioned in a multi-monitor setup; refusing cross-monitor move",
			targetWorkspace,
		)
	}
	return nil
}

func (a *MoverAeroSpace) MoveWindowToWorkspace(
	window *windows.Window,
	workspace *workspaces.Workspace,
	shouldSetFocus bool,
) error {
	if window == nil {
		return errors.New("window is nil")
	}
	if workspace == nil {
		return errors.New("workspace is nil")
	}

	// Use wrapper's MoveWindowToWorkspace if available (for dry-run support)
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		if err := wrapper.MoveWindowToWorkspace(
			window.WindowID,
			workspace.Workspace,
		); err != nil {
			return fmt.Errorf(
				"unable to move window '%+v' to workspace '%s': %w",
				window,
				workspace.Workspace,
				err,
			)
		}
	} else {
		// Fallback to direct service call
		windowID := window.WindowID
		if err := a.aerospace.Workspaces().MoveWindowToWorkspaceWithOpts(
			workspaces.MoveWindowToWorkspaceArgs{
				WorkspaceName: workspace.Workspace,
			},
			workspaces.MoveWindowToWorkspaceOpts{
				WindowID: &windowID,
			},
		); err != nil {
			return fmt.Errorf(
				"unable to move window '%+v' to workspace '%s': %w",
				window,
				workspace.Workspace,
				err,
			)
		}
	}

	if !shouldSetFocus {
		return nil
	}

	// Use wrapper's SetFocusByWindowID if available (for dry-run support)
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		if err := wrapper.SetFocusByWindowID(window.WindowID); err != nil {
			return fmt.Errorf(
				"unable to set focus to window '%+v': %w",
				window,
				err,
			)
		}
	} else {
		// Fallback to direct service call
		if err := a.aerospace.Focus().SetFocusByWindowID(window.WindowID); err != nil {
			return fmt.Errorf(
				"unable to set focus to window '%+v': %w",
				window,
				err,
			)
		}
	}

	return nil
}

func (a *MoverAeroSpace) moveWindowToScratchpadWorkspace(
	window windows.Window,
	targetWorkspace string,
) error {
	logger := logger.GetDefaultLogger()
	// Use wrapper's MoveWindowToWorkspace if available (for dry-run support)
	var err error
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		err = wrapper.MoveWindowToWorkspace(
			window.WindowID,
			targetWorkspace,
		)
	} else {
		windowID := window.WindowID
		err = a.aerospace.Workspaces().MoveWindowToWorkspaceWithOpts(
			workspaces.MoveWindowToWorkspaceArgs{
				WorkspaceName: targetWorkspace,
			},
			workspaces.MoveWindowToWorkspaceOpts{
				WindowID: &windowID,
			},
		)
	}
	logger.LogDebug(
		"MOVING: after MoveWindowToWorkspace",
		"window", window,
		"to-workspace", targetWorkspace,
		"error", err,
	)
	if err != nil {
		return err
	}

	// Use wrapper's SetLayout if available (for dry-run support)
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		err = wrapper.SetLayout(window.WindowID, floatingLayout)
	} else {
		err = a.aerospace.Layout().SetLayout([]string{floatingLayout}, layout.SetLayoutOpts{
			WindowID: layout.IntPtr(window.WindowID),
		})
	}
	if err != nil {
		logger.LogDebug(
			"MOVER: unable to set layout to floating",
			"window", window,
			"error", err,
		)
	}
	return nil
}
