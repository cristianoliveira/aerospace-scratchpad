package aerospace

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

type AerospaceQuerier interface {
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

type AeroSpaceQueryMaker struct {
	cli aerospacecli.AeroSpaceClient
}

func (a *AeroSpaceQueryMaker) IsWindowInWorkspace(windowID int, workspaceName string) (bool, error) {
	// Get all windows from the workspace
	windows, err := a.cli.GetAllWindowsByWorkspace(workspaceName)
	if err != nil {
		return false, fmt.Errorf("unable to get windows from workspace '%s'. Reason: %w", workspaceName, err)
	}

	// Check if the window is in the workspace
	for _, window := range windows {
		if window.WindowID == windowID {
			return true, nil
		}
	}

	return false, nil
}

func (a *AeroSpaceQueryMaker) IsWindowInFocusedWorkspace(windowID int) (bool, error) {
	// Get the focused workspace
	focusedWorkspace, err := a.cli.GetFocusedWorkspace()
	if err != nil {
		return false, fmt.Errorf("unable to get focused workspace, reason %w", err)
	}

	// Check if the window is in the focused workspace
	return a.IsWindowInWorkspace(windowID, focusedWorkspace.Workspace)
}

func (a *AeroSpaceQueryMaker) IsWindowFocused(windowID int) (bool, error) {
	// Get the focused window
	focusedWindow, err := a.cli.GetFocusedWindow()
	if err != nil {
		return false, fmt.Errorf("unable to get focused window, reason %w", err)
	}

	// Check if the window is focused
	return focusedWindow.WindowID == windowID, nil
}

func (a *AeroSpaceQueryMaker) GetNextScratchpadWindow() (*aerospacecli.Window, error) {
	// Get all windows from the workspace
	windows, err := a.cli.GetAllWindowsByWorkspace(
		constants.DefaultScratchpadWorkspaceName,
	)
	if err != nil {
		return nil, err
	}

	if len(windows) == 0 {
		return nil, errors.New("no scratchpad windows found")
	}

	return &windows[0], nil
}

// Filter represents a filter with property and regex pattern.
type Filter struct {
	Property string
	Pattern  *regexp.Regexp
}

func (a *AeroSpaceQueryMaker) GetFilteredWindows(
	appNamePattern string,
	filterFlags []string,
) ([]aerospacecli.Window, error) {
	logger := logger.GetDefaultLogger()

	// instantiate the regex
	appPattern, err := regexp.Compile(appNamePattern)
	if err != nil {
		logger.LogError(
			"FILTER: unable to compile window pattern",
			"pattern",
			appNamePattern,
			"error",
			err,
		)
		return nil, fmt.Errorf(
			"invalid app-name-pattern, %w",
			err,
		)
	}
	logger.LogDebug("FILTER: compiled window pattern", "pattern", appPattern)

	filters, err := parseFilters(filterFlags)
	if err != nil {
		logger.LogError("FILTER: unable to parse filters", "error", err)
		return nil, err
	}

	windows, err := a.cli.GetAllWindows()
	if err != nil {
		logger.LogError("FILTER: unable to get all windows", "error", err)
		return nil, fmt.Errorf("unable to get windows: %w", err)
	}

	var filteredWindows []aerospacecli.Window
	for _, window := range windows {
		if !appPattern.MatchString(window.AppName) {
			continue
		}

		// Apply filters
		filtered, err := applyFilters(window, filters)
		if err != nil {
			return nil, fmt.Errorf(
				"error applying filters to window '%s': %w",
				window.AppName, err,
			)
		}
		if !filtered {
			continue
		}

		filteredWindows = append(filteredWindows, window)
	}

	if len(filteredWindows) == 0 {
		logger.LogDebug(
			"FILTER: no windows matched the pattern",
			"pattern", appNamePattern,
		)

		if len(filters) > 0 {
			return nil, fmt.Errorf("no windows matched the pattern '%s' with the given filters", appNamePattern)
		}

		return nil, fmt.Errorf("no windows matched the pattern '%s'", appNamePattern)
	}

	return filteredWindows, nil
}

// parseFilters parses filter flags and returns a slice of Filter structs.
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

// applyFilters applies all filters to a window and returns true if all filters pass.
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
				"FILTER: filter did not match",
				"property", filter.Property,
				"value", value,
				"pattern", filter.Pattern.String(),
			)
			return false, nil
		}
	}

	if len(filters) > 0 {
		logger.LogDebug("FILTER: filters applied", "filters", filters)
	}

	return true, nil
}

// NewAerospaceQuerier creates a new AerospaceQuerier.
func NewAerospaceQuerier(cli aerospacecli.AeroSpaceClient) AerospaceQuerier {
	return &AeroSpaceQueryMaker{
		cli: cli,
	}
}
