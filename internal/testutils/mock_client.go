package testutils

import (
	"encoding/json"
	"strconv"

	"go.uber.org/mock/gomock"

	focus_mock "github.com/cristianoliveira/aerospace-ipc/mocks/aerospace/focus"
	layout_mock "github.com/cristianoliveira/aerospace-ipc/mocks/aerospace/layout"
	windows_mock "github.com/cristianoliveira/aerospace-ipc/mocks/aerospace/windows"
	workspaces_mock "github.com/cristianoliveira/aerospace-ipc/mocks/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/focus"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-ipc/pkg/client"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
)

// MockAeroSpaceWM wraps the aerospace-ipc MockClient to make it compatible with *AeroSpaceWM
// It creates real Service instances that use a routing connection to delegate to mocks.
type MockAeroSpaceWM struct {
	conn              client.AeroSpaceConnection
	routingConn       *routingConnection
	windowsService    *windows_mock.MockWindowsService
	workspacesService *workspaces_mock.MockWorkspacesService
	focusService      *focus_mock.MockFocusService
	layoutService     *layout_mock.MockLayoutService
	windowsSvc        *windows.Service
	workspacesSvc     *workspaces.Service
	focusSvc          *focus.Service
	layoutSvc         *layout.Service
}

// NewMockAeroSpaceWM creates a new mock AeroSpaceWM instance.
func NewMockAeroSpaceWM(ctrl *gomock.Controller) *MockAeroSpaceWM {
	// Create separate mock services
	windowsMock := windows_mock.NewMockWindowsService(ctrl)
	workspacesMock := workspaces_mock.NewMockWorkspacesService(ctrl)
	focusMock := focus_mock.NewMockFocusService(ctrl)
	layoutMock := layout_mock.NewMockLayoutService(ctrl)

	// Create routing connection that delegates to service mocks
	routingConn := &routingConnection{
		windowsMock:             windowsMock,
		workspacesMock:          workspacesMock,
		focusMock:               focusMock,
		layoutMock:              layoutMock,
		workspaceMonitors:       []aerospace.WorkspaceMonitor{},
		summonPlacementOverride: -1,
		ctrl:                    ctrl,
	}

	// Create real Service instances with the routing connection
	windowsSvc := windows.NewService(routingConn)
	workspacesSvc := workspaces.NewService(routingConn)
	focusSvc := focus.NewService(routingConn)
	layoutSvc := layout.NewService(routingConn)

	return &MockAeroSpaceWM{
		conn:              routingConn,
		routingConn:       routingConn,
		windowsService:    windowsMock,
		workspacesService: workspacesMock,
		focusService:      focusMock,
		layoutService:     layoutMock,
		windowsSvc:        windowsSvc,
		workspacesSvc:     workspacesSvc,
		focusSvc:          focusSvc,
		layoutSvc:         layoutSvc,
	}
}

// Windows returns the windows service (which routes to the mock via connection).
func (m *MockAeroSpaceWM) Windows() *windows.Service {
	return m.windowsSvc
}

// Workspaces returns the workspaces service (which routes to the mock via connection).
func (m *MockAeroSpaceWM) Workspaces() *workspaces.Service {
	return m.workspacesSvc
}

// Focus returns the focus service (which routes to the mock via connection).
func (m *MockAeroSpaceWM) Focus() *focus.Service {
	return m.focusSvc
}

// Layout returns the layout service (which routes to the mock via connection).
func (m *MockAeroSpaceWM) Layout() *layout.Service {
	return m.layoutSvc
}

// Connection returns the routing connection that delegates to mocks.
func (m *MockAeroSpaceWM) Connection() client.AeroSpaceConnection {
	return m.routingConn
}

// CloseConnection mocks closing the connection.
func (m *MockAeroSpaceWM) CloseConnection() error {
	return nil
}

// GetWindowsMock returns the underlying windows mock for setting expectations.
func (m *MockAeroSpaceWM) GetWindowsMock() *windows_mock.MockWindowsService {
	return m.windowsService
}

// GetWorkspacesMock returns the underlying workspaces mock for setting expectations.
func (m *MockAeroSpaceWM) GetWorkspacesMock() *workspaces_mock.MockWorkspacesService {
	return m.workspacesService
}

// GetFocusMock returns the underlying focus mock for setting expectations.
func (m *MockAeroSpaceWM) GetFocusMock() *focus_mock.MockFocusService {
	return m.focusService
}

// GetLayoutMock returns the underlying layout mock for setting expectations.
func (m *MockAeroSpaceWM) GetLayoutMock() *layout_mock.MockLayoutService {
	return m.layoutService
}

