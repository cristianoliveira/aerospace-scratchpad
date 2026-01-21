package aerospace_test

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/focus"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-ipc/pkg/client"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	client_mock "github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/client"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

//nolint:gochecknoinits // init function is used to set up logger for all tests in this package
func init() {
	// Silence logger for tests
	logger.SetDefaultLogger(&logger.EmptyLogger{})
}

//nolint:gocyclo,gocognit // Test function aggregates multiple test scenarios for readability
func TestAeroSpaceQuerier(t *testing.T) {
	t.Run("IsWindowInWorkspace true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspace := "ws1"
		windowsList := []windows.Window{{WindowID: 1}, {WindowID: 2}}

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		mockClient.GetWindowsMock().EXPECT().
			GetAllWindowsByWorkspace(workspace).
			Return(windowsList, nil).
			Times(1)

		q := aerospace.NewAerospaceQuerier(mockClient)
		in, err := q.IsWindowInWorkspace(2, workspace)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if !in {
			t.Fatalf("expected true, got false")
		}
	})

	t.Run("IsWindowInWorkspace false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspace := "ws1"
		windowsList := []windows.Window{{WindowID: 1}, {WindowID: 2}}

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		mockClient.GetWindowsMock().EXPECT().
			GetAllWindowsByWorkspace(workspace).
			Return(windowsList, nil).
			Times(1)
		q := aerospace.NewAerospaceQuerier(mockClient)
		in, err := q.IsWindowInWorkspace(3, workspace)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if in {
			t.Fatalf("expected false, got true")
		}
	})

	t.Run("IsWindowInFocusedWorkspace", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ws := &workspaces.Workspace{Workspace: "wsX"}
		windowsList := []windows.Window{{WindowID: 5}}

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			mockClient.GetWorkspacesMock().EXPECT().
				GetFocusedWorkspace().
				Return(ws, nil).
				Times(1),
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(ws.Workspace).
				Return(windowsList, nil).
				Times(1),
		)

		q := aerospace.NewAerospaceQuerier(mockClient)
		in, err := q.IsWindowInFocusedWorkspace(5)
		if err != nil || !in {
			t.Fatalf("expected true, got %v err=%v", in, err)
		}
	})

	t.Run("IsWindowFocused true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		focused := &windows.Window{WindowID: 10}
		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		mockClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(focused, nil).
			Times(1)
		q := aerospace.NewAerospaceQuerier(mockClient)
		is, err := q.IsWindowFocused(10)
		if err != nil || !is {
			t.Fatalf("expected true, got %v err=%v", is, err)
		}
	})

	t.Run("IsWindowFocused false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		focused := &windows.Window{WindowID: 10}
		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		mockClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(focused, nil).
			Times(1)
		q := aerospace.NewAerospaceQuerier(mockClient)
		is, err := q.IsWindowFocused(11)
		if err != nil || is {
			t.Fatalf("expected false, got %v err=%v", is, err)
		}
	})

	t.Run("GetNextScratchpadWindow returns first window", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		spWin := []windows.Window{{WindowID: 77}}
		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return([]windows.Window{}, nil).
				Times(1),
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(".scratchpad").
				Return(spWin, nil).
				Times(1),
		)
		q := aerospace.NewAerospaceQuerier(mockClient)
		w, err := q.GetNextScratchpadWindow()
		if err != nil || w == nil || w.WindowID != 77 {
			t.Fatalf("expected 77, got %v err=%v", w, err)
		}
	})

	t.Run(
		"GetNextScratchpadWindow returns error when empty",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := testutils.NewMockAeroSpaceWM(ctrl)
			gomock.InOrder(
				mockClient.GetWindowsMock().EXPECT().
					GetAllWindows().
					Return([]windows.Window{}, nil).
					Times(1),
				mockClient.GetWindowsMock().EXPECT().
					GetAllWindowsByWorkspace(".scratchpad").
					Return([]windows.Window{}, nil).
					Times(1),
			)
			q := aerospace.NewAerospaceQuerier(mockClient)
			if _, err := q.GetNextScratchpadWindow(); err == nil {
				t.Fatalf("expected error when no scratchpad windows")
			}
		},
	)

	t.Run("GetScratchpadWindows returns windows from scratchpad workspace", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		allWindows := []windows.Window{
			{WindowID: 1, WindowLayout: "tiling", Workspace: "ws1"},
			{WindowID: 2, WindowLayout: "tiling", Workspace: "ws1"},
		}
		scratchpadWindows := []windows.Window{
			{WindowID: 3, WindowLayout: "floating", Workspace: ".scratchpad"},
		}

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(".scratchpad").
				Return(scratchpadWindows, nil).
				Times(1),
		)

		q := aerospace.NewAerospaceQuerier(mockClient)
		wins, err := q.GetScratchpadWindows()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wins) != 1 || wins[0].WindowID != 3 {
			t.Fatalf("expected 1 scratchpad window with ID 3, got %d windows", len(wins))
		}
	})

	t.Run("GetScratchpadWindows returns floating windows", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		allWindows := []windows.Window{
			{WindowID: 1, WindowLayout: "tiling", Workspace: "ws1"},
			{WindowID: 2, WindowLayout: "floating", Workspace: "ws1"},
		}

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(".scratchpad").
				Return([]windows.Window{}, nil).
				Times(1),
		)

		q := aerospace.NewAerospaceQuerier(mockClient)
		wins, err := q.GetScratchpadWindows()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wins) != 1 || wins[0].WindowID != 2 {
			t.Fatalf("expected 1 floating window with ID 2, got %d windows", len(wins))
		}
	})

	t.Run("GetScratchpadWindows avoids duplicates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		allWindows := []windows.Window{
			{WindowID: 1, WindowLayout: "floating", Workspace: ".scratchpad"},
		}
		scratchpadWindows := []windows.Window{
			{WindowID: 1, WindowLayout: "floating", Workspace: ".scratchpad"},
		}

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		gomock.InOrder(
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(".scratchpad").
				Return(scratchpadWindows, nil).
				Times(1),
		)

		q := aerospace.NewAerospaceQuerier(mockClient)
		wins, err := q.GetScratchpadWindows()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wins) != 1 {
			t.Fatalf("expected 1 window (no duplicates), got %d windows", len(wins))
		}
	})

	t.Run("GetScratchpadWindows collects per monitor scratchpads", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		mockClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
			{Workspace: ".scratchpad", MonitorID: 1},
			{Workspace: ".scratchpad.2", MonitorID: 2},
			{Workspace: "work", MonitorID: 2},
		})

		allWindows := []windows.Window{
			{WindowID: 1, WindowLayout: "floating", Workspace: ".scratchpad"},
			{WindowID: 2, WindowLayout: "tiling", Workspace: ".scratchpad.2"},
			{WindowID: 3, WindowLayout: "floating", Workspace: "ws1"},
		}

		gomock.InOrder(
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1),
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(".scratchpad").
				Return(
					[]windows.Window{
						{WindowID: 1, WindowLayout: "floating", Workspace: ".scratchpad"},
					},
					nil,
				).
				Times(1),
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(".scratchpad.2").
				Return(
					[]windows.Window{
						{WindowID: 2, WindowLayout: "tiling", Workspace: ".scratchpad.2"},
					},
					nil,
				).
				Times(1),
		)

		q := aerospace.NewAerospaceQuerier(mockClient)
		wins, err := q.GetScratchpadWindows()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(wins) != 3 {
			t.Fatalf("expected 3 scratchpad windows, got %d", len(wins))
		}

		ids := make(map[int]struct{})
		for _, win := range wins {
			ids[win.WindowID] = struct{}{}
		}

		if _, ok := ids[1]; !ok {
			t.Fatalf("expected window 1 from monitor 1 scratchpad")
		}
		if _, ok := ids[2]; !ok {
			t.Fatalf("expected window 2 from monitor 2 scratchpad")
		}
		if _, ok := ids[3]; !ok {
			t.Fatalf("expected floating window 3 to be included")
		}
	})

	t.Run(
		"GetFilteredWindows returns two matches with pattern only",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{{
				Windows: []windows.Window{
					{
						AppName:     "Finder1",
						WindowID:    1,
						WindowTitle: "Finder - foo",
						AppBundleID: "com.apple.finder",
					},
					{
						AppName:     "Finder2",
						WindowID:    2,
						WindowTitle: "Finder2 - bar",
						AppBundleID: "com.apple.finder",
					},
					{
						AppName:     "Terminal",
						WindowID:    3,
						WindowTitle: "Terminal",
						AppBundleID: "com.apple.terminal",
					},
				},
				Workspace: &workspaces.Workspace{Workspace: "ws1"},
			}}
			all := testutils.ExtractAllWindows(tree)

			mockClient := testutils.NewMockAeroSpaceWM(ctrl)
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(all, nil).
				Times(1)
			q := aerospace.NewAerospaceQuerier(mockClient)
			wins, err := q.GetFilteredWindows("Finder", nil)
			if err != nil || len(wins) != 2 {
				t.Fatalf(
					"expected 2 finder windows, got %d err=%v",
					len(wins),
					err,
				)
			}
		},
	)

	t.Run("GetFilteredWindows with filters narrows to one", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{{
			Windows: []windows.Window{
				{
					AppName:     "Finder1",
					WindowID:    1,
					WindowTitle: "Finder - foo",
					AppBundleID: "com.apple.finder",
				},
				{
					AppName:     "Finder2",
					WindowID:    2,
					WindowTitle: "Finder2 - bar",
					AppBundleID: "com.apple.finder",
				},
				{
					AppName:     "Terminal",
					WindowID:    3,
					WindowTitle: "Terminal",
					AppBundleID: "com.apple.terminal",
				},
			},
			Workspace: &workspaces.Workspace{Workspace: "ws1"},
		}}
		all := testutils.ExtractAllWindows(tree)

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		mockClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			Return(all, nil).
			Times(1)
		q := aerospace.NewAerospaceQuerier(mockClient)
		wins, err := q.GetFilteredWindows(
			"Finder",
			[]string{"window-title=foo", "app-bundle-id=apple"},
		)
		if err != nil || len(wins) != 1 || wins[0].WindowID != 1 {
			t.Fatalf("expected 1 window (id=1), got %v err=%v", wins, err)
		}
	})

	t.Run("GetFilteredWindows invalid regex returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := testutils.NewMockAeroSpaceWM(ctrl)
		q := aerospace.NewAerospaceQuerier(mockClient)
		if _, err := q.GetFilteredWindows("[invalid", nil); err == nil {
			t.Fatalf("expected invalid pattern error")
		}
	})

	t.Run(
		"GetFilteredWindows unknown filter property returns error",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{{}}
			all := testutils.ExtractAllWindows(tree)

			mockClient := testutils.NewMockAeroSpaceWM(ctrl)
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(all, nil).
				Times(1)
			q := aerospace.NewAerospaceQuerier(mockClient)
			if _, err := q.GetFilteredWindows("Finder", []string{"unknown=foo"}); err == nil {
				t.Fatalf("expected unknown property error")
			}
		},
	)

	t.Run(
		"GetFilteredWindows with pattern only and no matches returns error",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{{
				Windows: []windows.Window{
					{
						AppName:     "Terminal",
						WindowID:    3,
						WindowTitle: "Terminal",
						AppBundleID: "com.apple.terminal",
					},
				},
				Workspace: &workspaces.Workspace{Workspace: "ws1"},
			}}
			all := testutils.ExtractAllWindows(tree)

			mockClient := testutils.NewMockAeroSpaceWM(ctrl)
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(all, nil).
				Times(1)
			q := aerospace.NewAerospaceQuerier(mockClient)
			if _, err := q.GetFilteredWindows("Finder", nil); err == nil {
				t.Fatalf("expected no match error")
			}
		},
	)

	t.Run(
		"GetFilteredWindows returns error when GetAllWindows fails",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := testutils.NewMockAeroSpaceWM(ctrl)
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(nil, errors.New("mocked_error")).
				Times(1)
			q := aerospace.NewAerospaceQuerier(mockClient)
			if _, err := q.GetFilteredWindows("Finder", nil); err == nil {
				t.Fatalf("expected get windows error")
			}
		},
	)

	t.Run(
		"GetFilteredWindows with window-id filter matches",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{{
				Windows: []windows.Window{
					{
						AppName:     "Terminal",
						WindowID:    1,
						WindowTitle: "Terminal",
						AppBundleID: "com.apple.terminal",
					},
					{
						AppName:     "Finder",
						WindowID:    2,
						WindowTitle: "Documents",
						AppBundleID: "com.apple.finder",
					},
				},
				Workspace: &workspaces.Workspace{Workspace: "ws1"},
			}}
			all := testutils.ExtractAllWindows(tree)

			mockClient := testutils.NewMockAeroSpaceWM(ctrl)
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(all, nil).
				Times(1)
			q := aerospace.NewAerospaceQuerier(mockClient)
			wins, err := q.GetFilteredWindows(
				"Terminal",
				[]string{"window-id=1"},
			)
			if err != nil || len(wins) != 1 || wins[0].WindowID != 1 {
				t.Fatalf("expected 1 window (id=1), got %v err=%v", wins, err)
			}
		},
	)

	t.Run("ListWorkspacesWithMonitors returns workspace mapping", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			SendCommand(
				"list-workspaces",
				[]string{"--all", "--json", "--format", "%{workspace} %{monitor-id}"},
			).
			Return(&client.Response{
				ExitCode: 0,
				StdOut:   `[{"workspace":"1","monitor-id":1},{"workspace":"dev","monitor-id":2}]`,
			}, nil).
			Times(1)

		result, err := aerospace.ListWorkspacesWithMonitors(
			&mockConnectionAeroSpaceClient{conn: socket},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 ||
			result[0].Workspace != "1" || result[0].MonitorID != 1 ||
			result[1].Workspace != "dev" || result[1].MonitorID != 2 {
			t.Fatalf("unexpected workspace mapping: %+v", result)
		}
	})

	t.Run("ListWorkspacesWithMonitors returns error on non-zero exit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			SendCommand(
				"list-workspaces",
				[]string{"--all", "--json", "--format", "%{workspace} %{monitor-id}"},
			).
			Return(&client.Response{
				ExitCode: 1,
				StdErr:   "boom",
			}, nil).
			Times(1)

		if _, err := aerospace.ListWorkspacesWithMonitors(&mockConnectionAeroSpaceClient{conn: socket}); err == nil {
			t.Fatalf("expected error when command fails")
		}
	})

	t.Run("GetFocusedMonitor returns focused monitor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			SendCommand(
				"list-monitors",
				[]string{"--focused", "--json", "--format", "%{monitor-id} %{monitor-name}"},
			).
			Return(&client.Response{
				ExitCode: 0,
				StdOut:   `[{"monitor-id":2,"monitor-name":"DELL U2720"}]`,
			}, nil).
			Times(1)

		monitor, err := aerospace.GetFocusedMonitor(&mockConnectionAeroSpaceClient{conn: socket})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if monitor.MonitorID != 2 || monitor.MonitorName != "DELL U2720" {
			t.Fatalf("unexpected monitor info: %+v", monitor)
		}
	})

	t.Run("GetFocusedMonitor returns error when none focused", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			SendCommand(
				"list-monitors",
				[]string{"--focused", "--json", "--format", "%{monitor-id} %{monitor-name}"},
			).
			Return(&client.Response{
				ExitCode: 0,
				StdOut:   `[]`,
			}, nil).
			Times(1)

		if _, err := aerospace.GetFocusedMonitor(&mockConnectionAeroSpaceClient{conn: socket}); err == nil {
			t.Fatalf("expected error when no focused monitor is returned")
		}
	})

	t.Run("IsScratchpadWorkspace matches default and per-monitor", func(t *testing.T) {
		cases := map[string]bool{
			".scratchpad":   true,
			".scratchpad.2": true,
			"scratchpad":    false,
			".scratchpad-x": false,
		}

		for workspace, expected := range cases {
			if aerospace.IsScratchpadWorkspace(workspace) != expected {
				t.Fatalf("unexpected match result for %s", workspace)
			}
		}
	})

	t.Run(
		"ScratchpadWorkspaceNameForMonitor keeps compatibility on single monitor",
		func(t *testing.T) {
			if got := aerospace.ScratchpadWorkspaceNameForMonitor(2, 1); got != ".scratchpad" {
				t.Fatalf("expected default scratchpad for single monitor, got %s", got)
			}
		},
	)

	t.Run("ScratchpadWorkspaceNameForMonitor appends monitor on multi-monitor", func(t *testing.T) {
		if got := aerospace.ScratchpadWorkspaceNameForMonitor(3, 2); got != ".scratchpad.3" {
			t.Fatalf("expected per-monitor scratchpad name, got %s", got)
		}
	})

	t.Run("ResolveScratchpadWorkspaceNameForMonitor returns existing mapping", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			SendCommand(
				"list-workspaces",
				[]string{"--all", "--json", "--format", "%{workspace} %{monitor-id}"},
			).
			Return(&client.Response{
				ExitCode: 0,
				StdOut:   `[{"workspace":".scratchpad.2","monitor-id":2},{"workspace":"1","monitor-id":1}]`,
			}, nil).
			Times(1)

		name, err := aerospace.ResolveScratchpadWorkspaceNameForMonitor(
			&mockConnectionAeroSpaceClient{conn: socket},
			2,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != ".scratchpad.2" {
			t.Fatalf("expected existing scratchpad name, got %s", name)
		}
	})

	t.Run("ResolveScratchpadWorkspaceNameForMonitor builds name when missing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			SendCommand(
				"list-workspaces",
				[]string{"--all", "--json", "--format", "%{workspace} %{monitor-id}"},
			).
			Return(&client.Response{
				ExitCode: 0,
				StdOut:   `[{"workspace":"1","monitor-id":1},{"workspace":"2","monitor-id":2}]`,
			}, nil).
			Times(1)

		name, err := aerospace.ResolveScratchpadWorkspaceNameForMonitor(
			&mockConnectionAeroSpaceClient{conn: socket},
			2,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != ".scratchpad.2" {
			t.Fatalf("expected generated scratchpad name, got %s", name)
		}
	})

	t.Run(
		"ResolveScratchpadWorkspaceNameForMonitor skips mismatched workspace name",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			socket := client_mock.NewMockAeroSpaceConnection(ctrl)
			// First call from ResolveScratchpadWorkspaceNameForMonitor
			socket.EXPECT().
				SendCommand(
					"list-workspaces",
					[]string{"--all", "--json", "--format", "%{workspace} %{monitor-id}"},
				).
				Return(&client.Response{
					ExitCode: 0,
					StdOut:   `[{"workspace":".scratchpad.1","monitor-id":2},{"workspace":"1","monitor-id":1}]`,
				}, nil).
				Times(2)
				// Called once by ResolveScratchpadWorkspaceNameForMonitor, once by repairMismatchedScratchpadWorkspace
			socket.EXPECT().
				SendCommand(
					"rename-workspace",
					[]string{".scratchpad.1", ".scratchpad.2"},
				).
				Return(&client.Response{
					ExitCode: 0,
					StdOut:   "",
				}, nil).
				Times(1)

			name, err := aerospace.ResolveScratchpadWorkspaceNameForMonitor(
				&mockConnectionAeroSpaceClient{conn: socket},
				2,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if name != ".scratchpad.2" {
				t.Fatalf("expected generated scratchpad name, got %s", name)
			}
		},
	)

	t.Run("ListScratchpadWorkspaceNames returns detected scratchpads", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			SendCommand(
				"list-workspaces",
				[]string{"--all", "--json", "--format", "%{workspace} %{monitor-id}"},
			).
			Return(&client.Response{
				ExitCode: 0,
				StdOut:   `[{"workspace":".scratchpad","monitor-id":1},{"workspace":".scratchpad.2","monitor-id":2},{"workspace":"work","monitor-id":2}]`,
			}, nil).
			Times(1)

		names, err := aerospace.ListScratchpadWorkspaceNames(
			&mockConnectionAeroSpaceClient{conn: socket},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 2 || names[0] != ".scratchpad" || names[1] != ".scratchpad.2" {
			t.Fatalf("unexpected scratchpad names: %+v", names)
		}
	})

	t.Run("ListScratchpadWorkspaceNames falls back to default when absent", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			SendCommand(
				"list-workspaces",
				[]string{"--all", "--json", "--format", "%{workspace} %{monitor-id}"},
			).
			Return(&client.Response{
				ExitCode: 0,
				StdOut:   `[{"workspace":"1","monitor-id":1}]`,
			}, nil).
			Times(1)

		names, err := aerospace.ListScratchpadWorkspaceNames(
			&mockConnectionAeroSpaceClient{conn: socket},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 1 || names[0] != ".scratchpad" {
			t.Fatalf("expected default scratchpad fallback, got %+v", names)
		}
	})
}

// mockConnectionAeroSpaceClient implements AeroSpaceWMClient by exposing only the raw connection.
type mockConnectionAeroSpaceClient struct {
	conn client.AeroSpaceConnection
}

func (m *mockConnectionAeroSpaceClient) Windows() *windows.Service {
	return nil
}

func (m *mockConnectionAeroSpaceClient) Workspaces() *workspaces.Service {
	return nil
}

func (m *mockConnectionAeroSpaceClient) Focus() *focus.Service {
	return nil
}

func (m *mockConnectionAeroSpaceClient) Layout() *layout.Service {
	return nil
}

func (m *mockConnectionAeroSpaceClient) Connection() client.AeroSpaceConnection {
	return m.conn
}
