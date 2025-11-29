package cmd_test

import (
	"os"
	"reflect"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	clientipc "github.com/cristianoliveira/aerospace-ipc/pkg/client"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

type stubConnection struct {
	t               *testing.T
	expectedCommand string
	expectedArgs    []string
	called          bool
}

func (s *stubConnection) SendCommand(command string, args []string) (*clientipc.Response, error) {
	s.t.Helper()

	if command != s.expectedCommand {
		s.t.Fatalf("expected command %q, got %q", s.expectedCommand, command)
	}

	if !reflect.DeepEqual(s.expectedArgs, args) {
		s.t.Fatalf("expected args %v, got %v", s.expectedArgs, args)
	}

	s.called = true
	return &clientipc.Response{ExitCode: 0}, nil
}

func (s *stubConnection) CloseConnection() error         { return nil }
func (s *stubConnection) GetSocketPath() (string, error) { return "", nil }
func (s *stubConnection) GetServerVersion() (string, error) {
	return "", nil
}
func (s *stubConnection) CheckServerVersion() error { return nil }

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
			mockClient.GetWindowsMock().EXPECT().GetFocusedWindow().Return(focusedWindow, nil).Times(1),
			mockClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspace(focusedWindow.WindowID, "prev-ws").
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