// SetWorkspaceMonitors configures the workspace-monitor mapping returned by
// list-workspaces calls.
func (m *MockAeroSpaceWM) SetWorkspaceMonitors(monitors []aerospace.WorkspaceMonitor) {
	m.routingConn.workspaceMonitors = monitors
}

// SetFocusedMonitor configures the focused monitor returned by list-monitors.
func (m *MockAeroSpaceWM) SetFocusedMonitor(monitor aerospace.MonitorInfo) {
	m.routingConn.focusedMonitor = &monitor
}

// GetSummonedWorkspaces returns the workspace names passed to summon-workspace.
func (m *MockAeroSpaceWM) GetSummonedWorkspaces() []string {
	return m.routingConn.summonCalls
}

// SetSummonWorkspaceError injects a failure for summon-workspace commands.
func (m *MockAeroSpaceWM) SetSummonWorkspaceError(err error) {
	m.routingConn.summonErr = err
}

// SetWorkspaceBackAndForthError injects a failure for
// workspace-back-and-forth commands.
func (m *MockAeroSpaceWM) SetWorkspaceBackAndForthError(err error) {
	m.routingConn.backAndForthErr = err
}

// SetSummonPlacementMonitor forces the monitor a summoned workspace lands on,
// overriding the focused-window derivation; -1 restores derived placement.
func (m *MockAeroSpaceWM) SetSummonPlacementMonitor(monitorID int) {
	m.routingConn.summonPlacementOverride = monitorID
}

// GetWorkspaceBackAndForthCalls returns how many workspace-back-and-forth
// commands were issued.
func (m *MockAeroSpaceWM) GetWorkspaceBackAndForthCalls() int {
	return m.routingConn.backAndForthCalls
}

// GetFocusMonitorCalls returns how many focus-monitor commands were issued.
func (m *MockAeroSpaceWM) GetFocusMonitorCalls() int {
	return m.routingConn.focusMonitorCalls
}

// GetWorkspaceSwitchCalls returns the workspace names passed to the
// workspace (switch) command.
func (m *MockAeroSpaceWM) GetWorkspaceSwitchCalls() []string {
	return m.routingConn.workspaceSwitchCalls
}

// SetWorkspaceSwitchError injects a failure for workspace switch commands.
func (m *MockAeroSpaceWM) SetWorkspaceSwitchError(err error) {
	m.routingConn.workspaceSwitchErr = err
}

const (
	minArgsForMoveCommand = 3
	windowIDFlag          = "--window-id"
)

// routingConnection is a connection that routes Service method calls to the appropriate mocks
// It intercepts SendCommand calls and routes them to the service mocks.
type routingConnection struct {
	windowsMock             *windows_mock.MockWindowsService
	workspacesMock          *workspaces_mock.MockWorkspacesService
	focusMock               *focus_mock.MockFocusService
	layoutMock              *layout_mock.MockLayoutService
	workspaceMonitors       []aerospace.WorkspaceMonitor
	focusedMonitor          *aerospace.MonitorInfo
	focusedWindowID         int
	summonCalls             []string
	summonErr               error
	summonPlacementOverride int
	backAndForthCalls       int
	backAndForthErr         error
	focusMonitorCalls       int
	workspaceSwitchCalls    []string
	workspaceSwitchErr      error
	ctrl                    *gomock.Controller
}

func (r *routingConnection) SendCommand(command string, args []string) (*client.Response, error) {
	// Route commands to the appropriate mock based on command name and args
	switch command {
	case "list-windows":
		return r.handleListWindows(args)
	case "focus":
		return r.handleFocus(args)
	case "layout":
		return r.handleLayout(args)
	case "list-workspaces":
		return r.handleListWorkspaces(args)
	case "list-monitors":
		return r.handleListMonitors(args)
	case "move-node-to-workspace":
		return r.handleMoveNodeToWorkspace(args)
	case "summon-workspace":
		return r.handleSummonWorkspace(args)
	case "workspace-back-and-forth":
		return r.handleWorkspaceBackAndForth(args)
	case "focus-monitor":
		return r.handleFocusMonitor(args)
	case "workspace":
		return r.handleWorkspaceSwitch(args)
	default:
		return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
	}
}

