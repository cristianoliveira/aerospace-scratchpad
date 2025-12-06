package aerospace

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

type Querier interface {
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
	GetNextScratchpadWindow() (*windows.Window, error)

	// GetFilteredWindows returns all windows that match the given filters
	GetFilteredWindows(
		windowNamePattern string,
		filterFlags []string,
	) ([]windows.Window, error)

	// GetAllFloatingWindows returns all floating windows
	GetAllFloatingWindows() ([]windows.Window, error)

	// GetScratchpadWindows returns all scratchpad windows
	// A scratchpad window is defined as:
	// - A window in the .scratchpad workspace, OR
	// - A floating window (WindowLayout == "floating")
	GetScratchpadWindows() ([]windows.Window, error)
}

type QueryMaker struct {
	cli AeroSpaceWMClient
}

func (a *QueryMaker) IsWindowInWorkspace(
	windowID int,
	workspaceName string,
) (bool, error) {
	// Get all windows from the workspace
	wsWindows, err := a.cli.Windows().GetAllWindowsByWorkspace(workspaceName)
	if err != nil {
		return false, fmt.Errorf(
			"unable to get windows from workspace '%s'. Reason: %w",
			workspaceName,
			err,
		)
	}

	// Check if the window is in the workspace
	for _, window := range wsWindows {
		if window.WindowID == windowID {
			return true, nil
		}
	}

	return false, nil
}

func (a *QueryMaker) IsWindowInFocusedWorkspace(
	windowID int,
) (bool, error) {
	// Get the focused workspace
	focusedWorkspace, err := a.cli.Workspaces().GetFocusedWorkspace()
	if err != nil {
		return false, fmt.Errorf(
			"unable to get focused workspace, reason %w",
			err,
		)
	}

	// Check if the window is in the focused workspace
	return a.IsWindowInWorkspace(windowID, focusedWorkspace.Workspace)
}

func (a *QueryMaker) IsWindowFocused(windowID int) (bool, error) {
	// Get the focused window
	focusedWindow, err := a.cli.Windows().GetFocusedWindow()
	if err != nil {
		return false, fmt.Errorf("unable to get focused window, reason %w", err)
	}

	// Check if the window is focused
	return focusedWindow.WindowID == windowID, nil
}

func (a *QueryMaker) GetNextScratchpadWindow() (*windows.Window, error) {
	// Get all windows from the workspace
	wsWindows, err := a.cli.Windows().GetAllWindowsByWorkspace(
		constants.DefaultScratchpadWorkspaceName,
	)
	if err != nil {
		return nil, err
	}

	if len(wsWindows) == 0 {
		return nil, errors.New("no scratchpad windows found")
	}

	return &wsWindows[0], nil
}

// Filter represents a filter with property and regex pattern.
type Filter struct {
	Property string
	Pattern  *regexp.Regexp
}

const filterPartsExpected = 2

