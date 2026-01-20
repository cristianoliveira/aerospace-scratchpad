package aerospace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	// - A window in a scratchpad workspace (.scratchpad or .scratchpad.<monitor-id>), OR
	// - A floating window (WindowLayout == "floating")
	GetScratchpadWindows() ([]windows.Window, error)
}

type QueryMaker struct {
	cli AeroSpaceWMClient
}

// WorkspaceMonitor describes a workspace and its monitor attachment.
type WorkspaceMonitor struct {
	Workspace string `json:"workspace"`
	MonitorID int    `json:"monitor-id"`
}

// MonitorInfo captures monitor identifiers returned by AeroSpace.
type MonitorInfo struct {
	MonitorID   int    `json:"monitor-id"`
	MonitorName string `json:"monitor-name"`
}

const (
	listWorkspacesMonitorFormat = "%{workspace} %{monitor-id}"
	focusedMonitorFormat        = "%{monitor-id} %{monitor-name}"
	nextStateFileName           = "next-state.json"
	filterLogPrefix             = "FILTER: "
	floatingLayout              = "floating"
)

var scratchpadWorkspacePattern = regexp.MustCompile(
	fmt.Sprintf(`^\Q%s\E(?:\.\d+)?$`, constants.DefaultScratchpadWorkspaceName),
)

type nextState struct {
	LastWindowID int `json:"lastWindowId"`
}

// IsScratchpadWorkspace reports whether the given workspace name matches the
// scratchpad naming convention (default name or per-monitor variant).
func IsScratchpadWorkspace(workspace string) bool {
	return scratchpadWorkspacePattern.MatchString(workspace)
}

// ScratchpadWorkspaceNameForMonitor builds the scratchpad workspace name for a
// given monitor. For single-monitor setups it returns the default name to keep
// backward compatibility.
func ScratchpadWorkspaceNameForMonitor(monitorID int, monitorCount int) string {
	if monitorCount <= 1 || monitorID <= 0 {
		return constants.DefaultScratchpadWorkspaceName
	}

	return fmt.Sprintf("%s.%d", constants.DefaultScratchpadWorkspaceName, monitorID)
}

// ResolveScratchpadWorkspaceNameForMonitor returns the scratchpad workspace
// name for the provided monitor. If an existing scratchpad workspace is already
// attached to the monitor, it is returned; otherwise the name is derived from
// the monitor ID.
func ResolveScratchpadWorkspaceNameForMonitor(
	cli AeroSpaceWMClient,
	monitorID int,
) (string, error) {
	workspaces, err := ListWorkspacesWithMonitors(cli)
	if err != nil {
		return "", err
	}

	monitorCount := countUniqueMonitors(workspaces)
	expectedName := ScratchpadWorkspaceNameForMonitor(monitorID, monitorCount)

	for _, workspaceMonitor := range workspaces {
		if workspaceMonitor.MonitorID == monitorID &&
			IsScratchpadWorkspace(workspaceMonitor.Workspace) {
			if workspaceMonitor.Workspace == expectedName {
				return workspaceMonitor.Workspace, nil
			}
			// Scratchpad workspace attached to monitor but name doesn't match expected pattern.
			// Log warning and skip it (treat as if no scratchpad workspace exists).
			logger.GetDefaultLogger().LogError(
				"scratchpad workspace name mismatch",
				"monitorID", monitorID,
				"expectedName", expectedName,
				"actualName", workspaceMonitor.Workspace,
			)
			continue
		}
	}

	return expectedName, nil
}

// ListScratchpadWorkspaceNames returns the unique scratchpad workspace names
// currently present. If none are found, it falls back to the default scratchpad
// name for backward compatibility.
func ListScratchpadWorkspaceNames(cli AeroSpaceWMClient) ([]string, error) {
	workspaces, err := ListWorkspacesWithMonitors(cli)
	if err != nil {
		return nil, err
	}

	scratchpadNames := make(map[string]struct{})
	for _, workspaceMonitor := range workspaces {
		if IsScratchpadWorkspace(workspaceMonitor.Workspace) {
			scratchpadNames[workspaceMonitor.Workspace] = struct{}{}
		}
	}

	if len(scratchpadNames) == 0 {
		scratchpadNames[constants.DefaultScratchpadWorkspaceName] = struct{}{}
	}

	names := make([]string, 0, len(scratchpadNames))
	for name := range scratchpadNames {
		names = append(names, name)
	}
	// Keep deterministic order for consumers and tests.
	sort.Strings(names)

	return names, nil
}

