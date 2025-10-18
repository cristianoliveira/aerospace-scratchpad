package aerospace

import (
	"fmt"
	"os"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	socketcli "github.com/cristianoliveira/aerospace-ipc/pkg/client"
)

// AeroSpaceClient implements the AeroSpaceClient interface for interacting with AeroSpaceWM.
//
//revive:disable:exported
type AeroSpaceClient struct {
	ogClient aerospacecli.AeroSpaceClient
	dryRun   bool
}

// ClientOpts defines options for creating a new AeroSpaceClient.
type ClientOpts struct {
	DryRun bool
}

// NewAeroSpaceClient creates a new AeroSpaceClient with the default settings.
func NewAeroSpaceClient(client aerospacecli.AeroSpaceClient) *AeroSpaceClient {
	return &AeroSpaceClient{
		ogClient: client,
		dryRun:   false, // Default dry-run is false
	}
}

// SetOptions the dry-run flag for the AeroSpaceClient.
func (c *AeroSpaceClient) SetOptions(opts ClientOpts) {
	c.dryRun = opts.DryRun
}

// GetAllWindows retrieves all windows managed by AeroSpaceWM.
func (c *AeroSpaceClient) GetAllWindows() ([]aerospacecli.Window, error) {
	return c.ogClient.GetAllWindows()
}

func (c *AeroSpaceClient) GetAllWindowsByWorkspace(
	workspaceName string,
) ([]aerospacecli.Window, error) {
	return c.ogClient.GetAllWindowsByWorkspace(workspaceName)
}

func (c *AeroSpaceClient) GetFocusedWindow() (*aerospacecli.Window, error) {
	return c.ogClient.GetFocusedWindow()
}

func (c *AeroSpaceClient) SetFocusByWindowID(windowID int) error {
	if c.dryRun {
		fmt.Fprintf(os.Stdout, "[dry-run] SetFocusByWindowID(%d)\n", windowID)
		return nil
	}
	return c.ogClient.SetFocusByWindowID(windowID)
}

// FocusNextTilingWindow moves focus to the next tiled window in depth-first order, ignoring floating windows.
// Equivalent to: `aerospace focus dfs-next --ignore-floating`.
func (c *AeroSpaceClient) FocusNextTilingWindow() error {
	if c.dryRun {
		fmt.Fprintln(os.Stdout, "[dry-run] FocusNextTilingWindow()")
		return nil
	}
	client := c.ogClient.Connection()
	response, err := client.SendCommand(
		"focus",
		[]string{
			"dfs-prev",
		},
	)
	if err != nil || response.ExitCode != 0 {
		response, err = client.SendCommand(
			"focus",
			[]string{
				"dfs-next",
			},
		)
		if err != nil {
			return err
		}

		if response.ExitCode != 0 {
			return fmt.Errorf("failed to focus next tiling window: %s", response.StdErr)
		}
	}

	return nil
}

func (c *AeroSpaceClient) GetFocusedWorkspace() (*aerospacecli.Workspace, error) {
	return c.ogClient.GetFocusedWorkspace()
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
	return c.ogClient.MoveWindowToWorkspace(windowID, workspaceName)
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
	return c.ogClient.SetLayout(windowID, layout)
}

func (c *AeroSpaceClient) Connection() socketcli.AeroSpaceConnection {
	return c.ogClient.Connection()
}

func (c *AeroSpaceClient) CloseConnection() error {
	if c.dryRun {
		fmt.Fprintln(os.Stdout, "[dry-run] CloseConnection()")
		return nil
	}
	return c.ogClient.CloseConnection()
}
