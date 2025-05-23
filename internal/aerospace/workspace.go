package aerospace

import (
	"fmt"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
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
		return false, fmt.Errorf("Error: unable to get focused workspace: %v", err)
	}

	// Check if the window is in the focused workspace
	return a.IsWindowInWorkspace(windowID, focusedWorkspace.Workspace)
}

func (a *AeroSpaceWM) IsWindowFocused(windowID int) (bool, error) {
	// Get the focused window
	focusedWindow, err := a.cli.GetFocusedWindow()
	if err != nil {
		return false, fmt.Errorf("Error: unable to get focused window: %v", err)
	}

	// Check if the window is focused
	return focusedWindow.WindowID == windowID, nil
}

// NewAerospaceQuerier creates a new AerospaceQuerier
func NewAerospaceQuerier(cli aerospacecli.AeroSpaceClient) AerospaceWorkspace {
	return &AeroSpaceWM{
		cli: cli,
	}
}
