package cmd_test

import (
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

func TestNextCmd(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)

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

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
		testutils.MatchSnapshot(t, tree, cmdAsString, out, err)
	})

	t.Run(
		"fails when getting focused workspace returns an error",
		func(t *testing.T) {
			command := "next"
			args := []string{command}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			gomock.InOrder(
				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(nil, errors.New("mocked_error")).
					Times(1),
			)

			wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
			_ = wrappedClient
			cmd := cmd.RootCmd(aerospaceClient)
			out, err := testutils.CmdExecute(cmd, args...)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			if out != "" {
				t.Errorf("Expected empty output, got %s", out)
			}

			cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
			testutils.MatchSnapshot(t, nil, cmdAsString, out, err)
		},
	)

	t.Run(
		"fails when no scratchpad windows available",
		func(t *testing.T) {
			command := "next"
			args := []string{command}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}
			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			gomock.InOrder(
				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(focusedWorkspace, nil).
					Times(1),
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
					Return([]windows.Window{}, nil).
					Times(1),
			)

			wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
			_ = wrappedClient
			cmd := cmd.RootCmd(aerospaceClient)
			out, err := testutils.CmdExecute(cmd, args...)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			if out != "" {
				t.Errorf("Expected empty output, got %s", out)
			}

			cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
			testutils.MatchSnapshot(t, nil, cmdAsString, out, err)
		},
	)

	t.Run(
		"fails when moving window to workspace returns an error",
		func(t *testing.T) {
			command := "next"
			args := []string{command}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}
			scratchpadWindows := []windows.Window{
				{
					AppName:  "Scratchpad Window",
					WindowID: 9999,
				},
			}
			windowID := 9999
			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			gomock.InOrder(
				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(focusedWorkspace, nil).
					Times(1),
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
					Return(scratchpadWindows, nil).
					Times(1),
				aerospaceClient.GetWorkspacesMock().EXPECT().
					MoveWindowToWorkspaceWithOpts(
						workspaces.MoveWindowToWorkspaceArgs{
							WorkspaceName: focusedWorkspace.Workspace,
						},
						workspaces.MoveWindowToWorkspaceOpts{
							WindowID: &windowID,
						},
					).
					Return(errors.New("mocked_move_error")).
					Times(1),
			)

			wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
			_ = wrappedClient
			cmd := cmd.RootCmd(aerospaceClient)
			out, err := testutils.CmdExecute(cmd, args...)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			if out != "" {
				t.Errorf("Expected empty output, got %s", out)
			}

			cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
			testutils.MatchSnapshot(t, nil, cmdAsString, out, err)
		},
	)
}
