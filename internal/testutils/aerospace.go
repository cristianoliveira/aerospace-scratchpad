package testutils

import (
	"regexp"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
)

type AeroSpaceTree struct {
	Windows         []windows.Window
	Workspace       *workspaces.Workspace
	FocusedWindowID int
}

func ExtractAllWindows(tree []AeroSpaceTree) []windows.Window {
	var allWindows []windows.Window
	for _, t := range tree {
		for _, window := range t.Windows {
			// Newer AeroSpace versions always provide the workspace; older tests may omit it
			if window.Workspace == "" && t.Workspace != nil {
				window.Workspace = t.Workspace.Workspace
			}
			allWindows = append(allWindows, window)
		}
	}
	return allWindows
}

func ExtractWindowsByName(
	tree []AeroSpaceTree,
	name string,
) []windows.Window {
	pattern := regexp.MustCompile(name)
	var matchedWindows []windows.Window
	for _, t := range tree {
		for _, window := range t.Windows {
			if pattern.MatchString(window.AppName) {
				matchedWindows = append(matchedWindows, window)
			}
		}
	}

	return matchedWindows
}

func ExtractFocusedTree(tree []AeroSpaceTree) *AeroSpaceTree {
	for _, t := range tree {
		if t.FocusedWindowID != 0 {
			return &t
		}
	}
	return nil
}

func ExtractFocusedWindow(tree []AeroSpaceTree) *windows.Window {
	for _, t := range tree {
		if t.FocusedWindowID != 0 {
			for _, window := range t.Windows {
				if window.WindowID == t.FocusedWindowID {
					return &window
				}
			}
		}
	}
	return nil
}

func ExtractScratchpadWindows(tree []AeroSpaceTree) *AeroSpaceTree {
	for _, t := range tree {
		if t.Workspace != nil &&
			t.Workspace.Workspace == constants.DefaultScratchpadWorkspaceName {
			return &t
		}
	}
	return nil
}

func ExtractWindowByID(tree []AeroSpaceTree, id int) *windows.Window {
	for _, t := range tree {
		for _, window := range t.Windows {
			if window.WindowID == id {
				return &window
			}
		}
	}
	return nil
}
