package aerospace

import (
	"fmt"
	"os"
	"strconv"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/focus"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-ipc/pkg/client"
)

// AeroSpaceClient implements the AeroSpaceClient interface for interacting with AeroSpaceWM.
//
//revive:disable:exported
type AeroSpaceClient struct {
	ogClient *aerospacecli.AeroSpaceWM
	client   AeroSpaceWMClient // Interface for Windows()/Workspaces() access
	dryRun   bool
}

// ClientOpts defines options for creating a new AeroSpaceClient.
type ClientOpts struct {
	DryRun bool
}

// NewAeroSpaceClient creates a new AeroSpaceClient with the default settings.
func NewAeroSpaceClient(client AeroSpaceWMClient) *AeroSpaceClient {
	// Type assert to get the underlying *AeroSpaceWM for storage
	var ogClient *aerospacecli.AeroSpaceWM
	if realClient, ok := client.(*aerospacecli.AeroSpaceWM); ok {
		ogClient = realClient
	}
	return &AeroSpaceClient{
		ogClient: ogClient,
		client:   client,
		dryRun:   false, // Default dry-run is false
	}
}

// SetOptions the dry-run flag for the AeroSpaceClient.
func (c *AeroSpaceClient) SetOptions(opts ClientOpts) {
	c.dryRun = opts.DryRun
}

// Windows returns the windows service.
func (c *AeroSpaceClient) Windows() *windows.Service {
	return c.client.Windows()
}

// Workspaces returns the workspaces service.
func (c *AeroSpaceClient) Workspaces() *workspaces.Service {
	return c.client.Workspaces()
}

// Focus returns the focus service.
func (c *AeroSpaceClient) Focus() *focus.Service {
	return c.client.Focus()
}

// Layout returns the layout service.
func (c *AeroSpaceClient) Layout() *layout.Service {
	return c.client.Layout()
}

// GetAllWindows retrieves all windows managed by AeroSpaceWM.
func (c *AeroSpaceClient) GetAllWindows() ([]windows.Window, error) {
	return c.client.Windows().GetAllWindows()
}

func (c *AeroSpaceClient) GetAllWindowsByWorkspace(
	workspaceName string,
) ([]windows.Window, error) {
	return c.client.Windows().GetAllWindowsByWorkspace(workspaceName)
}

func (c *AeroSpaceClient) GetFocusedWindow() (*windows.Window, error) {
	return c.client.Windows().GetFocusedWindow()
}

func (c *AeroSpaceClient) SetFocusByWindowID(windowID int) error {
	if c.dryRun {
		fmt.Fprintf(os.Stdout, "[dry-run] SetFocusByWindowID(%d)\n", windowID)
		return nil
	}
	return c.client.Focus().SetFocusByWindowID(windowID)
}

func (c *AeroSpaceClient) GetFocusedWorkspace() (*workspaces.Workspace, error) {
	return c.client.Workspaces().GetFocusedWorkspace()
}

func (c *AeroSpaceClient) MoveWindowToWorkspace(
	windowID int,
	workspaceName string,
) error {
	if c.dryRun {
		fmt.Fprintf(
			os.Stdout,
			"[dry-run] MoveWindowToWorkspace(windowID=%d, workspace=%s)\n",
			windowID,
			workspaceName,
		)
		return nil
	}
	return c.client.Workspaces().MoveWindowToWorkspaceWithOpts(
		workspaces.MoveWindowToWorkspaceArgs{
			WorkspaceName: workspaceName,
		},
		workspaces.MoveWindowToWorkspaceOpts{
			WindowID: &windowID,
		},
	)
}

func (c *AeroSpaceClient) SetLayout(windowID int, layoutName string) error {
	if c.dryRun {
		fmt.Fprintf(
			os.Stdout,
			"[dry-run] SetLayout(windowID=%d, layout=%s)\n",
			windowID,
			layoutName,
		)
		return nil
	}
	return c.client.Layout().SetLayout([]string{layoutName}, layout.SetLayoutOpts{
		WindowID: layout.IntPtr(windowID),
	})
}

// SummonWorkspace creates an absent workspace on the currently focused
// monitor via the summon-workspace command. It must only be called for names
// that do not exist yet: summoning an existing workspace relocates it —
// windows included — to the focused monitor.
func (c *AeroSpaceClient) SummonWorkspace(name string) error {
	if c.dryRun {
		fmt.Fprintf(os.Stdout, "[dry-run] SummonWorkspace(%s)\n", name)
		return nil
	}
	response, err := c.Connection().SendCommand("summon-workspace", []string{name})
	if err != nil {
		return fmt.Errorf("unable to summon workspace '%s': %w", name, err)
	}
	if response.ExitCode != 0 {
		return fmt.Errorf(
			"unable to summon workspace '%s': %s",
			name,
			response.StdErr,
		)
	}
	return nil
}

// FocusMonitor focuses the given monitor (focus-monitor), making it the
// target for workspace creation.
func (c *AeroSpaceClient) FocusMonitor(monitorID int) error {
	if c.dryRun {
		fmt.Fprintf(os.Stdout, "[dry-run] FocusMonitor(%d)\n", monitorID)
		return nil
	}
	response, err := c.Connection().SendCommand(
		"focus-monitor",
		[]string{strconv.Itoa(monitorID)},
	)
	if err != nil {
		return fmt.Errorf("unable to focus monitor %d: %w", monitorID, err)
	}
	if response.ExitCode != 0 {
		return fmt.Errorf(
			"unable to focus monitor %d: %s",
			monitorID,
			response.StdErr,
		)
	}
	return nil
}

// SwitchWorkspace focuses the named workspace (workspace command), used to
// restore a monitor's active workspace after provisioning.
func (c *AeroSpaceClient) SwitchWorkspace(name string) error {
	if c.dryRun {
		fmt.Fprintf(os.Stdout, "[dry-run] SwitchWorkspace(%s)\n", name)
		return nil
	}
	response, err := c.Connection().SendCommand("workspace", []string{name})
	if err != nil {
		return fmt.Errorf("unable to switch to workspace '%s': %w", name, err)
	}
	if response.ExitCode != 0 {
		return fmt.Errorf(
			"unable to switch to workspace '%s': %s",
			name,
			response.StdErr,
		)
	}
	return nil
}

// IsDryRun reports whether write commands are suppressed.
func (c *AeroSpaceClient) IsDryRun() bool {
	return c.dryRun
}

func (c *AeroSpaceClient) Connection() client.AeroSpaceConnection {
	if c.client != nil {
		return c.client.Connection()
	}
	return c.ogClient.Connection()
}

func (c *AeroSpaceClient) CloseConnection() error {
	if c.dryRun {
		fmt.Fprintln(os.Stdout, "[dry-run] CloseConnection()")
		return nil
	}
	if c.client != nil {
		if closer, ok := c.client.(interface{ CloseConnection() error }); ok {
			return closer.CloseConnection()
		}
		return nil
	}
	return c.ogClient.CloseConnection()
}

// AeroSpaceWMClient defines the interface for clients that provide Windows(), Workspaces(), Focus(), and Layout() services.
type AeroSpaceWMClient interface {
	Windows() *windows.Service
	Workspaces() *workspaces.Service
	Focus() *focus.Service
	Layout() *layout.Service
	Connection() client.AeroSpaceConnection
}

// GetUnderlyingClient returns the underlying AeroSpaceWM client.
// This is needed for components that need direct access to Windows() and Workspaces() methods.
func (c *AeroSpaceClient) GetUnderlyingClient() AeroSpaceWMClient {
	if c.client != nil {
		return c.client
	}
	return c.ogClient
}
