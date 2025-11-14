package aerospace_test //nolint:cyclop // test suite validates many scenarios in one package

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	mock_aerospace "github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

//nolint:gocognit // Integration-style test aggregates several window scenarios for readability
func TestAeroSpaceQuerier(t *testing.T) {
	// Silence logger for tests
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	t.Run("IsWindowInWorkspace true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspace := "ws1"
		windows := []aerospacecli.Window{{WindowID: 1}, {WindowID: 2}}

		client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		client.EXPECT().
			GetAllWindowsByWorkspace(workspace).
			Return(windows, nil).
			Times(1)

		q := aerospace.NewAerospaceQuerier(client)
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
		windows := []aerospacecli.Window{{WindowID: 1}, {WindowID: 2}}

		client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		client.EXPECT().
			GetAllWindowsByWorkspace(workspace).
			Return(windows, nil).
			Times(1)
		q := aerospace.NewAerospaceQuerier(client)
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

		ws := &aerospacecli.Workspace{Workspace: "wsX"}
		windows := []aerospacecli.Window{{WindowID: 5}}

		client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			client.EXPECT().
				GetFocusedWorkspace().
				Return(ws, nil).
				Times(1),
			client.EXPECT().
				GetAllWindowsByWorkspace(ws.Workspace).
				Return(windows, nil).
				Times(1),
		)

		q := aerospace.NewAerospaceQuerier(client)
		in, err := q.IsWindowInFocusedWorkspace(5)
		if err != nil || !in {
			t.Fatalf("expected true, got %v err=%v", in, err)
		}
	})

	t.Run("IsWindowFocused true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		focused := &aerospacecli.Window{WindowID: 10}
		client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		client.EXPECT().
			GetFocusedWindow().
			Return(focused, nil).
			Times(1)
		q := aerospace.NewAerospaceQuerier(client)
		is, err := q.IsWindowFocused(10)
		if err != nil || !is {
			t.Fatalf("expected true, got %v err=%v", is, err)
		}
	})

	t.Run("IsWindowFocused false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		focused := &aerospacecli.Window{WindowID: 10}
		client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		client.EXPECT().
			GetFocusedWindow().
			Return(focused, nil).
			Times(1)
		q := aerospace.NewAerospaceQuerier(client)
		is, err := q.IsWindowFocused(11)
		if err != nil || is {
			t.Fatalf("expected false, got %v err=%v", is, err)
		}
	})

	t.Run("GetNextScratchpadWindow returns first window", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		spWin := []aerospacecli.Window{{WindowID: 77}}
		client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		client.EXPECT().
			GetAllWindowsByWorkspace(".scratchpad").
			Return(spWin, nil).
			Times(1)
		q := aerospace.NewAerospaceQuerier(client)
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

			client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
			client.EXPECT().
				GetAllWindowsByWorkspace(".scratchpad").
				Return([]aerospacecli.Window{}, nil).
				Times(1)
			q := aerospace.NewAerospaceQuerier(client)
			if _, err := q.GetNextScratchpadWindow(); err == nil {
				t.Fatalf("expected error when no scratchpad windows")
			}
		},
	)

	t.Run(
		"GetFilteredWindows returns two matches with pattern only",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{{
				Windows: []aerospacecli.Window{
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
				Workspace: &aerospacecli.Workspace{Workspace: "ws1"},
			}}
			all := testutils.ExtractAllWindows(tree)

			client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
			client.EXPECT().
				GetAllWindows().
				Return(all, nil).
				Times(1)
			q := aerospace.NewAerospaceQuerier(client)
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
			Windows: []aerospacecli.Window{
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
			Workspace: &aerospacecli.Workspace{Workspace: "ws1"},
		}}
		all := testutils.ExtractAllWindows(tree)

		client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		client.EXPECT().
			GetAllWindows().
			Return(all, nil).
			Times(1)
		q := aerospace.NewAerospaceQuerier(client)
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

		client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		q := aerospace.NewAerospaceQuerier(client)
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

			client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
			client.EXPECT().
				GetAllWindows().
				Return(all, nil).
				Times(1)
			q := aerospace.NewAerospaceQuerier(client)
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
				Windows: []aerospacecli.Window{
					{
						AppName:     "Terminal",
						WindowID:    3,
						WindowTitle: "Terminal",
						AppBundleID: "com.apple.terminal",
					},
				},
				Workspace: &aerospacecli.Workspace{Workspace: "ws1"},
			}}
			all := testutils.ExtractAllWindows(tree)

			client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
			client.EXPECT().
				GetAllWindows().
				Return(all, nil).
				Times(1)
			q := aerospace.NewAerospaceQuerier(client)
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

			client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
			client.EXPECT().
				GetAllWindows().
				Return(nil, errors.New("mocked_error")).
				Times(1)
			q := aerospace.NewAerospaceQuerier(client)
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
				Windows: []aerospacecli.Window{
					{
						AppName:     "Terminal",
						WindowID:    1,
						WindowTitle: "Terminal",
						AppBundleID: "com.apple.terminal",
					},
				},
				Workspace: &aerospacecli.Workspace{Workspace: "ws1"},
			}}
			all := testutils.ExtractAllWindows(tree)

			client := mock_aerospace.NewMockAeroSpaceClient(ctrl)
			client.EXPECT().
				GetAllWindows().
				Return(all, nil).
				Times(1)
			q := aerospace.NewAerospaceQuerier(client)
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
