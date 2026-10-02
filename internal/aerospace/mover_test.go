package aerospace_test

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	client_mock "github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/client"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

//nolint:gocognit // Test function aggregates multiple mover scenarios for readability
func TestMoverAeroSpaceMoveWindowToScratchpadForMonitor(t *testing.T) {
	t.Run("moves window to scratchpad attached to its monitor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Notepad", WindowID: 111, Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// Focus is on another monitor to prove it does not drive the target.
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 2, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: ".scratchpad", MonitorID: 1},
			{Workspace: "work", MonitorID: 2},
		})

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(
				gomock.Any(),
				gomock.Any(),
			).
			DoAndReturn(func(
				args workspaces.MoveWindowToWorkspaceArgs,
				opts workspaces.MoveWindowToWorkspaceOpts,
			) error {
				if args.WorkspaceName != ".scratchpad" {
					t.Errorf("expected target .scratchpad, got %s", args.WorkspaceName)
				}
				if opts.WindowID == nil || *opts.WindowID != window.WindowID {
					t.Errorf("expected window %d to be moved", window.WindowID)
				}
				return nil
			}).
			Times(1)

		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(
				[]string{"floating"},
				layout.SetLayoutOpts{WindowID: &window.WindowID},
			).
			Return(nil).
			Times(1)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		target, err := mover.MoveWindowToScratchpadForMonitor(window, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target != ".scratchpad" {
			t.Fatalf("expected target .scratchpad, got %s", target)
		}
	})

	t.Run("legacy mover routes by window source monitor when another is focused", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Notepad", WindowID: 112, Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 2, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
			{Workspace: ".scratchpad.1", MonitorID: 1},
			{Workspace: ".scratchpad.2", MonitorID: 2},
		})

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(
				workspaces.MoveWindowToWorkspaceArgs{WorkspaceName: ".scratchpad.1"},
				workspaces.MoveWindowToWorkspaceOpts{WindowID: &window.WindowID},
			).
			Return(nil).
			Times(1)
		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(
				[]string{"floating"},
				layout.SetLayoutOpts{WindowID: &window.WindowID},
			).
			Return(nil).
			Times(1)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		target, err := mover.MoveWindowToScratchpad(window)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target != ".scratchpad.1" {
			t.Fatalf("expected source monitor scratchpad .scratchpad.1, got %s", target)
		}
	})

	t.Run("fails closed when target scratchpad is attached to another monitor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Notepad", WindowID: 222, Workspace: "ws2"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 2, MonitorName: "HDMI"})
		// Stale mapping: .scratchpad.2 lives on monitor 1 while the window is on monitor 2.
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws2", MonitorID: 2},
			{Workspace: ".scratchpad.2", MonitorID: 1},
		})

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(0)
		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(0)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		_, err := mover.MoveWindowToScratchpadForMonitor(window, 2)
		if err == nil {
			t.Fatalf("expected error for cross-monitor scratchpad target")
		}
	})

	t.Run("fails closed when absent target in multi-monitor setup", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Notepad", WindowID: 333, Workspace: "ws2"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// No scratchpad exists yet; AeroSpace would create it on the main
		// monitor, which may not be the window's monitor. Focus is deliberately
		// unset to prove it is not consulted for the decision.
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
		})

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(0)
		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(0)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		_, err := mover.MoveWindowToScratchpadForMonitor(window, 2)
		if err == nil {
			t.Fatalf("expected error when new scratchpad cannot be safely provisioned")
		}
	})

	t.Run("fails closed when absent target and focused monitor match source", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Notepad", WindowID: 334, Workspace: "ws2"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// The focus/source match does not prove where an absent workspace will
		// be created: AeroSpace may place it on the main monitor instead.
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 2, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
		})

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(0)
		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(0)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		if _, err := mover.MoveWindowToScratchpadForMonitor(window, 2); err == nil {
			t.Fatal("expected error rather than trusting focused monitor for absent target")
		}
	})

	t.Run("allows absent target on proven single-monitor setup", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Notepad", WindowID: 444, Workspace: "ws2"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// Single monitor: there is no other monitor to cross, so provisioning
		// the default scratchpad is safe.
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws2", MonitorID: 2},
		})

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(
				gomock.Any(),
				gomock.Any(),
			).
			DoAndReturn(func(
				args workspaces.MoveWindowToWorkspaceArgs,
				opts workspaces.MoveWindowToWorkspaceOpts,
			) error {
				if args.WorkspaceName != ".scratchpad" {
					t.Errorf("expected target .scratchpad, got %s", args.WorkspaceName)
				}
				if opts.WindowID == nil || *opts.WindowID != window.WindowID {
					t.Errorf("expected window %d to be moved", window.WindowID)
				}
				return nil
			}).
			Times(1)

		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(
				[]string{"floating"},
				layout.SetLayoutOpts{WindowID: &window.WindowID},
			).
			Return(nil).
			Times(1)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		target, err := mover.MoveWindowToScratchpadForMonitor(window, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target != ".scratchpad" {
			t.Fatalf("expected target .scratchpad, got %s", target)
		}
	})

	t.Run("fails closed when workspace query fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			SendCommand(
				"list-workspaces",
				[]string{"--all", "--json", "--format", "%{workspace} %{monitor-id}"},
			).
			Return(nil, errors.New("socket closed")).
			Times(1)

		window := windows.Window{AppName: "Notepad", WindowID: 666, Workspace: "ws1"}
		mover := aerospace.NewAeroSpaceMover(
			&mockConnectionAeroSpaceClient{conn: socket},
		)
		if _, err := mover.MoveWindowToScratchpadForMonitor(window, 1); err == nil {
			t.Fatalf("expected error when workspace query fails")
		}
	})

	t.Run("fails closed for unknown source monitor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(0)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		window := windows.Window{AppName: "Notepad", WindowID: 555, Workspace: "ws1"}
		if _, err := mover.MoveWindowToScratchpadForMonitor(window, 0); err == nil {
			t.Fatalf("expected error for unknown source monitor")
		}
	})
}
