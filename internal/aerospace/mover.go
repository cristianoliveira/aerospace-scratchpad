package aerospace

import (
	"errors"
	"fmt"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

type Mover interface {
	// MoveWindowToScratchpad sends a window to a workspace
	MoveWindowToScratchpad(window windows.Window) (string, error)

	// MoveWindowToScratchpadForMonitor sends a window to a scratchpad workspace for a specific monitor.
	// If monitorID <= 0, falls back to the default scratchpad workspace.
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

func (a *MoverAeroSpace) MoveWindowToScratchpad(
	window windows.Window,
) (string, error) {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("MOVING: MoveWindowToScratchpad", "window", window)

	targetWorkspace := a.resolveScratchpadWorkspace()

	if err := a.moveWindowToScratchpadWorkspace(window, targetWorkspace); err != nil {
		return targetWorkspace, err
	}
	return targetWorkspace, nil
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

	targetWorkspace := a.resolveScratchpadWorkspaceForMonitor(monitorID)

	if err := a.moveWindowToScratchpadWorkspace(window, targetWorkspace); err != nil {
		return targetWorkspace, err
	}
	return targetWorkspace, nil
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

func (a *MoverAeroSpace) resolveScratchpadWorkspace() string {
	logger := logger.GetDefaultLogger()
	targetWorkspace := constants.DefaultScratchpadWorkspaceName

	monitor, err := GetFocusedMonitor(a.aerospace)
	if err != nil {
		logger.LogError(
			"MOVER: unable to get focused monitor, defaulting to base scratchpad",
			"error", err,
		)
		logger.LogDebug(
			"MOVER: returning default scratchpad workspace due to monitor detection error",
			"targetWorkspace", targetWorkspace,
		)
		return targetWorkspace
	}

	logger.LogDebug(
		"MOVER: focused monitor retrieved",
		"monitorID", monitor.MonitorID,
		"monitorName", monitor.MonitorName,
	)

	workspaceName, resolveErr := ResolveScratchpadWorkspaceNameForMonitor(
		a.aerospace,
		monitor.MonitorID,
	)
	if resolveErr != nil {
		logger.LogError(
			"MOVER: unable to resolve scratchpad workspace for monitor, defaulting to base scratchpad",
			"monitorId",
			monitor.MonitorID,
			"error",
			resolveErr,
		)
		logger.LogDebug(
			"MOVER: returning default scratchpad workspace due to resolution error",
			"targetWorkspace", targetWorkspace,
		)
		return targetWorkspace
	}

	logger.LogDebug(
		"MOVER: resolved scratchpad workspace name",
		"workspaceName", workspaceName,
	)

	if workspaceName == "" {
		logger.LogDebug(
			"MOVER: resolved workspace name is empty, returning default",
			"targetWorkspace", targetWorkspace,
		)
		return targetWorkspace
	}

	logger.LogDebug(
		"MOVER: returning resolved scratchpad workspace",
		"targetWorkspace", workspaceName,
	)
	return workspaceName
}

func (a *MoverAeroSpace) resolveScratchpadWorkspaceForMonitor(monitorID int) string {
	logger := logger.GetDefaultLogger()
	targetWorkspace := constants.DefaultScratchpadWorkspaceName

	if monitorID <= 0 {
		logger.LogDebug(
			"MOVER: monitorID <= 0, defaulting to base scratchpad",
			"monitorID", monitorID,
			"targetWorkspace", targetWorkspace,
		)
		return targetWorkspace
	}

	logger.LogDebug(
		"MOVER: resolving scratchpad workspace for monitor",
		"monitorID", monitorID,
	)

	workspaceName, resolveErr := ResolveScratchpadWorkspaceNameForMonitor(
		a.aerospace,
		monitorID,
	)
	if resolveErr != nil {
		logger.LogError(
			"MOVER: unable to resolve scratchpad workspace for monitor, defaulting to base scratchpad",
			"monitorID",
			monitorID,
			"error",
			resolveErr,
		)
		logger.LogDebug(
			"MOVER: returning default scratchpad workspace due to resolution error",
			"targetWorkspace", targetWorkspace,
		)
		return targetWorkspace
	}

	logger.LogDebug(
		"MOVER: resolved scratchpad workspace name",
		"workspaceName", workspaceName,
	)

	if workspaceName == "" {
		logger.LogDebug(
			"MOVER: resolved workspace name is empty, returning default",
			"targetWorkspace", targetWorkspace,
		)
		return targetWorkspace
	}

	logger.LogDebug(
		"MOVER: returning resolved scratchpad workspace",
		"targetWorkspace", workspaceName,
	)
	return workspaceName
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
		err = wrapper.SetLayout(window.WindowID, "floating")
	} else {
		err = a.aerospace.Layout().SetLayout([]string{"floating"}, layout.SetLayoutOpts{
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