// handleSummonWorkspace models AeroSpace summon-workspace: the named workspace
// is created on the focused monitor (derived from the last focused window's
// workspace, falling back to the configured focused monitor) and recorded so
// subsequent list-workspaces queries observe it.
func (r *routingConnection) handleSummonWorkspace(args []string) (*client.Response, error) {
	if len(args) < 1 {
		return &client.Response{
			ExitCode: 1,
			StdErr:   "invalid summon-workspace command",
		}, nil
	}
	if r.summonErr != nil {
		return &client.Response{ExitCode: 1, StdOut: "", StdErr: r.summonErr.Error()}, r.summonErr
	}

	name := args[0]
	placedOn := 0
	if r.summonPlacementOverride >= 0 {
		placedOn = r.summonPlacementOverride
	} else if monitorID := r.monitorOfFocusedWindow(); monitorID != 0 {
		placedOn = monitorID
	} else if r.focusedMonitor != nil {
		placedOn = r.focusedMonitor.MonitorID
	}

	r.summonCalls = append(r.summonCalls, name)
	r.workspaceMonitors = append(r.workspaceMonitors, aerospace.WorkspaceMonitor{
		Workspace: name,
		MonitorID: placedOn,
	})
	return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
}

// handleWorkspaceBackAndForth records the workspace visibility toggle. The
// generated workspaces mock does not cover MoveBackAndForth, so calls are
// tracked as state and asserted via GetWorkspaceBackAndForthCalls.
func (r *routingConnection) handleWorkspaceBackAndForth(_ []string) (*client.Response, error) {
	if r.backAndForthErr != nil {
		return &client.Response{
			ExitCode: 1,
			StdErr:   r.backAndForthErr.Error(),
		}, r.backAndForthErr
	}
	r.backAndForthCalls++
	return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
}

// handleFocusMonitor models AeroSpace focus-monitor: the monitor becomes the
// focused one (affecting summon placement) and the focused window tracking is
// invalidated.
func (r *routingConnection) handleFocusMonitor(args []string) (*client.Response, error) {
	if len(args) < 1 {
		return &client.Response{ExitCode: 1, StdOut: "", StdErr: "invalid focus-monitor command"}, nil
	}
	r.focusMonitorCalls++
	monitorID, convErr := strconv.Atoi(args[0])
	if convErr != nil {
		return &client.Response{ExitCode: 1, StdOut: "", StdErr: convErr.Error()}, convErr
	}
	r.focusedMonitor = &aerospace.MonitorInfo{MonitorID: monitorID}
	r.focusedWindowID = 0
	return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
}

// handleWorkspaceSwitch models AeroSpace workspace <name>: it records the
// switch so tests can assert restoration, with injectable failure.
func (r *routingConnection) handleWorkspaceSwitch(args []string) (*client.Response, error) {
	if len(args) < 1 {
		return &client.Response{ExitCode: 1, StdOut: "", StdErr: "invalid workspace command"}, nil
	}
	if r.workspaceSwitchErr != nil {
		return &client.Response{
			ExitCode: 1,
			StdErr:   r.workspaceSwitchErr.Error(),
		}, r.workspaceSwitchErr
	}
	r.workspaceSwitchCalls = append(r.workspaceSwitchCalls, args[0])
	return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
}

// monitorOfFocusedWindow resolves the monitor of the last focused window via
// window -> workspace -> monitor mapping. Returns 0 when unknown.
func (r *routingConnection) monitorOfFocusedWindow() int {
	if r.focusedWindowID == 0 {
		return 0
	}
	wins, err := r.windowsMock.GetAllWindows()
	if err != nil {
		return 0
	}
	workspaceByName := make(map[string]aerospace.WorkspaceMonitor)
	for _, wm := range r.workspaceMonitors {
		workspaceByName[wm.Workspace] = wm
	}
	for _, w := range wins {
		if w.WindowID == r.focusedWindowID {
			if wm, ok := workspaceByName[w.Workspace]; ok {
				return wm.MonitorID
			}
		}
	}
	return 0
}

func (r *routingConnection) handleListWindows(args []string) (*client.Response, error) {
	// Check for --all flag
	for i, arg := range args {
		if arg == "--all" {
			// GetAllWindows
			wins, err := r.windowsMock.GetAllWindows()
			if err != nil {
				return &client.Response{ExitCode: 1, StdOut: "", StdErr: err.Error()}, err
			}
			jsonData, _ := json.Marshal(wins)
			return &client.Response{ExitCode: 0, StdOut: string(jsonData), StdErr: ""}, nil
		}
		if arg == "--workspace" && i+1 < len(args) {
			// GetAllWindowsByWorkspace
			workspace := args[i+1]
			wins, err := r.windowsMock.GetAllWindowsByWorkspace(workspace)
			if err != nil {
				return &client.Response{ExitCode: 1, StdOut: "", StdErr: err.Error()}, err
			}
			jsonData, _ := json.Marshal(wins)
			return &client.Response{ExitCode: 0, StdOut: string(jsonData), StdErr: ""}, nil
		}
		if arg == "--focused" {
			// GetFocusedWindow
			win, err := r.windowsMock.GetFocusedWindow()
			if err != nil {
				return &client.Response{ExitCode: 1, StdOut: "", StdErr: err.Error()}, err
			}
			jsonData, _ := json.Marshal([]windows.Window{*win})
			return &client.Response{ExitCode: 0, StdOut: string(jsonData), StdErr: ""}, nil
		}
	}
	return &client.Response{ExitCode: 0, StdOut: "[]", StdErr: ""}, nil
}

