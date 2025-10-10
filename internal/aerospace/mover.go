package aerospace

import (
	"errors"
	"fmt"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

type AerospaceMover interface {
	// MoveWindowToScratchpad sends a window to a workspace
	MoveWindowToScratchpad(window aerospacecli.Window) error

	// MoveWindowToWorkspace sends a window to a workspace and set focus
	MoveWindowToWorkspace(
		window aerospacecli.Window,
		workspace aerospacecli.Workspace,
		shouldSetFocus bool,
	) error
}

type MoverAeroSpace struct {
	aerospace aerospacecli.AeroSpaceClient
}

func (a *MoverAeroSpace) MoveWindowToScratchpad(window aerospacecli.Window) error {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("MOVING: MoveWindowToScratchpad", "window", window)

	err := a.aerospace.MoveWindowToWorkspace(
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

	err = a.aerospace.SetLayout(
		window.WindowID,
		"floating",
	)
	if err != nil {
		fmt.Printf(
			"Warn: unable to set layout for window '%+v' to floating\n%s",
			window,
			err,
		)
	}

	fmt.Printf("Window '%+v' hidden to scratchpad\n", window)
	return nil
}

func (a *MoverAeroSpace) MoveWindowToWorkspace(
	window *aerospacecli.Window,
	workspace *aerospacecli.Workspace,
	shouldSetFocus bool,
) error {
	if window == nil {
		return errors.New("window is nil")
	}
	if workspace == nil {
		return errors.New("workspace is nil")
	}

	if err := a.aerospace.MoveWindowToWorkspace(
		window.WindowID,
		workspace.Workspace,
	); err != nil {
		return fmt.Errorf("unable to move window '%+v' to workspace '%s': %w", window, workspace.Workspace, err)
	}

	if shouldSetFocus {
		if err := a.aerospace.SetFocusByWindowID(window.WindowID); err != nil {
			return fmt.Errorf("unable to set focus to window '%+v': %w", window, err)
		}
	}

	fmt.Printf("Window '%+v' is moved to workspace '%s'\n", window, workspace.Workspace)
	return nil
}

func NewAeroSpaceMover(aerospace aerospacecli.AeroSpaceClient) MoverAeroSpace {
	return MoverAeroSpace{
		aerospace: aerospace,
	}
}
