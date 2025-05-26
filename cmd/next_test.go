package cmd

import (
	"fmt"
	"strings"
	"testing"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"go.uber.org/mock/gomock"
)

func TestNextCmd(t *testing.T) {
	t.Run("summon next window from scratchpad", func(t *testing.T) {
		command := "next"
		args := []string{command}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []aerospacecli.Window{
					{
						AppName:  "Notepad",
						WindowID: 1234,
					},
					{
						AppName:  "Finder",
						WindowID: 5678,
					},
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: "ws1",
				},

				FocusedWindowId: 5678,
			},
			{
				Windows: []aerospacecli.Window{
					{
						AppName:  "Scratchpad Window",
						WindowID: 9999,
					},
					{
						AppName:  "Another Scratchpad Window",
						WindowID: 8888,
					},
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: constants.DefaultScratchpadWorkspaceName,
				},

				FocusedWindowId: 0,
			},
		}

		focusedTree := testutils.ExtractFocusedTree(tree)
		scratchpadWindows := testutils.ExtractScratchpadWindows(tree)

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedTree.Workspace, nil).
				Times(1),
			aerospaceClient.EXPECT().
				GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
				Return(scratchpadWindows.Windows, nil).
				Times(1),
			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(
					9999, // The ID of the first scratchpad window
					focusedTree.Workspace.Workspace,
				).
				Return(nil).
				Times(1),
			aerospaceClient.EXPECT().
				SetFocusByWindowID(9999). // Focus the moved window
				Return(nil).
				Times(1),
		)

		cmd := RootCmd(aerospaceClient)
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
