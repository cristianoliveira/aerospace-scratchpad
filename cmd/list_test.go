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
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

func TestListCmd(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)

	t.Run("lists scratchpad windows from workspace", func(t *testing.T) {
		command := "list"
		args := []string{command}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []windows.Window{
					{
						AppName:      "Notepad",
						WindowID:     1234,
						WindowLayout: "tiling",
						Workspace:    "ws1",
					},
					{
						AppName:      "Finder",
						WindowID:     5678,
						WindowLayout: "tiling",
						Workspace:    "ws1",
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
						AppName:      "Scratchpad Window",
						WindowID:     9999,
						WindowLayout: "floating",
						Workspace:    constants.DefaultScratchpadWorkspaceName,
					},
					{
						AppName:      "Another Scratchpad Window",
						WindowID:     8888,
						WindowLayout: "floating",
						Workspace:    constants.DefaultScratchpadWorkspaceName,
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: constants.DefaultScratchpadWorkspaceName,
				},
				FocusedWindowID: 0,
			},
		}

		allWindows := testutils.ExtractAllWindows(tree)
		scratchpadWindows := testutils.ExtractScratchpadWindows(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
				Return(scratchpadWindows.Windows, nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})

	t.Run("lists floating windows as scratchpad windows", func(t *testing.T) {
		command := "list"
		args := []string{command}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []windows.Window{
					{
						AppName:      "Floating Window",
						WindowID:     1111,
						WindowLayout: "floating",
						Workspace:    "ws1",
					},
					{
						AppName:      "Tiling Window",
						WindowID:     2222,
						WindowLayout: "tiling",
						Workspace:    "ws1",
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: "ws1",
				},
				FocusedWindowID: 2222,
			},
		}

		allWindows := testutils.ExtractAllWindows(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
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
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})

	t.Run("lists scratchpad windows with filters", func(t *testing.T) {
		command := "list"
		args := []string{command, "--filter", "app-name=^Scratchpad"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []windows.Window{
					{
						AppName:      "Scratchpad Window",
						WindowID:     9999,
						WindowLayout: "floating",
						Workspace:    constants.DefaultScratchpadWorkspaceName,
					},
					{
						AppName:      "Another Window",
						WindowID:     8888,
						WindowLayout: "floating",
						Workspace:    constants.DefaultScratchpadWorkspaceName,
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: constants.DefaultScratchpadWorkspaceName,
				},
				FocusedWindowID: 0,
			},
		}

		allWindows := testutils.ExtractAllWindows(tree)
		scratchpadWindows := testutils.ExtractScratchpadWindows(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
				Return(scratchpadWindows.Windows, nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})

	t.Run("lists scratchpad windows in json format", func(t *testing.T) {
		command := "list"
		args := []string{command, "--output", "json"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []windows.Window{
					{
						AppName:      "Scratchpad Window",
						WindowID:     9999,
						WindowLayout: "floating",
						Workspace:    constants.DefaultScratchpadWorkspaceName,
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: constants.DefaultScratchpadWorkspaceName,
				},
				FocusedWindowID: 0,
			},
		}

		allWindows := testutils.ExtractAllWindows(tree)
		scratchpadWindows := testutils.ExtractScratchpadWindows(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
				Return(scratchpadWindows.Windows, nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})

	t.Run("returns empty result when no scratchpad windows", func(t *testing.T) {
		command := "list"
		args := []string{command}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []windows.Window{
					{
						AppName:      "Tiling Window",
						WindowID:     2222,
						WindowLayout: "tiling",
						Workspace:    "ws1",
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: "ws1",
				},
				FocusedWindowID: 2222,
			},
		}

		allWindows := testutils.ExtractAllWindows(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
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
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})

	t.Run("works with ls alias", func(t *testing.T) {
		command := "ls"
		args := []string{command}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []windows.Window{
					{
						AppName:      "Scratchpad Window",
						WindowID:     9999,
						WindowLayout: "floating",
						Workspace:    constants.DefaultScratchpadWorkspaceName,
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: constants.DefaultScratchpadWorkspaceName,
				},
				FocusedWindowID: 0,
			},
		}

		allWindows := testutils.ExtractAllWindows(tree)
		scratchpadWindows := testutils.ExtractScratchpadWindows(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
				Return(scratchpadWindows.Windows, nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})
}
