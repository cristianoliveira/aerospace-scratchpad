package cmd_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

//nolint:gocognit // Integration test covers multiple window flows in one place
func TestSummonCmd(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)

	t.Run("successfully summons a window by pattern", func(t *testing.T) {
		command := "summon"
		args := []string{command, "Notepad"}

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
		}
		allWindows := testutils.ExtractAllWindows(tree)
		focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}
		matchedWindows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(matchedWindows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(matchedWindows))
		}
		notepadWindow := matchedWindows[0]

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			// New behavior: get focused workspace before filtering windows
			aerospaceClient.GetWorkspacesMock().EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspaceWithOpts(
					workspaces.MoveWindowToWorkspaceArgs{
						WorkspaceName: focusedWorkspace.Workspace,
					},
					workspaces.MoveWindowToWorkspaceOpts{
						WindowID: &notepadWindow.WindowID,
					},
				).
				Return(nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				SetFocusByWindowID(windows.SetFocusArgs{
					WindowID: notepadWindow.WindowID,
				}).
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

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("fails when pattern doesn't match any window", func(t *testing.T) {
		command := "summon"
		args := []string{command, "NonExistentApp"}

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
		}
		allWindows := testutils.ExtractAllWindows(tree)
		focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWorkspacesMock().EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})

	t.Run(
		"fails when getting all windows returns an error",
		func(t *testing.T) {
			command := "summon"
			args := []string{command, "test"}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			gomock.InOrder(
				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(&workspaces.Workspace{Workspace: "ws1"}, nil).
					Times(1),
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindows().
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

			cmdAsString := "aerospace-scratchpad " + strings.Join(
				args,
				" ",
			) + "\n"
			errorMessage := fmt.Sprintf("Error\n %+v", err)
			snaps.MatchSnapshot(t, cmdAsString, "Output", out, errorMessage)
		},
	)

	t.Run(
		"fails when getting focused workspace returns an error",
		func(t *testing.T) {
			command := "summon"
			args := []string{command, "Notepad"}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{
				{
					Windows: []windows.Window{
						{
							AppName:  "Notepad",
							WindowID: 1234,
						},
					},
					Workspace: &workspaces.Workspace{
						Workspace: "ws1",
					},
					FocusedWindowID: 1234,
				},
			}
			_ = testutils.ExtractAllWindows(tree)

			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			// New behavior: focused workspace is fetched before listing windows
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

			cmdAsString := "aerospace-scratchpad " + strings.Join(
				args,
				" ",
			) + "\n"
			errorMessage := fmt.Sprintf("Error\n %+v", err)
			snaps.MatchSnapshot(t, cmdAsString, "Output", out, errorMessage)
		},
	)

	t.Run("fails when regex pattern is invalid", func(t *testing.T) {
		command := "summon"
		args := []string{command, "[invalid"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(focusedWorkspace, nil)

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

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, cmdAsString, "Output", out, errorMessage)
	})

	t.Run(
		"fails when moving window to workspace returns an error",
		func(t *testing.T) {
			command := "summon"
			args := []string{command, "Notepad"}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{
				{
					Windows: []windows.Window{
						{
							AppName:  "Notepad",
							WindowID: 1234,
						},
					},
					Workspace: &workspaces.Workspace{
						Workspace: "ws1",
					},
					FocusedWindowID: 1234,
				},
			}
			allWindows := testutils.ExtractAllWindows(tree)
			focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}
			windows := testutils.ExtractWindowsByName(tree, "Notepad")
			if len(windows) != 1 {
				t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
			}
			notepadWindow := windows[0]

			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			gomock.InOrder(
				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(focusedWorkspace, nil).
					Times(1),

				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1),

				aerospaceClient.GetWorkspacesMock().EXPECT().
					MoveWindowToWorkspaceWithOpts(
						workspaces.MoveWindowToWorkspaceArgs{
							WorkspaceName: focusedWorkspace.Workspace,
						},
						workspaces.MoveWindowToWorkspaceOpts{
							WindowID: &notepadWindow.WindowID,
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

			cmdAsString := "aerospace-scratchpad " + strings.Join(
				args,
				" ",
			) + "\n"
			errorMessage := fmt.Sprintf("Error\n %+v", err)
			snaps.MatchSnapshot(t, cmdAsString, "Output", out, errorMessage)
		},
	)

	t.Run("fails when setting focus returns an error", func(t *testing.T) {
		command := "summon"
		args := []string{command, "Notepad"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []windows.Window{
					{
						AppName:  "Notepad",
						WindowID: 1234,
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: "ws1",
				},
				FocusedWindowID: 1234,
			},
		}
		allWindows := testutils.ExtractAllWindows(tree)
		focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}
		matchedWindows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(matchedWindows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(matchedWindows))
		}
		notepadWindow := matchedWindows[0]

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWorkspacesMock().EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspaceWithOpts(
					workspaces.MoveWindowToWorkspaceArgs{
						WorkspaceName: focusedWorkspace.Workspace,
					},
					workspaces.MoveWindowToWorkspaceOpts{
						WindowID: &notepadWindow.WindowID,
					},
				).
				Return(nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				SetFocusByWindowID(windows.SetFocusArgs{
					WindowID: notepadWindow.WindowID,
				}).
				Return(errors.New("mocked_focus_error")).
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

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("summons multiple windows matching the pattern", func(t *testing.T) {
		command := "summon"
		args := []string{command, ".*(Notepad|TextEdit).*"}

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
						AppName:  "TextEdit",
						WindowID: 5678,
					},
					{
						AppName:  "Finder",
						WindowID: 9012,
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: "ws1",
				},
				FocusedWindowID: 9012,
			},
		}
		allWindows := testutils.ExtractAllWindows(tree)
		focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}
		matchedWindows := testutils.ExtractWindowsByName(
			tree,
			".*(Notepad|TextEdit).*",
		)
		if len(matchedWindows) != 2 {
			t.Fatalf(
				"Expected 2 windows matching pattern, got %d",
				len(matchedWindows),
			)
		}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWorkspacesMock().EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspaceWithOpts(
					workspaces.MoveWindowToWorkspaceArgs{
						WorkspaceName: focusedWorkspace.Workspace,
					},
					workspaces.MoveWindowToWorkspaceOpts{
						WindowID: &matchedWindows[0].WindowID,
					},
				).
				Return(nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				SetFocusByWindowID(windows.SetFocusArgs{
					WindowID: matchedWindows[0].WindowID,
				}).
				Return(nil).
				Times(1),

			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspaceWithOpts(
					workspaces.MoveWindowToWorkspaceArgs{
						WorkspaceName: focusedWorkspace.Workspace,
					},
					workspaces.MoveWindowToWorkspaceOpts{
						WindowID: &matchedWindows[1].WindowID,
					},
				).
				Return(nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				SetFocusByWindowID(windows.SetFocusArgs{
					WindowID: matchedWindows[1].WindowID,
				}).
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

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("handles empty pattern gracefully", func(t *testing.T) {
		command := "summon"
		args := []string{command, ""}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error for empty pattern, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("handles whitespace-only pattern gracefully", func(t *testing.T) {
		command := "summon"
		args := []string{command, "   "}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error for whitespace-only pattern, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("[dry-run] summons a window by pattern", func(t *testing.T) {
		command := "summon"
		args := []string{command, "Notepad", "--dry-run"}

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
		}
		allWindows := testutils.ExtractAllWindows(tree)
		focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			aerospaceClient.GetWorkspacesMock().EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			// In dry-run mode, the aerospace client wrapper intercepts these calls
			// and prints debug messages instead of calling the actual methods
			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(0), // DO NOT RUN in dry-run mode

			aerospaceClient.GetWindowsMock().EXPECT().
				SetFocusByWindowID(gomock.Any()).
				Return(nil).
				Times(0), // DO NOT RUN in dry-run mode
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})
}
