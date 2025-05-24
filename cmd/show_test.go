package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"go.uber.org/mock/gomock"
)

func TestFunction(t *testing.T) {
	t.Run("moves the current focused window to scratchpad when empty", func(t *testing.T) {
		command := "show"
		args := []string{command, ""}

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
		}

		allWindows := testutils.ExtractAllWindows(tree)
		focusedTree := testutils.ExtractFocusedTree(tree)
		focusedWindow := testutils.ExtractFocusedWindow(tree)

		aerospaceClient := aerospacecli_mock.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetFocusedWindow().
				Return(focusedWindow, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedTree.Workspace, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetAllWindowsByWorkspace(focusedTree.Workspace.Workspace).
				Return(focusedTree.Windows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetFocusedWindow().
				Return(focusedWindow, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(focusedWindow.WindowID, "scratchpad").
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetLayout(focusedWindow.WindowID, "floating").
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

	t.Run(
		"set focus to window if already in the focused workspace but not focused",
		func(t *testing.T) {
			command := "show"
			args := []string{command, "Finder"}

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

					FocusedWindowId: 1234,
				},
			}

			allWindows := testutils.ExtractAllWindows(tree)
			focusedTree := testutils.ExtractFocusedTree(tree)
			focusedWindow := testutils.ExtractFocusedWindow(tree)

			aerospaceClient := aerospacecli_mock.NewMockAeroSpaceClient(ctrl)
			gomock.InOrder(
				aerospaceClient.EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1),

				aerospaceClient.EXPECT().
					GetFocusedWorkspace().
					Return(focusedTree.Workspace, nil).
					Times(1),

				aerospaceClient.EXPECT().
					GetAllWindowsByWorkspace(focusedTree.Workspace.Workspace).
					Return(focusedTree.Windows, nil).
					Times(1),

				aerospaceClient.EXPECT().
					GetFocusedWindow().
					Return(focusedWindow, nil).
					Times(1),

				aerospaceClient.EXPECT().
					SetFocusByWindowID(focusedTree.Windows[1].WindowID).
					Return(nil).
					Times(1),

				// DO NOT set the layout to floating
				aerospaceClient.EXPECT().
					SetLayout(gomock.Any(), "floating").
					Return(nil).
					Times(0),
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

	t.Run("moves a window to scratchpad by pattern", func(t *testing.T) {
		command := "show"
		args := []string{command, "Finder"}

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
		}

		allWindows := testutils.ExtractAllWindows(tree)
		focusedTree := testutils.ExtractFocusedTree(tree)
		focusedWindow := testutils.ExtractFocusedWindow(tree)

		aerospaceClient := aerospacecli_mock.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedTree.Workspace, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetAllWindowsByWorkspace("ws1").
				Return(focusedTree.Windows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetFocusedWindow().
				Return(focusedWindow, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(focusedWindow.WindowID, "scratchpad").
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetFocusByWindowID(focusedWindow.WindowID).
				Return(nil).
				Times(0),

			// When moving to scratchpad, set the layout to floating
			aerospaceClient.EXPECT().
				SetLayout(focusedWindow.WindowID, "floating").
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

	t.Run("summon the window to the current workspace if in another workspace", func(t *testing.T) {
		command := "show"
		args := []string{command, "Finder"}

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
				FocusedWindowId: 0, // Not focused
			},
			{
				Windows: []aerospacecli.Window{
					{
						AppName:  "Terminal",
						WindowID: 91011,
					},
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: "ws2",
				},
				FocusedWindowId: 91011,
			},
		}

		allWindows := testutils.ExtractAllWindows(tree)
		focusedTree := testutils.ExtractFocusedTree(tree)
		focusedWindow := testutils.ExtractFocusedWindow(tree)

		aerospaceClient := aerospacecli_mock.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedTree.Workspace, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetAllWindowsByWorkspace(focusedTree.Workspace.Workspace).
				Return(focusedTree.Windows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(
					tree[0].Windows[1].WindowID,
					focusedTree.Workspace.Workspace).
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetFocusByWindowID(
					tree[0].Windows[1].WindowID).
				Return(nil).
				Times(1),

			// When moving to scratchpad, set the layout to floating
			aerospaceClient.EXPECT().
				SetLayout(focusedWindow.WindowID, "floating").
				Return(nil).
				Times(0),
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
