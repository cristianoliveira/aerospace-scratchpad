/*
Copyright © 2025 Cristian Oliveira licence@cristianoliveira.dev
*/
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	aerospaceipc "github.com/cristianoliveira/aerospace-ipc"

	"github.com/spf13/cobra"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

const (
	bringWindowToWorkspaceCmd = "bring-window-to-workspace"

	minArgsBringWindow = 3
)

// MoveScratchpadResult holds the response from list-workspaces.
type MoveScratchpadResult struct {
	Workspace string `json:"workspace"`
	MonitorID int    `json:"monitor-id"`
}

func WorkspaceHandlerCmd(
	aerospaceClient aerospaceipc.AeroSpaceClient,
) *cobra.Command {
	var wsHandlerCmd = &cobra.Command{
		Use:   "workspace-handler <workspace>",
		Short: "This command handles when a window in scratchpad is focused (which shouldn't happen)",
		Long: `This command handles when a window within the scratchpad workspace is focused. It'll move the window to the last focused workspace and take the window too, behaving more like "summoning the window".
Add this snippet in your ~/aerospace.toml config:

'''toml
exec-on-workspace-change = ["/bin/bash", "-c",
  "aerospace-scratchpad wsh bring-window-to-workspace $AEROSPACE_PREV_WORKSPACE $AEROSPACE_FOCUSED_WORKSPACE"
]
'''
`,
		Aliases: []string{"wsh"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handler := newWorkspaceHandler(cmd, aerospaceClient)
			return handler.execute(args)
		},
	}

	return wsHandlerCmd
}

type workspaceHandler struct {
	cmd    *cobra.Command
	client aerospaceipc.AeroSpaceClient
	logger logger.Logger
}

func newWorkspaceHandler(
	cmd *cobra.Command,
	client aerospaceipc.AeroSpaceClient,
) *workspaceHandler {
	return &workspaceHandler{
		cmd:    cmd,
		client: client,
		logger: logger.GetDefaultLogger(),
	}
}

func (h *workspaceHandler) execute(args []string) error {
	if len(args) == 0 {
		return h.fail(
			"Error: missing subcommand",
			nil,
			"WSH: missing subcommand",
		)
	}

	switch args[0] {
	case bringWindowToWorkspaceCmd:
		if len(args) < minArgsBringWindow {
			return h.fail(
				"Error: missing <previous-workspace> <focused-workspace> arguments",
				nil,
				"WSH: not enough arguments",
			)
		}

		return h.handleBringWindowToWorkspace(args[1], args[2])
	default:
		return h.fail(
			fmt.Sprintf("Error: unknown subcommand %q", args[0]),
			nil,
			"WSH: unknown subcommand",
		)
	}
}

func (h *workspaceHandler) handleBringWindowToWorkspace(
	prevWorkspace string,
	focusedWorkspace string,
) error {
	h.logger.LogInfo(
		"WSH: bring-window-to-workspace invoked",
		"previous-workspace", prevWorkspace,
		"focused-workspace", focusedWorkspace,
	)

	if prevWorkspace == constants.DefaultScratchpadWorkspaceName {
		h.logger.LogDebug(
			"WSH: previous workspace is scratchpad, nothing to do",
			"workspace", prevWorkspace,
		)
		return nil
	}

	if focusedWorkspace != constants.DefaultScratchpadWorkspaceName {
		h.logger.LogDebug(
			"WSH: focused workspace is not scratchpad",
			"workspace", focusedWorkspace,
		)
		return nil
	}

	h.logger.LogInfo("WSH: focused workspace is scratchpad")

	focusedWindow, err := h.client.GetFocusedWindow()
	if err != nil {
		return h.fail(
			"Error: unable to get focused window",
			err,
			"WSH: unable to get focused window",
		)
	}

	h.logger.LogInfo("WSH: focused window", "window", focusedWindow)

	if focusedWindow.Workspace != constants.DefaultScratchpadWorkspaceName {
		h.logger.LogDebug(
			"WSH: focused window is no longer in scratchpad, skipping move",
			"workspace", focusedWindow.Workspace,
		)
		return nil
	}

	cleared, markerErr := h.clearMovingMarker()
	if markerErr != nil {
		return h.fail(
			"Error: unable to remove temp file",
			markerErr,
			"WSH: unable to remove temp file",
		)
	}

	if cleared {
		h.logger.LogInfo("WSH: temp file exists, returning")
		return nil
	}

	if moveErr := h.moveWindowToWorkspace(focusedWindow.WindowID, prevWorkspace); moveErr != nil {
		return moveErr
	}

	h.logger.LogInfo(
		"WSH: [final] moved window to new focused workspace",
		"workspace", prevWorkspace,
		"window", focusedWindow,
	)

	return nil
}

func (h *workspaceHandler) clearMovingMarker() (bool, error) {
	_, err := os.Stat(constants.TempScratchpadMovingFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	if removeErr := os.Remove(constants.TempScratchpadMovingFile); removeErr != nil {
		return true, removeErr
	}

	return true, nil
}

func (h *workspaceHandler) moveWindowToWorkspace(windowID int, workspace string) error {
	client := h.client.Connection()

	response, err := client.SendCommand(
		"move-node-to-workspace",
		[]string{
			workspace,
			"--window-id", strconv.Itoa(windowID),
			"--focus-follows-window",
		},
	)
	if err != nil {
		return h.fail(
			fmt.Sprintf("Error: unable to move window %d to workspace %s", windowID, workspace),
			err,
			"WSH: unable to move window to workspace",
		)
	}

	if response.ExitCode != 0 {
		return h.fail(
			fmt.Sprintf("Error: unable to move window %d to workspace %s", windowID, workspace),
			errors.New(response.StdErr),
			"WSH: unable to move window to workspace - non-zero exit",
		)
	}

	return nil
}

func (h *workspaceHandler) fail(userMessage string, err error, logMessage string) error {
	if err != nil {
		h.logger.LogError(logMessage, "error", err)
		h.cmd.PrintErrf("%s: %v\n", userMessage, err)
		return fmt.Errorf("%s: %w", userMessage, err)
	}

	h.logger.LogError(logMessage)
	h.cmd.PrintErrln(userMessage)
	return errors.New(userMessage)
}
