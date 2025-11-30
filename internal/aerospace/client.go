package aerospace

import (
	"fmt"
	"os"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace"
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
	return c.client.Windows().SetFocusByWindowID(windows.SetFocusArgs{
		WindowID: windowID,
	})
}

// FocusNextTilingWindow moves focus to the next tiled window in depth-first order, ignoring floating windows.
// Equivalent to: `aerospace focus dfs-next --ignore-floating`.
func (c *AeroSpaceClient) FocusNextTilingWindow() error {
	if c.dryRun {
		fmt.Fprintln(os.Stdout, "[dry-run] FocusNextTilingWindow()")
		return nil
	}
	dfsDir := "dfs-next"
	err := c.client.Windows().SetFocusByDFSWithOpts(
		windows.SetFocusByDFSArgs{
			Direction: dfsDir,
		},
		windows.SetFocusByDFSOpts{
			IgnoreFloating: true,
		},
	)
	if err != nil {
		// Try dfs-prev if dfs-next fails
		dfsDir = "dfs-prev"
		err = c.client.Windows().SetFocusByDFSWithOpts(
			windows.SetFocusByDFSArgs{
				Direction: dfsDir,
			},
			windows.SetFocusByDFSOpts{
				IgnoreFloating: true,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to focus next tiling window: %w", err)
		}
	}

	return nil
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

func (c *AeroSpaceClient) SetLayout(windowID int, layout string) error {
	if c.dryRun {
		fmt.Fprintf(
			os.Stdout,
			"[dry-run] SetLayout(windowID=%d, layout=%s)\n",
			windowID,
			layout,
		)
		return nil
	}
	return c.client.Windows().SetLayoutWithOpts(
		windows.SetLayoutArgs{
			Layouts: []string{layout},
		},
		windows.SetLayoutOpts{
			WindowID: &windowID,
		},
	)
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

// AeroSpaceWMClient defines the interface for clients that provide Windows() and Workspaces() services.
type AeroSpaceWMClient interface {
	Windows() *windows.Service
	Workspaces() *workspaces.Service
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
