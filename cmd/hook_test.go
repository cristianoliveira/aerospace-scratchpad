package cmd_test

import (
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

func cleanupMarkerFile(t *testing.T) {
	t.Helper()

	err := os.Remove(constants.TempScratchpadMovingFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to clean marker file: %v", err)
	}
}

func TestHookPullWindow(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})

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
}
