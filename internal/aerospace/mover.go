package aerospace

import (
	"errors"
	"fmt"
	"os"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

type Mover interface {
	// MoveWindowToScratchpad sends a window to a workspace
	MoveWindowToScratchpad(window windows.Window) error

	// MoveWindowToWorkspace sends a window to a workspace and set focus
	MoveWindowToWorkspace(
		window windows.Window,
		workspace workspaces.Workspace,
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
) error {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("MOVING: MoveWindowToScratchpad", "window", window)

	// Use wrapper's MoveWindowToWorkspace if available (for dry-run support)
	var err error
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		err = wrapper.MoveWindowToWorkspace(
			window.WindowID,
			constants.DefaultScratchpadWorkspaceName,
		)
	} else {
		windowID := window.WindowID
		err = a.aerospace.Workspaces().MoveWindowToWorkspaceWithOpts(
			workspaces.MoveWindowToWorkspaceArgs{
				WorkspaceName: constants.DefaultScratchpadWorkspaceName,
			},
			workspaces.MoveWindowToWorkspaceOpts{
				WindowID: &windowID,
			},
		)
	}
	logger.LogDebug(
		"MOVING: after MoveWindowToWorkspace",
		"window", window,
		"to-workspace", constants.DefaultScratchpadWorkspaceName,
		"error", err,
	)
	if err != nil {
		return err
	}

	// Use wrapper's SetLayout if available (for dry-run support)
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		err = wrapper.SetLayout(window.WindowID, "floating")
	} else {
		windowID := window.WindowID
		err = a.aerospace.Windows().SetLayoutWithOpts(
			windows.SetLayoutArgs{
				Layouts: []string{"floating"},
			},
			windows.SetLayoutOpts{
				WindowID: &windowID,
			},
		)
	}
	if err != nil {
		fmt.Fprintf(
			os.Stdout,
			"Warn: unable to set layout for window '%+v' to floating\n%s",
			window,
			err,
		)
	}

	fmt.Fprintf(os.Stdout, "Window '%+v' hidden to scratchpad\n", window)
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
		fmt.Fprintf(
			os.Stdout,
			"Window '%+v' is moved to workspace '%s'\n",
			window,
			workspace.Workspace,
		)
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
		if err := a.aerospace.Windows().SetFocusByWindowID(windows.SetFocusArgs{
			WindowID: window.WindowID,
		}); err != nil {
			return fmt.Errorf(
				"unable to set focus to window '%+v': %w",
				window,
				err,
			)
		}
	}

	fmt.Fprintf(
		os.Stdout,
		"Window '%+v' is moved to workspace '%s'\n",
		window,
		workspace.Workspace,
	)
	return nil
}