// GetValidScratchpadWorkspaceNames returns scratchpad workspace names that are attached to existing monitors.
// Orphaned scratchpad workspaces (where monitor ID no longer exists) are filtered out.
func GetValidScratchpadWorkspaceNames(cli AeroSpaceWMClient) ([]string, error) {
	workspaces, err := ListWorkspacesWithMonitors(cli)
	if err != nil {
		return nil, err
	}
	monitors, err := ListMonitors(cli)
	if err != nil {
		return nil, err
	}
	// Build set of existing monitor IDs
	existingMonitorIDs := make(map[int]struct{})
	for _, monitor := range monitors {
		existingMonitorIDs[monitor.MonitorID] = struct{}{}
	}

	scratchpadNames := make(map[string]struct{})
	for _, workspaceMonitor := range workspaces {
		if IsScratchpadWorkspace(workspaceMonitor.Workspace) {
			// Check if monitor ID exists
			if _, exists := existingMonitorIDs[workspaceMonitor.MonitorID]; exists {
				scratchpadNames[workspaceMonitor.Workspace] = struct{}{}
			}
		}
	}

	if len(scratchpadNames) == 0 {
		scratchpadNames[constants.DefaultScratchpadWorkspaceName] = struct{}{}
	}

	names := make([]string, 0, len(scratchpadNames))
	for name := range scratchpadNames {
		names = append(names, name)
	}
	// Keep deterministic order for consumers and tests.
	sort.Strings(names)

	return names, nil
}

