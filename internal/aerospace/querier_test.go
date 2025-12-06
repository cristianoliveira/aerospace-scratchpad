package aerospace_test

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
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
		mockClient.GetWindowsMock().EXPECT().
			GetAllWindowsByWorkspace(".scratchpad").
			Return(spWin, nil).
			Times(1)
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
			mockClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(".scratchpad").
				Return([]windows.Window{}, nil).
				Times(1)
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
}
