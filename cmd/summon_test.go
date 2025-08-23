package cmd

import (
	"fmt"
	"strings"
	"testing"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	mock_aerospace "github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"go.uber.org/mock/gomock"
)

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
		focusedWorkspace := &aerospacecli.Workspace{Workspace: "ws1"}
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}
		notepadWindow := windows[0]

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			// New behavior: get focused workspace before filtering windows
			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(notepadWindow.WindowID, focusedWorkspace.Workspace).
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetFocusByWindowID(notepadWindow.WindowID).
				Return(nil).
				Times(1),
		)

		cmd := RootCmd(aerospaceClient)
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
		focusedWorkspace := &aerospacecli.Workspace{Workspace: "ws1"}

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

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

	t.Run("fails when getting all windows returns an error", func(t *testing.T) {
		command := "summon"
		args := []string{command, "test"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(&aerospacecli.Workspace{Workspace: "ws1"}, nil).
				Times(1),
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(nil, fmt.Errorf("mocked_error")).
				Times(1),
		)

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

	t.Run("fails when getting focused workspace returns an error", func(t *testing.T) {
		command := "summon"
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
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: "ws1",
				},
				FocusedWindowId: 1234,
			},
		}
		_ = testutils.ExtractAllWindows(tree)

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		// New behavior: focused workspace is fetched before listing windows
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(nil, fmt.Errorf("mocked_error")).
				Times(1),
		)

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

	t.Run("fails when regex pattern is invalid", func(t *testing.T) {
		command := "summon"
		args := []string{command, "[invalid"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		focusedWorkspace := &aerospacecli.Workspace{Workspace: "ws1"}

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		aerospaceClient.EXPECT().
			GetFocusedWorkspace().
			Return(focusedWorkspace, nil)

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

	t.Run("fails when moving window to workspace returns an error", func(t *testing.T) {
		command := "summon"
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
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: "ws1",
				},
				FocusedWindowId: 1234,
			},
		}
		allWindows := testutils.ExtractAllWindows(tree)
		focusedWorkspace := &aerospacecli.Workspace{Workspace: "ws1"}
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}
		notepadWindow := windows[0]

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(notepadWindow.WindowID, focusedWorkspace.Workspace).
				Return(fmt.Errorf("mocked_move_error")).
				Times(1),
		)

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

	t.Run("fails when setting focus returns an error", func(t *testing.T) {
		command := "summon"
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
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: "ws1",
				},
				FocusedWindowId: 1234,
			},
		}
		allWindows := testutils.ExtractAllWindows(tree)
		focusedWorkspace := &aerospacecli.Workspace{Workspace: "ws1"}
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}
		notepadWindow := windows[0]

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(notepadWindow.WindowID, focusedWorkspace.Workspace).
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetFocusByWindowID(notepadWindow.WindowID).
				Return(fmt.Errorf("mocked_focus_error")).
				Times(1),
		)

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

	t.Run("summons multiple windows matching the pattern", func(t *testing.T) {
		command := "summon"
		args := []string{command, ".*(Notepad|TextEdit).*"}

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
						AppName:  "TextEdit",
						WindowID: 5678,
					},
					{
						AppName:  "Finder",
						WindowID: 9012,
					},
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: "ws1",
				},
				FocusedWindowId: 9012,
			},
		}
		allWindows := testutils.ExtractAllWindows(tree)
		focusedWorkspace := &aerospacecli.Workspace{Workspace: "ws1"}
		windows := testutils.ExtractWindowsByName(tree, ".*(Notepad|TextEdit).*")
		if len(windows) != 2 {
			t.Fatalf("Expected 2 windows matching pattern, got %d", len(windows))
		}

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(windows[0].WindowID, focusedWorkspace.Workspace).
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetFocusByWindowID(windows[0].WindowID).
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(windows[1].WindowID, focusedWorkspace.Workspace).
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetFocusByWindowID(windows[1].WindowID).
				Return(nil).
				Times(1),
		)

		cmd := RootCmd(aerospaceClient)
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

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)

		cmd := RootCmd(aerospaceClient)
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

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)

		cmd := RootCmd(aerospaceClient)
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
		focusedWorkspace := &aerospacecli.Workspace{Workspace: "ws1"}
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}
		notepadWindow := windows[0]

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedWorkspace, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			// In dry-run mode, the aerospace client wrapper intercepts these calls
			// and prints debug messages instead of calling the actual methods
			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(notepadWindow.WindowID, focusedWorkspace.Workspace).
				Return(nil).
				Times(0), // DO NOT RUN in dry-run mode

			aerospaceClient.EXPECT().
				SetFocusByWindowID(notepadWindow.WindowID).
				Return(nil).
				Times(0), // DO NOT RUN in dry-run mode
		)

		cmd := RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, errorMessage)
	})
}
