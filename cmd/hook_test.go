package cmd_test

import (
	"errors"
	"os"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

type hookLogEvent struct {
	level string
	msg   string
	args  []any
}

type hookRecordingLogger struct {
	events []hookLogEvent
}

func (l *hookRecordingLogger) record(level, msg string, args ...any) {
	l.events = append(l.events, hookLogEvent{level: level, msg: msg, args: args})
}
func (l *hookRecordingLogger) LogInfo(msg string, args ...any)  { l.record("INFO", msg, args...) }
func (l *hookRecordingLogger) LogError(msg string, args ...any) { l.record("ERROR", msg, args...) }
func (l *hookRecordingLogger) LogWarn(msg string, args ...any)  { l.record("WARN", msg, args...) }
func (l *hookRecordingLogger) LogDebug(msg string, args ...any) { l.record("DEBUG", msg, args...) }
func (l *hookRecordingLogger) Close() error                     { return nil }
func (l *hookRecordingLogger) GetConfig() logger.LogConfig      { return logger.LogConfig{} }
func (l *hookRecordingLogger) AsJSON(any) string                { return "" }

func cleanupMarkerFile(t *testing.T) {
	t.Helper()

	err := os.Remove(constants.TempScratchpadMovingFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to clean marker file: %v", err)
	}
}

// scratchpadWorkspaceNames returns a slice of scratchpad workspace names to test.
func scratchpadWorkspaceNames() []string {
	return []string{
		".scratchpad",
		".scratchpad.1",
		".scratchpad.2",
		".scratchpad.10",
	}
}

func TestHookPullWindow(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})

	t.Run("logs one info outcome for a move", func(t *testing.T) {
		cleanupMarkerFile(t)
		recorder := &hookRecordingLogger{}
		logger.SetDefaultLogger(recorder)
		t.Cleanup(func() { logger.SetDefaultLogger(&logger.EmptyLogger{}) })

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		focusedWindow := &windows.Window{WindowID: 99, Workspace: constants.DefaultScratchpadWorkspaceName}
		gomock.InOrder(
			mockClient.GetWindowsMock().EXPECT().GetFocusedWindow().Return(focusedWindow, nil),
			mockClient.GetWorkspacesMock().EXPECT().MoveWindowToWorkspaceWithOpts(
				workspaces.MoveWindowToWorkspaceArgs{WorkspaceName: "prev-ws"},
				workspaces.MoveWindowToWorkspaceOpts{WindowID: &focusedWindow.WindowID},
			).Return(nil),
		)

		_, err := testutils.CmdExecute(cmd.RootCmd(mockClient), "hook", "pull-window", "prev-ws", constants.DefaultScratchpadWorkspaceName)
		if err != nil {
			t.Fatalf("expected success, got error %v", err)
		}
		if len(recorder.events) != 1 || recorder.events[0].level != "INFO" || recorder.events[0].msg != "HOOK: [final] moved window to new focused workspace" {
			t.Fatalf("expected one move outcome log, got %+v", recorder.events)
		}
	})

	t.Run("logs one debug outcome for a non-scratchpad workspace", func(t *testing.T) {
		recorder := &hookRecordingLogger{}
		logger.SetDefaultLogger(recorder)
		t.Cleanup(func() { logger.SetDefaultLogger(&logger.EmptyLogger{}) })
		_, err := testutils.CmdExecute(cmd.RootCmd(testutils.NewMockAeroSpaceWM(gomock.NewController(t))), "hook", "pull-window", "prev-ws", "work")
		if err != nil {
			t.Fatalf("expected success, got error %v", err)
		}
		if len(recorder.events) != 1 || recorder.events[0].level != "DEBUG" || recorder.events[0].msg != "HOOK: pull-window skipped" {
			t.Fatalf("expected one skip outcome log, got %+v", recorder.events)
		}
	})

	t.Run("moves focused scratchpad window to previous workspace", func(t *testing.T) {
		cleanupMarkerFile(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)

		focusedWindow := &windows.Window{
			WindowID:  99,
			Workspace: constants.DefaultScratchpadWorkspaceName,
		}

		gomock.InOrder(
			mockClient.GetWindowsMock().
				EXPECT().
				GetFocusedWindow().
				Return(focusedWindow, nil).
				Times(1),
			mockClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspaceWithOpts(
					workspaces.MoveWindowToWorkspaceArgs{
						WorkspaceName: "prev-ws",
					},
					workspaces.MoveWindowToWorkspaceOpts{
						WindowID: &focusedWindow.WindowID,
					},
				).
				Return(nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(mockClient)
		_ = wrappedClient
		rootCmd := cmd.RootCmd(mockClient)
		_, err := testutils.CmdExecute(
			rootCmd,
			"hook",
			"pull-window",
			"prev-ws",
			constants.DefaultScratchpadWorkspaceName,
		)

		if err != nil {
			t.Fatalf("expected success, got error %v", err)
		}
	})

	t.Run("skips when previous workspace is scratchpad", func(t *testing.T) {
		cleanupMarkerFile(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)

		rootCmd := cmd.RootCmd(mockClient)
		_, err := testutils.CmdExecute(
			rootCmd,
			"hook",
			"pull-window",
			constants.DefaultScratchpadWorkspaceName,
			constants.DefaultScratchpadWorkspaceName,
		)

		if err != nil {
			t.Fatalf("expected success, got error %v", err)
		}
	})

	t.Run("skips move when marker file exists", func(t *testing.T) {
		cleanupMarkerFile(t)

		err := os.WriteFile(constants.TempScratchpadMovingFile, []byte("moving"), 0o600)
		if err != nil {
			t.Fatalf("failed to create marker file: %v", err)
		}
		t.Cleanup(func() {
			cleanupMarkerFile(t)
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)

		focusedWindow := &windows.Window{
			WindowID:  124,
			Workspace: constants.DefaultScratchpadWorkspaceName,
		}

		mockClient.GetWindowsMock().EXPECT().GetFocusedWindow().Return(focusedWindow, nil).Times(1)

		wrappedClient := aerospace.NewAeroSpaceClient(mockClient)
		_ = wrappedClient
		rootCmd := cmd.RootCmd(mockClient)
		_, execErr := testutils.CmdExecute(
			rootCmd,
			"hook",
			"pull-window",
			"prev-ws",
			constants.DefaultScratchpadWorkspaceName,
		)

		if execErr != nil {
			t.Fatalf("expected success, got error %v", execErr)
		}
	})

	t.Run(
		"fails when getting focused window returns an error",
		func(t *testing.T) {
			cleanupMarkerFile(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := testutils.NewMockAeroSpaceWM(ctrl)
			mockClient.GetWindowsMock().EXPECT().
				GetFocusedWindow().
				Return(nil, errors.New("mocked_error")).
				Times(1)

			wrappedClient := aerospace.NewAeroSpaceClient(mockClient)
			_ = wrappedClient
			rootCmd := cmd.RootCmd(mockClient)
			_, err := testutils.CmdExecute(
				rootCmd,
				"hook",
				"pull-window",
				"prev-ws",
				constants.DefaultScratchpadWorkspaceName,
			)

			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		},
	)

	t.Run(
		"fails when moving window returns an error",
		func(t *testing.T) {
			cleanupMarkerFile(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := testutils.NewMockAeroSpaceWM(ctrl)

			focusedWindow := &windows.Window{
				WindowID:  99,
				Workspace: constants.DefaultScratchpadWorkspaceName,
			}

			gomock.InOrder(
				mockClient.GetWindowsMock().
					EXPECT().
					GetFocusedWindow().
					Return(focusedWindow, nil).
					Times(1),
				mockClient.GetWorkspacesMock().EXPECT().
					MoveWindowToWorkspaceWithOpts(
						workspaces.MoveWindowToWorkspaceArgs{
							WorkspaceName: "prev-ws",
						},
						workspaces.MoveWindowToWorkspaceOpts{
							WindowID: &focusedWindow.WindowID,
						},
					).
					Return(errors.New("mocked_move_error")).
					Times(1),
			)

			wrappedClient := aerospace.NewAeroSpaceClient(mockClient)
			_ = wrappedClient
			rootCmd := cmd.RootCmd(mockClient)
			_, err := testutils.CmdExecute(
				rootCmd,
				"hook",
				"pull-window",
				"prev-ws",
				constants.DefaultScratchpadWorkspaceName,
			)

			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		},
	)
}

func TestHookPullWindowPerMonitor(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})

	t.Run("works with per-monitor scratchpad workspaces", func(t *testing.T) {
		for _, scratchpadWorkspace := range scratchpadWorkspaceNames() {
			t.Run(scratchpadWorkspace, func(t *testing.T) {
				cleanupMarkerFile(t)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockClient := testutils.NewMockAeroSpaceWM(ctrl)

				focusedWindow := &windows.Window{
					WindowID:  99,
					Workspace: scratchpadWorkspace,
				}

				gomock.InOrder(
					mockClient.GetWindowsMock().
						EXPECT().
						GetFocusedWindow().
						Return(focusedWindow, nil).
						Times(1),
					mockClient.GetWorkspacesMock().EXPECT().
						MoveWindowToWorkspaceWithOpts(
							workspaces.MoveWindowToWorkspaceArgs{
								WorkspaceName: "prev-ws",
							},
							workspaces.MoveWindowToWorkspaceOpts{
								WindowID: &focusedWindow.WindowID,
							},
						).
						Return(nil).
						Times(1),
				)

				wrappedClient := aerospace.NewAeroSpaceClient(mockClient)
				_ = wrappedClient
				rootCmd := cmd.RootCmd(mockClient)
				_, err := testutils.CmdExecute(
					rootCmd,
					"hook",
					"pull-window",
					"prev-ws",
					scratchpadWorkspace,
				)

				if err != nil {
					t.Fatalf("expected success, got error %v", err)
				}
			})
		}
	})

	t.Run("skips when previous workspace is per-monitor scratchpad", func(t *testing.T) {
		for _, scratchpadWorkspace := range scratchpadWorkspaceNames() {
			t.Run(scratchpadWorkspace, func(t *testing.T) {
				cleanupMarkerFile(t)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockClient := testutils.NewMockAeroSpaceWM(ctrl)

				rootCmd := cmd.RootCmd(mockClient)
				_, err := testutils.CmdExecute(
					rootCmd,
					"hook",
					"pull-window",
					scratchpadWorkspace,
					scratchpadWorkspace,
				)

				if err != nil {
					t.Fatalf("expected success, got error %v", err)
				}
			})
		}
	})
}