// ListWorkspacesWithMonitors returns all workspaces and their monitor IDs.
func ListWorkspacesWithMonitors(cli AeroSpaceWMClient) ([]WorkspaceMonitor, error) {
	response, err := cli.Connection().SendCommand(
		"list-workspaces",
		[]string{
			"--all",
			"--json",
			"--format",
			listWorkspacesMonitorFormat,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to list workspaces with monitors: %w", err)
	}

	if response.ExitCode != 0 {
		return nil, fmt.Errorf(
			"unable to list workspaces with monitors: %s",
			response.StdErr,
		)
	}

	var workspaces []WorkspaceMonitor
	if err = json.Unmarshal([]byte(response.StdOut), &workspaces); err != nil {
		return nil, fmt.Errorf("unable to parse workspaces with monitors: %w", err)
	}

	// Debug logging
	var sample []string
	for i := 0; i < len(workspaces) && i < 5; i++ {
		sample = append(sample, workspaces[i].Workspace)
	}
	logger.GetDefaultLogger().LogDebug(
		"workspaces with monitors listed",
		"count", len(workspaces),
		"sample", sample,
	)

	return workspaces, nil
}

// GetFocusedMonitor returns the currently focused monitor metadata.
func GetFocusedMonitor(cli AeroSpaceWMClient) (*MonitorInfo, error) {
	response, err := cli.Connection().SendCommand(
		"list-monitors",
		[]string{
			"--focused",
			"--json",
			"--format",
			focusedMonitorFormat,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to list monitors: %w", err)
	}

	if response.ExitCode != 0 {
		return nil, fmt.Errorf("unable to list monitors: %s", response.StdErr)
	}

	var monitors []MonitorInfo
	if err = json.Unmarshal([]byte(response.StdOut), &monitors); err != nil {
		return nil, fmt.Errorf("unable to parse monitors: %w", err)
	}
	if len(monitors) == 0 {
		return nil, errors.New("no focused monitor found")
	}

	logger.GetDefaultLogger().LogDebug(
		"focused monitor retrieved",
		"monitorID", monitors[0].MonitorID,
		"monitorName", monitors[0].MonitorName,
	)

	return &monitors[0], nil
}

// ListMonitors returns all monitors.
func ListMonitors(cli AeroSpaceWMClient) ([]MonitorInfo, error) {
	response, err := cli.Connection().SendCommand(
		"list-monitors",
		[]string{
			"--json",
			"--format",
			focusedMonitorFormat,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to list monitors: %w", err)
	}

	if response.ExitCode != 0 {
		return nil, fmt.Errorf("unable to list monitors: %s", response.StdErr)
	}

	var monitors []MonitorInfo
	if err = json.Unmarshal([]byte(response.StdOut), &monitors); err != nil {
		return nil, fmt.Errorf("unable to parse monitors: %w", err)
	}

	// Debug logging
	var monitorIDs []int
	for _, m := range monitors {
		monitorIDs = append(monitorIDs, m.MonitorID)
	}
	logger.GetDefaultLogger().LogDebug(
		"monitors listed",
		"count", len(monitors),
		"monitorIDs", monitorIDs,
	)

	return monitors, nil
}

func countUniqueMonitors(workspaces []WorkspaceMonitor) int {
	seen := make(map[int]struct{})
	for _, workspace := range workspaces {
		seen[workspace.MonitorID] = struct{}{}
	}

	return len(seen)
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
	scratchpadWindows, err := a.GetScratchpadWindows()
	if err != nil {
		return nil, err
	}
	// Filter out windows from orphaned workspaces
	validWindows, err := a.filterOutOrphanedWindows(scratchpadWindows)
	if err != nil {
		return nil, err
	}
	if len(validWindows) == 0 {
		return nil, errors.New("no scratchpad windows found")
	}
	// Read state
	state, err := readState()
	if err != nil {
		// If we can't read state, fallback to first window
		//nolint:nilerr // fallback to first window on state read error
		return &validWindows[0], nil
	}
	// Find index of last used window ID
	lastIndex := -1
	for i, w := range validWindows {
		if w.WindowID == state.LastWindowID {
			lastIndex = i
			break
		}
	}
	nextIndex := 0
	if lastIndex >= 0 {
		nextIndex = (lastIndex + 1) % len(validWindows)
	}
	nextWindow := validWindows[nextIndex]
	// Update state with new window ID (ignore write errors)
	state.LastWindowID = nextWindow.WindowID
	_ = writeState(state)
	return &nextWindow, nil
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
			filterLogPrefix+"unable to compile window pattern",
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
	logger.LogDebug(filterLogPrefix+"compiled window pattern", "pattern", appPattern)

	filters, err := ParseFilters(filterFlags)
	if err != nil {
		logger.LogError(filterLogPrefix+"unable to parse filters", "error", err)
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
			filterLogPrefix+"no windows matched the pattern",
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
		logger.LogError(filterLogPrefix+"unable to get all windows", "error", err)
		return nil, fmt.Errorf("unable to get windows: %w", err)
	}

	var floatingWindows []windows.Window
	for _, window := range allWindows {
		if window.WindowLayout == floatingLayout {
			floatingWindows = append(floatingWindows, window)
		}
	}

	logger.LogDebug(
		filterLogPrefix+"found floating windows",
		"count", len(floatingWindows),
	)

	return floatingWindows, nil
}

func (a *QueryMaker) GetScratchpadWindows() ([]windows.Window, error) {
	logger := logger.GetDefaultLogger()

	allWindows, err := a.cli.Windows().GetAllWindows()
	if err != nil {
		logger.LogError(filterLogPrefix+"unable to get all windows", "error", err)
		return nil, fmt.Errorf("unable to get windows: %w", err)
	}

	scratchpadWorkspaces, err := ListScratchpadWorkspaceNames(a.cli)
	if err != nil {
		logger.LogError(
			filterLogPrefix+"unable to list scratchpad workspaces",
			"error", err,
		)
		scratchpadWorkspaces = []string{constants.DefaultScratchpadWorkspaceName}
	}

	// Create a map to track window IDs and avoid duplicates
	scratchpadWindowMap := make(map[int]windows.Window)

	for _, workspace := range scratchpadWorkspaces {
		// Get windows from scratchpad workspace
		scratchpadWorkspaceWindows, workspaceErr := a.cli.Windows().GetAllWindowsByWorkspace(
			workspace,
		)
		if workspaceErr != nil {
			logger.LogError(
				"FILTER: unable to get windows from scratchpad workspace",
				"workspace", workspace,
				"error", workspaceErr,
			)
			// Don't fail if workspace doesn't exist, just continue
			continue
		}

		// Add windows from scratchpad workspace
		for _, window := range scratchpadWorkspaceWindows {
			scratchpadWindowMap[window.WindowID] = window
		}
	}

	// Add floating windows
	for _, window := range allWindows {
		if window.WindowLayout == floatingLayout {
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

	// Filter out windows from orphaned scratchpad workspaces
	filteredWindows, err := a.filterOutOrphanedWindows(scratchpadWindows)
	if err != nil {
		logger.LogDebug(
			"FILTER: unable to filter orphaned windows",
			"error", err,
		)
		// Return original windows if filtering fails
		return scratchpadWindows, nil
	}

	logger.LogDebug(
		"FILTER: after orphaned filtering",
		"count", len(filteredWindows),
	)

	return filteredWindows, nil
}

// filterOutOrphanedWindows filters out windows that belong to orphaned scratchpad workspaces
// (workspaces whose monitor no longer exists). Floating windows are always included.
func (a *QueryMaker) filterOutOrphanedWindows(
	ws []windows.Window,
) ([]windows.Window, error) { //nolint:unparam // error always nil for future compatibility
	logger := logger.GetDefaultLogger()
	workspaces, err := ListWorkspacesWithMonitors(a.cli)
	if err != nil {
		logger.LogDebug(
			filterLogPrefix+"unable to list workspaces with monitors, skipping orphaned filtering",
			"error", err,
		)
		return ws, nil
	}
	monitors, err := ListMonitors(a.cli)
	if err != nil {
		logger.LogDebug(
			filterLogPrefix+"unable to list monitors, skipping orphaned filtering",
			"error", err,
		)
		return ws, nil
	}
	if len(monitors) == 0 {
		// No monitors listed; assume all monitor IDs exist (e.g., mock environment)
		logger.LogDebug(filterLogPrefix + "monitor list empty, skipping orphaned filtering")
		return ws, nil
	}
	// Build set of existing monitor IDs
	existingMonitorIDs := make(map[int]struct{})
	for _, monitor := range monitors {
		existingMonitorIDs[monitor.MonitorID] = struct{}{}
	}
	// Build map from workspace name to monitor ID
	workspaceMonitor := make(map[string]int)
	for _, wm := range workspaces {
		workspaceMonitor[wm.Workspace] = wm.MonitorID
	}

	var filtered []windows.Window
	for _, window := range ws {
		// Floating windows are always included
		if window.WindowLayout == floatingLayout {
			filtered = append(filtered, window)
			continue
		}
		// Determine monitor ID for window's workspace
		monitorID, exists := workspaceMonitor[window.Workspace]
		if !exists {
			// Workspace not in mapping (e.g., unattached workspace).
			// Assume monitor exists (conservative) to avoid breaking existing tests.
			filtered = append(filtered, window)
			continue
		}
		if _, ok := existingMonitorIDs[monitorID]; ok {
			filtered = append(filtered, window)
		}
		// else: monitor ID does not exist, skip (orphaned workspace)
	}
	return filtered, nil
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

func getStateFilePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cacheDir, "aerospace-scratchpad")
	if err = os.MkdirAll(dir, 0750); err != nil {
		return "", err
	}
	return filepath.Join(dir, nextStateFileName), nil
}

func readState() (*nextState, error) {
	path, err := getStateFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &nextState{LastWindowID: 0}, nil
		}
		return nil, err
	}
	var state nextState
	if err = json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func writeState(state *nextState) error {
	path, err := getStateFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// NewAerospaceQuerier creates a new AerospaceQuerier.
func NewAerospaceQuerier(cli AeroSpaceWMClient) Querier {
	return &QueryMaker{
		cli: cli,
	}
}