func (a *QueryMaker) GetFilteredWindows(
	appNamePattern string,
	filterFlags []string,
) ([]windows.Window, error) {
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

	filters, err := ParseFilters(filterFlags)
	if err != nil {
		logger.LogError("FILTER: unable to parse filters", "error", err)
		return nil, err
	}

	allWindows, err := a.cli.Windows().GetAllWindows()
	if err != nil {
		logger.LogError("FILTER: unable to get all windows", "error", err)
		return nil, fmt.Errorf("unable to get windows: %w", err)
	}

	var filteredWindows []windows.Window
	for _, window := range allWindows {
		if !appPattern.MatchString(window.AppName) {
			continue
		}

		// Apply filters
		filtered, applyErr := ApplyFilters(window, filters)
		if applyErr != nil {
			return nil, fmt.Errorf(
				"error applying filters to window '%s': %w",
				window.AppName, applyErr,
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
			return nil, fmt.Errorf(
				"no windows matched the pattern '%s' with the given filters",
				appNamePattern,
			)
		}

		return nil, fmt.Errorf(
			"no windows matched the pattern '%s'",
			appNamePattern,
		)
	}

	return filteredWindows, nil
}

func (a *QueryMaker) GetAllFloatingWindows() ([]windows.Window, error) {
	logger := logger.GetDefaultLogger()

	allWindows, err := a.cli.Windows().GetAllWindows()
	if err != nil {
		logger.LogError("FILTER: unable to get all windows", "error", err)
		return nil, fmt.Errorf("unable to get windows: %w", err)
	}

	var floatingWindows []windows.Window
	for _, window := range allWindows {
		if window.WindowLayout == "floating" {
			floatingWindows = append(floatingWindows, window)
		}
	}

	logger.LogDebug(
		"FILTER: found floating windows",
		"count", len(floatingWindows),
	)

	return floatingWindows, nil
}

func (a *QueryMaker) GetScratchpadWindows() ([]windows.Window, error) {
	logger := logger.GetDefaultLogger()

	allWindows, err := a.cli.Windows().GetAllWindows()
	if err != nil {
		logger.LogError("FILTER: unable to get all windows", "error", err)
		return nil, fmt.Errorf("unable to get windows: %w", err)
	}

	// Get windows from scratchpad workspace
	scratchpadWorkspaceWindows, err := a.cli.Windows().GetAllWindowsByWorkspace(
		constants.DefaultScratchpadWorkspaceName,
	)
	if err != nil {
		logger.LogError(
			"FILTER: unable to get windows from scratchpad workspace",
			"error", err,
		)
		// Don't fail if workspace doesn't exist, just continue
		scratchpadWorkspaceWindows = []windows.Window{}
	}

	// Create a map to track window IDs and avoid duplicates
	scratchpadWindowMap := make(map[int]windows.Window)

	// Add windows from scratchpad workspace
	for _, window := range scratchpadWorkspaceWindows {
		scratchpadWindowMap[window.WindowID] = window
	}

	// Add floating windows
	for _, window := range allWindows {
		if window.WindowLayout == "floating" {
			// Only add if not already in map (avoid duplicates)
			if _, exists := scratchpadWindowMap[window.WindowID]; !exists {
				scratchpadWindowMap[window.WindowID] = window
			}
		}
	}

	// Convert map to slice
	scratchpadWindows := make([]windows.Window, 0, len(scratchpadWindowMap))
	for _, window := range scratchpadWindowMap {
		scratchpadWindows = append(scratchpadWindows, window)
	}

	logger.LogDebug(
		"FILTER: found scratchpad windows",
		"count", len(scratchpadWindows),
	)

	return scratchpadWindows, nil
}

// ParseFilters parses filter flags and returns a slice of Filter structs.
// This is exported so it can be reused by other packages.
func ParseFilters(filterFlags []string) ([]Filter, error) {
	var filters []Filter

	for _, filterFlag := range filterFlags {
		parts := strings.SplitN(filterFlag, "=", filterPartsExpected)
		if len(parts) != filterPartsExpected {
			return nil, fmt.Errorf(
				"invalid filter format: %s. Expected format: property=regex",
				filterFlag,
			)
		}

		property := strings.TrimSpace(parts[0])
		patternStr := strings.TrimSpace(parts[1])

		if property == "" || patternStr == "" {
			return nil, fmt.Errorf(
				"invalid filter format: %s. Property and pattern cannot be empty",
				filterFlag,
			)
		}

		pattern, err := regexp.Compile(patternStr)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid regex pattern '%s': %w",
				patternStr,
				err,
			)
		}

		filters = append(filters, Filter{
			Property: property,
			Pattern:  pattern,
		})
	}

	return filters, nil
}

// ApplyFilters applies all filters to a window and returns true if all filters pass.
// This is exported so it can be reused by other packages.
func ApplyFilters(window windows.Window, filters []Filter) (bool, error) {
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
		case "window-id":
			value = strconv.Itoa(window.WindowID)
		case "workspace":
			value = window.Workspace
		case "window-layout":
			value = window.WindowLayout
		default:
			return false, fmt.Errorf(
				"unknown filter property: %s",
				filter.Property,
			)
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
func NewAerospaceQuerier(cli AeroSpaceWMClient) Querier {
	return &QueryMaker{
		cli: cli,
	}
}
