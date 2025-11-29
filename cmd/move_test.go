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
	clientipc "github.com/cristianoliveira/aerospace-ipc/pkg/client"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

// dummyConn is a no-op implementation of clientipc.AeroSpaceConnection used by tests to satisfy FocusNextTilingWindow.
type dummyConn struct{}

// CloseConnection is a no-op.
func (d dummyConn) CloseConnection() error { return nil }

// GetSocketPath is a no-op.
func (d dummyConn) GetSocketPath() (string, error) { return "", nil }

// GetServerVersion is a no-op.
func (d dummyConn) GetServerVersion() (string, error) { return "", nil }

// CheckServerVersion is a no-op.
func (d dummyConn) CheckServerVersion() error { return nil }

// SendCommand always returns a successful zero-exit response.
func (d dummyConn) SendCommand(cmd string, args []string) (*clientipc.Response, error) {
	return &clientipc.Response{ExitCode: 0}, nil
}

//nolint:gocognit // Integration-style test exercises multiple window scenarios for coverage
func TestMoveCmd(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)

	t.Run("fails when pattern doesnt match any window", func(t *testing.T) {
		logger.SetDefaultLogger(&testutils.TestingLogger{
			Logger: func(msg string, largs ...any) {
				t.Logf(msg, largs...)
			},
		})
		t.Cleanup(func() {
			logger.SetDefaultLogger(&logger.EmptyLogger{})
		})

		command := "move"
		args := []string{command, "foo"}

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
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// Connection() is handled by routing connection, no need to mock
		gomock.InOrder(
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

	t.Run("moves current focused window to scratchpad when empty", func(t *testing.T) {
		command := "move"
		args := []string{command, ""}

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
		focusedWindow := testutils.ExtractFocusedWindow(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// allow FocusNextTilingWindow to run without mocking Connection details
		// Connection() is handled by routing connection, no need to mock
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetFocusedWindow().
				Return(focusedWindow, nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspace(
					focusedWindow.WindowID,
					constants.DefaultScratchpadWorkspaceName,
				).
				Return(nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				SetLayout(focusedWindow.WindowID, "floating").
				Return(nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(
			args,
			" ",
		) + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(
			t,
			tree,
			cmdAsString,
			"Output",
			out,
			errorMessage,
		)
	})

	t.Run("moves only the focused window when multiple matches exist", func(t *testing.T) {
		command := "move"
		args := []string{command, ""}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []windows.Window{
					{AppName: "Finder", WindowID: 1111},
					{AppName: "Finder", WindowID: 5678},
				},
				Workspace: &workspaces.Workspace{
					Workspace: "ws1",
				},

				FocusedWindowID: 5678,
			},
		}
		allWindows := testutils.ExtractAllWindows(tree)
		focusedWindow := testutils.ExtractFocusedWindow(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// allow FocusNextTilingWindow to run without mocking Connection details
		// Connection() is handled by routing connection, no need to mock
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetFocusedWindow().
				Return(focusedWindow, nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspace(
					focusedWindow.WindowID,
					constants.DefaultScratchpadWorkspaceName,
				).
				Return(nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				SetLayout(focusedWindow.WindowID, "floating").
				Return(nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(
			args,
			" ",
		) + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(
			t,
			tree,
			cmdAsString,
			"Output",
			out,
			errorMessage,
		)
	})

	t.Run("fails when getting all windows return an erro", func(t *testing.T) {
		command := "move"
		args := []string{command, "test"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// allow FocusNextTilingWindow if it were called; skip Connection
		// Connection() is handled by routing connection, no need to mock
		// Move now validates regex and then tries to query via querier which
		// calls GetAllWindows under the hood. We don't control that directly here,
		// so simulate that the overall result is an error by letting
		// GetAllWindows return an error and ensure we don't crash but surface the
		// generic no-match message (since querier error isn't surfaced).
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			Return(nil, errors.New("mocked_error")).
			Times(1)

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

	t.Run("moves a window to scratchpad by pattern", func(t *testing.T) {
		command := "move"
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
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}
		notepadWindow := windows[0]

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// allow FocusNextTilingWindow to run without mocking Connection details
		// Connection() is handled by routing connection, no need to mock
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspace(
					notepadWindow.WindowID,
					constants.DefaultScratchpadWorkspaceName,
				).
				Return(nil).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				SetLayout(notepadWindow.WindowID, "floating").
				Return(nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
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
		focusedWindow := testutils.ExtractFocusedWindow(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// allow FocusNextTilingWindow to run without mocking Connection details
		// Connection() is handled by routing connection, no need to mock
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspace(
					focusedWindow.WindowID,
					constants.DefaultScratchpadWorkspaceName,
				).
				Return(
					fmt.Errorf(
						"Window '%+v' already belongs to scratchpad",
						focusedWindow,
					),
				).
				Times(1),

			aerospaceClient.GetWindowsMock().EXPECT().
				SetLayout(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(0),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
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
		windows := testutils.ExtractWindowsByName(tree, "Notepad")
		if len(windows) != 1 {
			t.Fatalf("Expected 1 Notepad window, got %d", len(windows))
		}
		notepadWindow := windows[0]

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// allow wrapper.FocusNextTilingWindow in dry-run (will not call Connection)
		// Connection() is handled by routing connection, no need to mock
		gomock.InOrder(
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),

			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspace(
					notepadWindow.WindowID,
					constants.DefaultScratchpadWorkspaceName,
				).
				Return(nil).
				Times(0), // DO NOT RUN

			aerospaceClient.GetWindowsMock().EXPECT().
				SetLayout(notepadWindow.WindowID, "floating").
				Return(nil).
				Times(0), // DO NOT RUN
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(
			args,
			" ",
		) + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(
			t,
			tree,
			cmdAsString,
			"Output",
			out,
			errorMessage,
		)
	})

	t.Run(
		"moves all windows with the same app name as the focused window when --all is used",
		func(t *testing.T) {
			command := "move"
			args := []string{command, "", "--all"}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{
				{
					Windows: []windows.Window{
						{AppName: "Finder", WindowID: 1111},
						{AppName: "Finder", WindowID: 5678},
					},
					Workspace: &workspaces.Workspace{
						Workspace: "ws1",
					},

					FocusedWindowID: 5678,
				},
			}
			allWindows := testutils.ExtractAllWindows(tree)
			focusedWindow := testutils.ExtractFocusedWindow(tree)

			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			// allow FocusNextTilingWindow to run without mocking Connection details
			// Connection() is handled by routing connection, no need to mock

			// Setup expectations without strict ordering
			aerospaceClient.GetWindowsMock().EXPECT().
				GetFocusedWindow().
				Return(focusedWindow, nil)

			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil)

			// For each window, expect a pair of calls (MoveWindowToWorkspace followed by SetLayout)
			// Window 1 (5678)
			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspace(5678, constants.DefaultScratchpadWorkspaceName).
				Return(nil)
			aerospaceClient.GetWindowsMock().EXPECT().
				SetLayout(5678, "floating").
				Return(nil)

			// Window 2 (1111)
			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspace(1111, constants.DefaultScratchpadWorkspaceName).
				Return(nil)
			aerospaceClient.GetWindowsMock().EXPECT().
				SetLayout(1111, "floating").
				Return(nil)

			wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
			out, err := testutils.CmdExecute(cmd, args...)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			cmdAsString := "aerospace-scratchpad " + strings.Join(
				args,
				" ",
			) + "\n"
			errorMessage := fmt.Sprintf("Error\n %+v", err)
			snaps.MatchSnapshot(
				t,
				tree,
				cmdAsString,
				"Output",
				out,
				errorMessage,
			)
		},
	)

	t.Run(
		"[dry-run] moves all windows with the same app name as the focused window when --all is used",
		func(t *testing.T) {
			command := "move"
			args := []string{command, "", "--all", "--dry-run"}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{
				{
					Windows: []windows.Window{
						{AppName: "Finder", WindowID: 1111},
						{AppName: "Finder", WindowID: 5678},
					},
					Workspace: &workspaces.Workspace{
						Workspace: "ws1",
					},

					FocusedWindowID: 5678,
				},
			}
			allWindows := testutils.ExtractAllWindows(tree)
			focusedWindow := testutils.ExtractFocusedWindow(tree)

			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			// allow wrapper.FocusNextTilingWindow in dry-run (will not call Connection)
			// Connection() is handled by routing connection, no need to mock
			gomock.InOrder(
		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
					Return(focusedWindow, nil).
					Times(1),

				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1),

				// DO NOT RUN in dry-run mode
		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspace(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(0),

		aerospaceClient.GetWindowsMock().EXPECT().
			SetLayout(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(0),
			)

			wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
			out, err := testutils.CmdExecute(cmd, args...)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			cmdAsString := "aerospace-scratchpad " + strings.Join(
				args,
				" ",
			) + "\n"
			errorMessage := fmt.Sprintf("Error\n %+v", err)
			snaps.MatchSnapshot(
				t,
				tree,
				cmdAsString,
				"Output",
				out,
				errorMessage,
			)
		},
	)
}
