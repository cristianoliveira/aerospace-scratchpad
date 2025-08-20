package cmd

import (
	"fmt"
	"strings"
	"testing"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"go.uber.org/mock/gomock"
)

func TestMoveCmd(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)

	t.Run("fails when pattern doesnt match any window", func(t *testing.T) {
		logger.SetDefaultLogger(&testutils.TestingLogger{
			Logger: func(msg string, largs ...any) {
				t.Logf(msg, largs...)
			},
		})

		command := "move"
		args := []string{command, "foo"}

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
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
		)

		cmd := RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("moves current focused window to scratchpad when empty", func(t *testing.T) {
		command := "move"
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
		focusedWindow := testutils.ExtractFocusedWindow(tree)

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
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
				MoveWindowToWorkspace(focusedWindow.WindowID, constants.DefaultScratchpadWorkspaceName).
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
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("fails when getting all windows return an erro", func(t *testing.T) {
		command := "move"
		args := []string{command, "test"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		aerospaceClient.EXPECT().
			GetAllWindows().
			Return(nil, fmt.Errorf("mocked_error")).
			Times(1)

		cmd := RootCmd(aerospaceClient)
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

	t.Run("moves a window to scratchpad by pattern", func(t *testing.T) {
		command := "move"
		args := []string{command, "Notepad"}

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
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}
		notepadWindow := windows[0]

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(notepadWindow.WindowID, constants.DefaultScratchpadWorkspaceName).
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetLayout(notepadWindow.WindowID, "floating").
				Return(nil).
				Times(1),
		)

		cmd := RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("fails when moving a window to scratchpad", func(t *testing.T) {
		command := "move"
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
		focusedWindow := testutils.ExtractFocusedWindow(tree)

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(
					focusedWindow.WindowID,
					constants.DefaultScratchpadWorkspaceName).
				Return(fmt.Errorf("Window '%+v' already belongs to scratchpad", focusedWindow)).
				Times(1),

			aerospaceClient.EXPECT().
				SetLayout(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(0),
		)

		cmd := RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("[dry-run] move a window to scratchpad by pattern", func(t *testing.T) {
		command := "move"
		args := []string{command, "Notepad", "--dry-run"}

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
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}
		notepadWindow := windows[0]

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(notepadWindow.WindowID, constants.DefaultScratchpadWorkspaceName).
				Return(nil).
				Times(0), // DO NOT RUN

			aerospaceClient.EXPECT().
				SetLayout(notepadWindow.WindowID, "floating").
				Return(nil).
				Times(0), // DO NOT RUN
		)

		cmd := RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})
}
