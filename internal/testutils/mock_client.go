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
		windowsMock:    windowsMock,
		workspacesMock: workspacesMock,
		focusMock:      focusMock,
		layoutMock:     layoutMock,
		ctrl:           ctrl,
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

const (
	minArgsForMoveCommand = 3
	windowIDFlag          = "--window-id"
)

// routingConnection is a connection that routes Service method calls to the appropriate mocks
// It intercepts SendCommand calls and routes them to the service mocks.
type routingConnection struct {
	windowsMock    *windows_mock.MockWindowsService
	workspacesMock *workspaces_mock.MockWorkspacesService
	focusMock      *focus_mock.MockFocusService
	layoutMock     *layout_mock.MockLayoutService
	ctrl           *gomock.Controller
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
	case "move-node-to-workspace":
		return r.handleMoveNodeToWorkspace(args)
	default:
		return &client.Response{ExitCode: 0, StdOut: "", StdErr: ""}, nil
	}
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

func (r *routingConnection) handleFocus(args []string) (*client.Response, error) {
	// Find --window-id
	for i, arg := range args {
		if arg == windowIDFlag && i+1 < len(args) {
			windowID, _ := strconv.Atoi(args[i+1])
			err := r.focusMock.SetFocusByWindowID(windowID)
			if err != nil {
				return &client.Response{ExitCode: 1, StdOut: "", StdErr: err.Error()}, err
			}
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
	// Check for --focused flag
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
	}
	return &client.Response{ExitCode: 0, StdOut: "[]", StdErr: ""}, nil
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
