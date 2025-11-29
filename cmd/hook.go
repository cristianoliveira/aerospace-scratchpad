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

	"github.com/spf13/cobra"

	aerospaceipc "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

const (
	pullWindowSubcommand = "pull-window"

	minArgsPullWindow = 2
)

func HookCmd(
	aerospaceClient *aerospaceipc.AeroSpaceWM,
) *cobra.Command {
	hookCmd := &cobra.Command{
		Use:   "hook",
		Short: "Hook commands to react to specific actions outside AeroSpace WM",
		Long: `Hook commands to react to actions that aren't handled by AeroSpace WM.
Example of such action is when a window in the scratchpad workspace is focused, which happens when clicking in a notification or
when a program is focused by the launcher (alfred, raycast, etc).
`,
	}

	hookCmd.AddCommand(newPullWindowCmd(aerospaceClient))

	return hookCmd
}

func newPullWindowCmd(
	aerospaceClient *aerospaceipc.AeroSpaceWM,
) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s <previous-workspace> <focused-workspace>", pullWindowSubcommand),
		Short: "Pull the focused scratchpad window back to the previous workspace",
		Long: `Pull the focused scratchpad window back to the previous workspace so it behaves like it was summoned there.

This is usually hooked via exec-on-workspace-change.

Add this snippet in your aerospace.toml config:

'''toml
exec-on-workspace-change = ["/bin/bash", "-c",
  "aerospace-scratchpad hook pull-window $AEROSPACE_PREV_WORKSPACE $AEROSPACE_FOCUSED_WORKSPACE"
]
`,
		Aliases: []string{"pull"},
		Args:    cobra.ExactArgs(minArgsPullWindow),
		RunE: func(cmd *cobra.Command, args []string) error {
			handler := newHookHandler(cmd, aerospaceClient)
			return handler.handlePullWindow(args[0], args[1])
		},
	}
}

type hookHandler struct {
	cmd    *cobra.Command
	client *aerospaceipc.AeroSpaceWM
	logger logger.Logger
}

func newHookHandler(
	cmd *cobra.Command,
	client *aerospaceipc.AeroSpaceWM,
) *hookHandler {
	return &hookHandler{
		cmd:    cmd,
		client: client,
		logger: logger.GetDefaultLogger(),
	}
}

func (h *hookHandler) handlePullWindow(
	prevWorkspace string,
	focusedWorkspace string,
) error {
	h.logger.LogInfo(
		"HOOK: pull-window invoked",
		"previous-workspace", prevWorkspace,
		"focused-workspace", focusedWorkspace,
	)

	if prevWorkspace == constants.DefaultScratchpadWorkspaceName {
		h.logger.LogDebug(
			"HOOK: previous workspace is scratchpad, nothing to do",
			"workspace", prevWorkspace,
		)
		return nil
	}

	if focusedWorkspace != constants.DefaultScratchpadWorkspaceName {
		h.logger.LogDebug(
			"HOOK: focused workspace is not scratchpad",
			"workspace", focusedWorkspace,
		)
		return nil
	}

	h.logger.LogInfo("HOOK: focused workspace is scratchpad")

	focusedWindow, err := h.client.Windows().GetFocusedWindow()
	if err != nil {
		return h.fail(
			"Error: unable to get focused window",
			err,
			"HOOK: unable to get focused window",
		)
	}

	h.logger.LogInfo("HOOK: focused window", "window", focusedWindow)

	if focusedWindow.Workspace != constants.DefaultScratchpadWorkspaceName {
		h.logger.LogDebug(
			"HOOK: focused window is no longer in scratchpad, skipping move",
			"workspace", focusedWindow.Workspace,
		)
		return nil
	}

	cleared, markerErr := h.clearMovingMarker()
	if markerErr != nil {
		return h.fail(
			"Error: unable to remove temp file",
			markerErr,
			"HOOK: unable to remove temp file",
		)
	}

	if cleared {
		h.logger.LogInfo("HOOK: temp file exists, returning")
		return nil
	}

	if moveErr := h.moveWindowToWorkspace(focusedWindow.WindowID, prevWorkspace); moveErr != nil {
		return moveErr
	}

	h.logger.LogInfo(
		"HOOK: [final] moved window to new focused workspace",
		"workspace", prevWorkspace,
		"window", focusedWindow,
	)

	return nil
}

func (h *hookHandler) clearMovingMarker() (bool, error) {
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

func (h *hookHandler) moveWindowToWorkspace(windowID int, workspace string) error {
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
			"HOOK: unable to move window to workspace",
		)
	}

	if response.ExitCode != 0 {
		return h.fail(
			fmt.Sprintf("Error: unable to move window %d to workspace %s", windowID, workspace),
			errors.New(response.StdErr),
			"HOOK: unable to move window to workspace - non-zero exit",
		)
	}

	return nil
}

func (h *hookHandler) fail(userMessage string, err error, logMessage string) error {
	if err != nil {
		h.logger.LogError(logMessage, "error", err)
		h.cmd.PrintErrf("%s: %v\n", userMessage, err)
		return fmt.Errorf("%s: %w", userMessage, err)
	}

	h.logger.LogError(logMessage)
	h.cmd.PrintErrln(userMessage)
	return errors.New(userMessage)
}