// handleFocus records the focused window so summon-workspace placement can
// derive the focused monitor, then routes to the focus mock.
func (r *routingConnection) handleFocus(args []string) (*client.Response, error) {
	// Find --window-id
	for i, arg := range args {
		if arg == windowIDFlag && i+1 < len(args) {
			windowID, _ := strconv.Atoi(args[i+1])
			err := r.focusMock.SetFocusByWindowID(windowID)
			if err != nil {
				return &client.Response{ExitCode: 1, StdOut: "", StdErr: err.Error()}, err
			}
			r.focusedWindowID = windowID
			return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
		}
	}
	return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
}

func (r *routingConnection) handleLayout(args []string) (*client.Response, error) {
	// Layout command: layout <layout-name> --window-id <id>
	if len(args) < 1 {
		return &client.Response{ExitCode: 1, StdOut: "", StdErr: "invalid layout command"}, nil
	}
	layoutName := args[0]
	var windowIDPtr *int
	for i, arg := range args {
		if arg == windowIDFlag && i+1 < len(args) {
			windowID, _ := strconv.Atoi(args[i+1])
			windowIDPtr = &windowID
			break
		}
	}
	var err error
	if windowIDPtr != nil {
		err = r.layoutMock.SetLayout([]string{layoutName}, layout.SetLayoutOpts{
			WindowID: windowIDPtr,
		})
	} else {
		err = r.layoutMock.SetLayout([]string{layoutName})
	}
	if err != nil {
		return &client.Response{ExitCode: 1, StdOut: "", StdErr: err.Error()}, err
	}
	return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
}

func (r *routingConnection) handleListWorkspaces(args []string) (*client.Response, error) {
	// Check for --focused or --all flags
	for _, arg := range args {
		if arg == "--focused" {
			// GetFocusedWorkspace
			ws, err := r.workspacesMock.GetFocusedWorkspace()
			if err != nil {
				return &client.Response{ExitCode: 1, StdOut: "", StdErr: err.Error()}, err
			}
			jsonData, _ := json.Marshal([]workspaces.Workspace{*ws})
			return &client.Response{ExitCode: 0, StdOut: string(jsonData), StdErr: ""}, nil
		}

		if arg == "--all" {
			jsonData, _ := json.Marshal(r.workspaceMonitors)
			return &client.Response{ExitCode: 0, StdOut: string(jsonData), StdErr: ""}, nil
		}
	}
	return &client.Response{ExitCode: 0, StdOut: "[]", StdErr: ""}, nil
}

func (r *routingConnection) handleListMonitors(_ []string) (*client.Response, error) {
	monitors := []aerospace.MonitorInfo{}
	if r.focusedMonitor != nil {
		monitors = append(monitors, *r.focusedMonitor)
	}

	jsonData, _ := json.Marshal(monitors)
	return &client.Response{ExitCode: 0, StdOut: string(jsonData), StdErr: ""}, nil
}

func (r *routingConnection) handleMoveNodeToWorkspace(args []string) (*client.Response, error) {
	// move-node-to-workspace <workspace> --window-id <id>
	if len(args) < minArgsForMoveCommand {
		return &client.Response{ExitCode: 1, StdOut: "", StdErr: "invalid move command"}, nil
	}
	workspace := args[0]
	for i, arg := range args {
		if arg == windowIDFlag && i+1 < len(args) {
			windowID, _ := strconv.Atoi(args[i+1])
			windowIDPtr := &windowID
			err := r.workspacesMock.MoveWindowToWorkspaceWithOpts(
				workspaces.MoveWindowToWorkspaceArgs{
					WorkspaceName: workspace,
				},
				workspaces.MoveWindowToWorkspaceOpts{
					WindowID: windowIDPtr,
				},
			)
			if err != nil {
				return &client.Response{ExitCode: 1, StdOut: "", StdErr: err.Error()}, err
			}
			return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
		}
	}
	return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
}

func (r *routingConnection) GetSocketPath() (string, error) {
	return "/tmp/test.sock", nil
}

func (r *routingConnection) CheckServerVersion() error {
	return nil
}

func (r *routingConnection) GetServerVersion() (string, error) {
	return "0.3.0", nil
}

func (r *routingConnection) CloseConnection() error {
	return nil
}
