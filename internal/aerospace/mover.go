package aerospace

import (
	"errors"
	"fmt"
	"os"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace"
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
	aerospace *aerospacecli.AeroSpaceWM
}

func NewAeroSpaceMover(aerospace *aerospacecli.AeroSpaceWM) MoverAeroSpace {
	return MoverAeroSpace{
		aerospace: aerospace,
	}
}

func (a *MoverAeroSpace) MoveWindowToScratchpad(
	window windows.Window,
) error {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("MOVING: MoveWindowToScratchpad", "window", window)

	err := a.aerospace.Workspaces().MoveWindowToWorkspace(
		window.WindowID,
		constants.DefaultScratchpadWorkspaceName,
	)
	logger.LogDebug(
		"MOVING: after MoveWindowToWorkspace",
		"window", window,
		"to-workspace", constants.DefaultScratchpadWorkspaceName,
		"error", err,
	)
	if err != nil {
		return err
	}

	err = a.aerospace.Windows().SetLayout(
		window.WindowID,
		"floating",
	)
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

	if err := a.aerospace.Workspaces().MoveWindowToWorkspace(
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

	if shouldSetFocus {
		if err := a.aerospace.Windows().SetFocusByWindowID(window.WindowID); err != nil {
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
