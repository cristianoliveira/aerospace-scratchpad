package aerospace_test

import (
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	client_mock "github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/client"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-ipc/pkg/client"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

//nolint:gocognit,gocyclo // Test function aggregates multiple mover scenarios for readability
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

	t.Run("legacy mover routes by source monitor, ignoring focus", func(t *testing.T) {
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

	t.Run("non-zero workspace mapping response prevents summon and move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Spotify", WindowID: 667, Workspace: "ws2"}
		prevFocused := windows.Window{AppName: "Finder", WindowID: 999, Workspace: "ws2"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.SetWorkspaceListResponse(client.Response{
			ExitCode: 1,
			StdOut:   "[]",
			StdErr:   "mapping unavailable",
		})
		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(&prevFocused, nil).
			AnyTimes()
		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "ws2"}, nil).
			AnyTimes()
		aerospaceClient.GetFocusMock().EXPECT().
			SetFocusByWindowID(prevFocused.WindowID).
			Return(nil).
			AnyTimes()

		moveCalls := 0
		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				workspaces.MoveWindowToWorkspaceArgs,
				workspaces.MoveWindowToWorkspaceOpts,
			) error {
				moveCalls++
				return nil
			}).
			AnyTimes()

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		_, err := mover.MoveWindowToScratchpadForMonitor(window, 2)
		if err == nil || !strings.Contains(err.Error(), "mapping unavailable") {
			t.Fatalf("expected workspace mapping failure, got %v", err)
		}
		if summoned := aerospaceClient.GetSummonedWorkspaces(); len(summoned) != 0 {
			t.Fatalf("expected no summon after mapping failure, got %v", summoned)
		}
		if moveCalls != 0 {
			t.Fatalf("expected no move after mapping failure, got %d calls", moveCalls)
		}
	})

	t.Run("provisions absent scratchpad on source monitor when focus differs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Spotify", WindowID: 111, Workspace: "ws2"}
		prevFocused := windows.Window{AppName: "Finder", WindowID: 999, Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		// User focus is on monitor 1; Spotify lives on monitor 2 workspace ws2
		// with no scratchpad provisioned yet.
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 1, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
		})

		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(&prevFocused, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			DoAndReturn(func() ([]windows.Window, error) {
				return []windows.Window{window, prevFocused}, nil
			}).
			AnyTimes()

		// Focus monitor 2, capture its active workspace for restoration.
		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "ws2"}, nil).
			Times(1)

		// Restore the user's focused window afterwards.
		aerospaceClient.GetFocusMock().EXPECT().
			SetFocusByWindowID(999).
			Return(nil).
			Times(1)

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(
				gomock.Any(),
				gomock.Any(),
			).
			DoAndReturn(func(
				args workspaces.MoveWindowToWorkspaceArgs,
				opts workspaces.MoveWindowToWorkspaceOpts,
			) error {
				if args.WorkspaceName != ".scratchpad.2" {
					t.Errorf("expected target .scratchpad.2, got %s", args.WorkspaceName)
				}
				if opts.WindowID == nil || *opts.WindowID != window.WindowID {
					t.Errorf("expected window %d to be moved", window.WindowID)
				}
				// Order guard: summon must precede the move; the
				// active-workspace restore must follow it.
				if summoned := aerospaceClient.GetSummonedWorkspaces(); len(summoned) != 1 {
					t.Errorf("expected summon before move, got %v", summoned)
				}
				if switched := aerospaceClient.GetWorkspaceSwitchCalls(); len(switched) != 0 {
					t.Errorf("expected workspace restore after move, got %v", switched)
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
		if target != ".scratchpad.2" {
			t.Fatalf("expected target .scratchpad.2, got %s", target)
		}

		summoned := aerospaceClient.GetSummonedWorkspaces()
		if len(summoned) != 1 || summoned[0] != ".scratchpad.2" {
			t.Fatalf("expected summon of .scratchpad.2, got %v", summoned)
		}
		if got := aerospaceClient.GetFocusMonitorCalls(); got != 1 {
			t.Fatalf("expected focus-monitor on source monitor, got %d calls", got)
		}
		if switched := aerospaceClient.GetWorkspaceSwitchCalls(); len(switched) != 1 ||
			switched[0] != "ws2" {
			t.Fatalf("expected workspace restore to ws2, got %v", switched)
		}
	})

	t.Run("provisions absent scratchpad when focus already on source monitor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Spotify", WindowID: 222, Workspace: "ws2"}
		prevFocused := windows.Window{AppName: "Finder", WindowID: 999, Workspace: "ws2"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 2, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
		})

		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(&prevFocused, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			DoAndReturn(func() ([]windows.Window, error) {
				return []windows.Window{window, prevFocused}, nil
			}).
			AnyTimes()

		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "ws2"}, nil).
			Times(1)
		aerospaceClient.GetFocusMock().EXPECT().
			SetFocusByWindowID(999).
			Return(nil).
			Times(1)

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)
		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		if _, err := mover.MoveWindowToScratchpadForMonitor(window, 2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if switched := aerospaceClient.GetWorkspaceSwitchCalls(); len(switched) != 1 ||
			switched[0] != "ws2" {
			t.Fatalf("expected workspace restore to ws2, got %v", switched)
		}
	})

	t.Run("does not summon for stale existing target", func(t *testing.T) {
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
		if _, err := mover.MoveWindowToScratchpadForMonitor(window, 2); err == nil {
			t.Fatalf("expected error for cross-monitor scratchpad target")
		}
		if summoned := aerospaceClient.GetSummonedWorkspaces(); len(summoned) != 0 {
			t.Fatalf("expected no summon for existing cross-attached target, got %v", summoned)
		}
	})

	t.Run("fails closed and restores state when summon fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Spotify", WindowID: 333, Workspace: "ws2"}
		prevFocused := windows.Window{AppName: "Finder", WindowID: 999, Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 1, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
		})
		aerospaceClient.SetSummonWorkspaceError(errors.New("summon rejected"))

		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(&prevFocused, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			DoAndReturn(func() ([]windows.Window, error) {
				return []windows.Window{window, prevFocused}, nil
			}).
			AnyTimes()

		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "ws2"}, nil).
			Times(1)
		// Focus must be restored even on failure.
		aerospaceClient.GetFocusMock().EXPECT().
			SetFocusByWindowID(999).
			Return(nil).
			Times(1)

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
			t.Fatalf("expected error when summon fails")
		}
		if switched := aerospaceClient.GetWorkspaceSwitchCalls(); len(switched) != 1 ||
			switched[0] != "ws2" {
			t.Fatalf("expected source workspace restore on failure, got %v", switched)
		}
	})

	t.Run("fails closed when provisioned placement is wrong", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Spotify", WindowID: 444, Workspace: "ws2"}
		prevFocused := windows.Window{AppName: "Finder", WindowID: 999, Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 1, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
		})
		// Simulate a topology race: the summoned workspace lands on monitor 1.
		aerospaceClient.SetSummonPlacementMonitor(1)

		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(&prevFocused, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			DoAndReturn(func() ([]windows.Window, error) {
				return []windows.Window{window, prevFocused}, nil
			}).
			AnyTimes()

		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "ws2"}, nil).
			Times(1)
		aerospaceClient.GetFocusMock().EXPECT().
			SetFocusByWindowID(999).
			Return(nil).
			Times(1)

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
			t.Fatalf("expected error when provisioned workspace lands on another monitor")
		}
	})

	t.Run("skips user-focus restore when the moved window was focused", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Spotify", WindowID: 555, Workspace: "ws2"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 2, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
		})

		// The moved window itself was focused before the operation.
		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(&window, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			DoAndReturn(func() ([]windows.Window, error) {
				return []windows.Window{window}, nil
			}).
			AnyTimes()

		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "ws2"}, nil).
			Times(1)

		// The moved window must NOT be re-focused inside the scratchpad.
		aerospaceClient.GetFocusMock().EXPECT().
			SetFocusByWindowID(555).
			Return(nil).
			Times(0)

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)
		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		if _, err := mover.MoveWindowToScratchpadForMonitor(window, 2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if switched := aerospaceClient.GetWorkspaceSwitchCalls(); len(switched) != 1 ||
			switched[0] != "ws2" {
			t.Fatalf("expected workspace restore to ws2, got %v", switched)
		}
	})

	t.Run("restores state when the move fails after provisioning", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Spotify", WindowID: 666, Workspace: "ws2"}
		prevFocused := windows.Window{AppName: "Finder", WindowID: 999, Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 1, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
		})

		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(&prevFocused, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			DoAndReturn(func() ([]windows.Window, error) {
				return []windows.Window{window, prevFocused}, nil
			}).
			AnyTimes()

		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "ws2"}, nil).
			Times(1)
		// Cleanup must run even though the move failed.
		aerospaceClient.GetFocusMock().EXPECT().
			SetFocusByWindowID(999).
			Return(nil).
			Times(1)

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
			Return(errors.New("move rejected")).
			Times(1)
		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(0)

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		if _, err := mover.MoveWindowToScratchpadForMonitor(window, 2); err == nil {
			t.Fatalf("expected move error to surface")
		}
		if got := aerospaceClient.GetWorkspaceSwitchCalls(); len(got) != 1 {
			t.Fatalf("expected source workspace restore on move failure, got %d calls", len(got))
		}
	})

	t.Run("surfaces restoration failure after a successful provisioned move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		window := windows.Window{AppName: "Spotify", WindowID: 777, Workspace: "ws2"}
		prevFocused := windows.Window{AppName: "Finder", WindowID: 999, Workspace: "ws1"}

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 2, MonitorName: "HDMI"})
		aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: "ws1", MonitorID: 1},
			{Workspace: "ws2", MonitorID: 2},
		})

		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(&prevFocused, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			DoAndReturn(func() ([]windows.Window, error) {
				return []windows.Window{window, prevFocused}, nil
			}).
			AnyTimes()

		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "ws2"}, nil).
			Times(1)
		// User focus restore is still attempted when the active-workspace
		// restore fails.
		aerospaceClient.GetFocusMock().EXPECT().
			SetFocusByWindowID(999).
			Return(nil).
			Times(1)

		aerospaceClient.GetWorkspacesMock().EXPECT().
			MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)
		aerospaceClient.GetLayoutMock().EXPECT().
			SetLayout(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		// Inject the workspace restore failure through the harness.
		aerospaceClient.SetWorkspaceSwitchError(errors.New("restore rejected"))

		mover := aerospace.NewAeroSpaceMover(aerospaceClient)
		_, err := mover.MoveWindowToScratchpadForMonitor(window, 2)
		if err == nil {
			t.Fatalf("expected restoration failure to surface")
		}
		if !strings.Contains(err.Error(), "focus restoration failed") {
			t.Fatalf("expected actionable restoration error, got %v", err)
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
