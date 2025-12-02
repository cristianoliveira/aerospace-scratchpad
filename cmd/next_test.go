package cmd_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

func TestNextCmd(t *testing.T) {
	t.Run("summon next window from scratchpad", func(t *testing.T) {
		command := "next"
		args := []string{command}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []windows.Window{
					{
						AppName:  "Notepad",
						WindowID: 1234,
					},
					{
						AppName:  "Finder",
						WindowID: 5678,
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: "ws1",
				},

				FocusedWindowID: 5678,
			},
			{
				Windows: []windows.Window{
					{
						AppName:  "Scratchpad Window",
						WindowID: 9999,
					},
					{
						AppName:  "Another Scratchpad Window",
						WindowID: 8888,
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: constants.DefaultScratchpadWorkspaceName,
				},

				FocusedWindowID: 0,
			},
		}

		focusedTree := testutils.ExtractFocusedTree(tree)
		scratchpadWindows := testutils.ExtractScratchpadWindows(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		windowID := 9999
		gomock.InOrder(
			aerospaceClient.GetWorkspacesMock().EXPECT().
				GetFocusedWorkspace().
				Return(focusedTree.Workspace, nil).
				Times(1),
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
				Return(scratchpadWindows.Windows, nil).
				Times(1),
			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspaceWithOpts(
					workspaces.MoveWindowToWorkspaceArgs{
						WorkspaceName: focusedTree.Workspace.Workspace,
					},
					workspaces.MoveWindowToWorkspaceOpts{
						WindowID: &windowID,
					},
				).
				Return(nil).
				Times(1),
			aerospaceClient.GetFocusMock().EXPECT().
				SetFocusByWindowID(9999). // Focus the moved window
				Return(nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if out == "" {
			t.Errorf("Expected output, got empty string")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})
}
