package aerospace

import (
	"fmt"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	socketcli "github.com/cristianoliveira/aerospace-ipc/pkg/client"
)

// AeroSpaceClient implements the AeroSpaceClient interface for interacting with AeroSpaceWM.
type AeroSpaceClient struct {
	ogClient aerospacecli.AeroSpaceClient
	dryRun   bool
}

// AeroSpaceClientOpts defines options for creating a new AeroSpaceClient.
type AeroSpaceClientOpts struct {
	DryRun bool
}

// NewAeroSpaceClient creates a new AeroSpaceClient with the default settings.
func NewAeroSpaceClient(client aerospacecli.AeroSpaceClient) *AeroSpaceClient {
	return &AeroSpaceClient{
		ogClient: client,
		dryRun:   false, // Default dry-run is false
	}
}

// SetOptionssets the dry-run flag for the AeroSpaceClient.
func (c *AeroSpaceClient) SetOptions(opts AeroSpaceClientOpts) {
	c.dryRun = opts.DryRun
}

// All methods.
func (c *AeroSpaceClient) GetAllWindows() ([]aerospacecli.Window, error) {
	return c.ogClient.GetAllWindows()
}

func (c *AeroSpaceClient) GetAllWindowsByWorkspace(workspaceName string) ([]aerospacecli.Window, error) {
	return c.ogClient.GetAllWindowsByWorkspace(workspaceName)
}

func (c *AeroSpaceClient) GetFocusedWindow() (*aerospacecli.Window, error) {
	return c.ogClient.GetFocusedWindow()
}

func (c *AeroSpaceClient) SetFocusByWindowID(windowID int) error {
	if c.dryRun {
		fmt.Printf("[dry-run] SetFocusByWindowID(%d)\n", windowID)
		return nil
	}
	return c.ogClient.SetFocusByWindowID(windowID)
}

func (c *AeroSpaceClient) GetFocusedWorkspace() (*aerospacecli.Workspace, error) {
	return c.ogClient.GetFocusedWorkspace()
}

func (c *AeroSpaceClient) MoveWindowToWorkspace(windowID int, workspaceName string) error {
	if c.dryRun {
		fmt.Printf("[dry-run] MoveWindowToWorkspace(windowID=%d, workspace=%s)\n", windowID, workspaceName)
		return nil
	}
	return c.ogClient.MoveWindowToWorkspace(windowID, workspaceName)
}

func (c *AeroSpaceClient) SetLayout(windowID int, layout string) error {
	if c.dryRun {
		fmt.Printf("[dry-run] SetLayout(windowID=%d, layout=%s)\n", windowID, layout)
		return nil
	}
	return c.ogClient.SetLayout(windowID, layout)
}

func (c *AeroSpaceClient) Connection() socketcli.AeroSpaceConnection {
	return c.ogClient.Connection()
}

func (c *AeroSpaceClient) CloseConnection() error {
	if c.dryRun {
		fmt.Println("[dry-run] CloseConnection()")
		return nil
	}
	return c.ogClient.CloseConnection()
}
