package aerospace

import (
	"fmt"
	"regexp"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
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

	// GetFilteredWindows returns all windows that match the given filters
	GetFilteredWindows(windowNamePattern string, filterFlags []string) ([]aerospacecli.Window, error)
}

type AeroSpaceWM struct {
	cli     aerospacecli.AeroSpaceClient
	pattern *regexp.Regexp
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

// Filter represents a filter with property and regex pattern
type Filter struct {
	Property string
	Pattern  *regexp.Regexp
}

func (a *AeroSpaceWM) GetFilteredWindows(windowNamePattern string, filterFlags []string) ([]aerospacecli.Window, error) {
	logger := logger.GetDefaultLogger()

	// instantiate the regex
	windowPattern, err := regexp.Compile(windowNamePattern)
	if err != nil {
		logger.LogError(
			"SHOW: unable to compile window pattern",
			"pattern",
			windowNamePattern,
			"error",
			err,
		)
		return nil, fmt.Errorf(
			"invalid app-name-pattern '%s': %v",
			windowNamePattern, err,
		)
	}
	logger.LogDebug("SHOW: compiled window pattern", "pattern", windowPattern)

	filters, err := parseFilters(filterFlags)
	if err != nil {
		logger.LogError("SHOW: unable to parse filters", "error", err)
		return nil, err
	}

	windows, err := a.cli.GetAllWindows()
	if err != nil {
		return nil, fmt.Errorf("unable to get windows: %v", err)
	}

	var filteredWindows []aerospacecli.Window
	for _, window := range windows {
		if !windowPattern.MatchString(window.AppName) {
			continue
		}

		// Apply filters
		filtered, err := applyFilters(window, filters)
		if err != nil {
			return nil, fmt.Errorf(
				"error applying filters to window '%s': %v",
				window.AppName, err,
			)
		}
		if !filtered {
			continue
		}

		filteredWindows = append(filteredWindows, window)
	}

	return filteredWindows, nil
}

// parseFilters parses filter flags and returns a slice of Filter structs
func parseFilters(filterFlags []string) ([]Filter, error) {
	var filters []Filter

	for _, filterFlag := range filterFlags {
		parts := strings.SplitN(filterFlag, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid filter format: %s. Expected format: property=regex", filterFlag)
		}

		property := strings.TrimSpace(parts[0])
		patternStr := strings.TrimSpace(parts[1])

		if property == "" || patternStr == "" {
			return nil, fmt.Errorf("invalid filter format: %s. Property and pattern cannot be empty", filterFlag)
		}

		// Handle regex patterns that start with /
		if strings.HasPrefix(patternStr, "/") && strings.HasSuffix(patternStr, "/") {
			// Extract the pattern between / and /
			patternStr = patternStr[1 : len(patternStr)-1]
		}

		pattern, err := regexp.Compile(patternStr)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern '%s': %w", patternStr, err)
		}

		filters = append(filters, Filter{
			Property: property,
			Pattern:  pattern,
		})
	}

	return filters, nil
}

// applyFilters applies all filters to a window and returns true if all filters pass
func applyFilters(window aerospacecli.Window, filters []Filter) (bool, error) {
	logger := logger.GetDefaultLogger()

	for _, filter := range filters {
		var value string

		// FIXME: find a way to do it dynamically
		switch filter.Property {
		case "app-name":
			value = window.AppName
		case "window-title":
			value = window.WindowTitle
		case "app-bundle-id":
			value = window.AppBundleID
		default:
			return false, fmt.Errorf("unknown filter property: %s", filter.Property)
		}

		if !filter.Pattern.MatchString(value) {
			logger.LogDebug(
				"SHOW: filter did not match",
				"property", filter.Property,
				"value", value,
				"pattern", filter.Pattern.String(),
			)
			return false, nil
		}
	}

	if len(filters) > 0 {
		logger.LogDebug("SHOW: filters applied", "filters", filters)
	}

	return true, nil
}

// NewAerospaceQuerier creates a new AerospaceQuerier
func NewAerospaceQuerier(cli aerospacecli.AeroSpaceClient) AerospaceWorkspace {
	return &AeroSpaceWM{
		cli: cli,
	}
}
