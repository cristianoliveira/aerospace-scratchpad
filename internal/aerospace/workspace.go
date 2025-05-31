package aerospace

import (
	"fmt"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
)

type AerospaceWorkspace interface {
	// IsWindowInWorkspace checks if a window is in a workspace
	//
	// Returns true if the window is in the workspace
	IsWindowInWorkspace(windowID int, workspaceName string) (bool, error)

	// IsWindowInFocusedWorkspace checks if a window is in the focused workspace
	//
	// Returns true if the window is in the focused workspace
	IsWindowInFocusedWorkspace(windowID int) (bool, error)

	// IsWindowFocused checks if a window is focused
	//
	// Returns true if the window is focused
	IsWindowFocused(windowID int) (bool, error)

	// GetNextScratchpadWindow returns the next scratchpad window in the workspace
	GetNextScratchpadWindow() (*aerospacecli.Window, error)
}

type AeroSpaceWM struct {
	cli aerospacecli.AeroSpaceClient
}

func (a *AeroSpaceWM) IsWindowInWorkspace(windowID int, workspaceName string) (bool, error) {
	// Get all windows from the workspace
	windows, err := a.cli.GetAllWindowsByWorkspace(workspaceName)
	if err != nil {
		return false, fmt.Errorf("unable to get windows from workspace '%s'. Reason: %v", workspaceName, err)
	}

	// Check if the window is in the workspace
	for _, window := range windows {
		if window.WindowID == windowID {
			return true, nil
		}
	}

	return false, nil
}

func (a *AeroSpaceWM) IsWindowInFocusedWorkspace(windowID int) (bool, error) {
	// Get the focused workspace
	focusedWorkspace, err := a.cli.GetFocusedWorkspace()
	if err != nil {
		return false, fmt.Errorf("unable to get focused workspace, reason %v", err)
	}

	// Check if the window is in the focused workspace
	return a.IsWindowInWorkspace(windowID, focusedWorkspace.Workspace)
}

func (a *AeroSpaceWM) IsWindowFocused(windowID int) (bool, error) {
	// Get the focused window
	focusedWindow, err := a.cli.GetFocusedWindow()
	if err != nil {
		return false, fmt.Errorf("unable to get focused window, reason %v", err)
	}

	// Check if the window is focused
	return focusedWindow.WindowID == windowID, nil
}

func (a *AeroSpaceWM) GetNextScratchpadWindow() (*aerospacecli.Window, error) {
	// Get all windows from the workspace
	windows, err := a.cli.GetAllWindowsByWorkspace(
		constants.DefaultScratchpadWorkspaceName,
	)
	if err != nil {
		return nil, err
	}

	if len(windows) == 0 {
		return nil, fmt.Errorf("no scratchpad windows found")
	}

	return &windows[0], nil
}

// NewAerospaceQuerier creates a new AerospaceQuerier
func NewAerospaceQuerier(cli aerospacecli.AeroSpaceClient) AerospaceWorkspace {
	return &AeroSpaceWM{
		cli: cli,
	}
}
