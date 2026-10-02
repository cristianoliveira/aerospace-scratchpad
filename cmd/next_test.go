package cmd_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/focus"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

//nolint:gocognit // Test function aggregates command scenarios and dynamic state transitions.
func TestNextCmd(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)

	t.Run("summons first scratchpad window when no window is focused", func(t *testing.T) {
		command := "next"
		args := []string{command}

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
			{
				Windows: []windows.Window{
					{
						AppName:  "Scratchpad Window",
						WindowID: 9999,
					},
					{
						AppName:  "Another Scratchpad Window",
						WindowID: 8888,
					},
				},
				Workspace: &workspaces.Workspace{
					Workspace: constants.DefaultScratchpadWorkspaceName,
				},

				FocusedWindowID: 0,
			},
		}

		focusedTree := testutils.ExtractFocusedTree(tree)
		scratchpadWindows := testutils.ExtractScratchpadWindows(tree)

		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		windowID := 8888
		gomock.InOrder(
			aerospaceClient.GetWorkspacesMock().EXPECT().
				GetFocusedWorkspace().
				Return(focusedTree.Workspace, nil).
				Times(1),
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindows().
				Return(testutils.ExtractAllWindows(tree), nil).
				Times(1),
			aerospaceClient.GetWindowsMock().EXPECT().
				GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
				Return(scratchpadWindows.Windows, nil).
				Times(1),
			aerospaceClient.GetWindowsMock().EXPECT().
				GetFocusedWindow().
				Return(nil, errors.New("no windows focused found")).
				Times(1),
			aerospaceClient.GetWorkspacesMock().EXPECT().
				MoveWindowToWorkspaceWithOpts(
					workspaces.MoveWindowToWorkspaceArgs{
						WorkspaceName: focusedTree.Workspace.Workspace,
					},
					workspaces.MoveWindowToWorkspaceOpts{
						WindowID: &windowID,
					},
				).
				Return(nil).
				Times(1),
			aerospaceClient.GetFocusMock().EXPECT().
				SetFocusByWindowID(8888). // Focus the moved window
				Return(nil).
				Times(1),
		)

		wrappedClient := aerospace.NewAeroSpaceClient(aerospaceClient)
		_ = wrappedClient
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if out == "" {
			t.Errorf("Expected output, got empty string")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
		testutils.MatchSnapshot(t, tree, cmdAsString, out, err)
	})

	t.Run("successive calls cycle from the focused window", func(t *testing.T) {
		tests := []struct {
			name              string
			args              []string
			monitorID         int
			workspaceMonitors []aerospace.WorkspaceMonitor
			windows           []windows.Window
			wantWindowIDs     []int
		}{
			{
				name:      "single monitor",
				args:      []string{"next"},
				monitorID: 1,
				workspaceMonitors: []aerospace.WorkspaceMonitor{
					{Workspace: ".scratchpad", MonitorID: 1},
					{Workspace: "work", MonitorID: 1},
				},
				windows: []windows.Window{
					{WindowID: 300, WindowLayout: "floating", Workspace: ".scratchpad"},
					{WindowID: 100, WindowLayout: "floating", Workspace: ".scratchpad"},
					{WindowID: 200, WindowLayout: "floating", Workspace: ".scratchpad"},
				},
				wantWindowIDs: []int{100, 200, 300, 100},
			},
			{
				name:      "explicit current monitor",
				args:      []string{"next", "--monitor", "current"},
				monitorID: 2,
				workspaceMonitors: []aerospace.WorkspaceMonitor{
					{Workspace: ".scratchpad.1", MonitorID: 1},
					{Workspace: ".scratchpad.2", MonitorID: 2},
					{Workspace: "work", MonitorID: 2},
				},
				windows: []windows.Window{
					{WindowID: 300, WindowLayout: "floating", Workspace: ".scratchpad.2"},
					{WindowID: 100, WindowLayout: "floating", Workspace: ".scratchpad.1"},
					{WindowID: 200, WindowLayout: "floating", Workspace: ".scratchpad.2"},
				},
				wantWindowIDs: []int{200, 300, 200},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
				aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{
					MonitorID: test.monitorID,
				})
				aerospaceClient.SetWorkspaceMonitors(test.workspaceMonitors)

				currentWindows := slices.Clone(test.windows)
				focusedWindowID := 999
				var movedWindowIDs []int

				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(&workspaces.Workspace{Workspace: "work"}, nil).
					Times(len(test.wantWindowIDs))
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindows().
					DoAndReturn(func() ([]windows.Window, error) {
						return slices.Clone(currentWindows), nil
					}).
					Times(len(test.wantWindowIDs))
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindowsByWorkspace(gomock.Any()).
					DoAndReturn(func(workspace string) ([]windows.Window, error) {
						var matching []windows.Window
						for _, window := range currentWindows {
							if window.Workspace == workspace {
								matching = append(matching, window)
							}
						}
						return matching, nil
					}).
					AnyTimes()
				aerospaceClient.GetWindowsMock().EXPECT().
					GetFocusedWindow().
					DoAndReturn(func() (*windows.Window, error) {
						return &windows.Window{WindowID: focusedWindowID}, nil
					}).
					Times(len(test.wantWindowIDs))
				aerospaceClient.GetWorkspacesMock().EXPECT().
					MoveWindowToWorkspaceWithOpts(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						args workspaces.MoveWindowToWorkspaceArgs,
						opts workspaces.MoveWindowToWorkspaceOpts,
					) error {
						if opts.WindowID == nil {
							t.Fatal("expected window ID for move")
						}
						movedWindowIDs = append(movedWindowIDs, *opts.WindowID)
						for index := range currentWindows {
							if currentWindows[index].WindowID == *opts.WindowID {
								currentWindows[index].Workspace = args.WorkspaceName
							}
						}
						return nil
					}).
					Times(len(test.wantWindowIDs))
				aerospaceClient.GetFocusMock().EXPECT().
					SetFocusByWindowID(gomock.Any()).
					DoAndReturn(func(windowID int, _ ...focus.SetFocusOpts) error {
						focusedWindowID = windowID
						return nil
					}).
					Times(len(test.wantWindowIDs))

				for range test.wantWindowIDs {
					root := cmd.RootCmd(aerospaceClient)
					if _, err := testutils.CmdExecute(root, test.args...); err != nil {
						t.Fatalf("next command failed: %v", err)
					}
				}

				if !slices.Equal(movedWindowIDs, test.wantWindowIDs) {
					t.Fatalf("moved windows %v, want %v", movedWindowIDs, test.wantWindowIDs)
				}
			})
		}
	})

	t.Run("defaults to all monitors and scopes explicit monitor flags", func(t *testing.T) {
		tests := []struct {
			name         string
			args         []string
			wantWindowID int
		}{
			{
				name:         "all monitors by default",
				args:         []string{"next"},
				wantWindowID: 100,
			},
			{
				name:         "explicit current monitor",
				args:         []string{"next", "--monitor", "current"},
				wantWindowID: 200,
			},
			{
				name:         "explicit monitor ID",
				args:         []string{"next", "--monitor", "2"},
				wantWindowID: 100,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
				aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 1})
				aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
					{Workspace: ".scratchpad.1", MonitorID: 1},
					{Workspace: ".scratchpad.2", MonitorID: 2},
					{Workspace: "work", MonitorID: 1},
				})
				scratchpadWindows := []windows.Window{
					{WindowID: 100, WindowLayout: "floating", Workspace: ".scratchpad.2"},
					{WindowID: 200, WindowLayout: "floating", Workspace: ".scratchpad.1"},
				}

				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(&workspaces.Workspace{Workspace: "work"}, nil).
					Times(1)
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindows().
					Return(scratchpadWindows, nil).
					Times(1)
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindowsByWorkspace(gomock.Any()).
					DoAndReturn(func(workspace string) ([]windows.Window, error) {
						var matching []windows.Window
						for _, window := range scratchpadWindows {
							if window.Workspace == workspace {
								matching = append(matching, window)
							}
						}
						return matching, nil
					}).
					AnyTimes()
				aerospaceClient.GetWindowsMock().EXPECT().
					GetFocusedWindow().
					Return(&windows.Window{WindowID: 999}, nil).
					Times(1)
				aerospaceClient.GetWorkspacesMock().EXPECT().
					MoveWindowToWorkspaceWithOpts(
						workspaces.MoveWindowToWorkspaceArgs{WorkspaceName: "work"},
						gomock.Any(),
					).
					DoAndReturn(func(
						_ workspaces.MoveWindowToWorkspaceArgs,
						opts workspaces.MoveWindowToWorkspaceOpts,
					) error {
						if opts.WindowID == nil {
							t.Fatal("expected window ID for move")
						}
						if *opts.WindowID != test.wantWindowID {
							t.Fatalf("moved window %d, want %d", *opts.WindowID, test.wantWindowID)
						}
						return nil
					}).
					Times(1)
				aerospaceClient.GetFocusMock().EXPECT().
					SetFocusByWindowID(test.wantWindowID).
					Return(nil).
					Times(1)

				root := cmd.RootCmd(aerospaceClient)
				if _, err := testutils.CmdExecute(root, test.args...); err != nil {
					t.Fatalf("next command failed: %v", err)
				}
			})
		}
	})

	t.Run("fails when getting the focused window returns an unexpected error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
		scratchpadWindows := []windows.Window{
			{
				WindowID:  8888,
				Workspace: constants.DefaultScratchpadWorkspaceName,
			},
		}

		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "work"}, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			Return(scratchpadWindows, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
			Return(scratchpadWindows, nil).
			Times(1)
		aerospaceClient.GetWindowsMock().EXPECT().
			GetFocusedWindow().
			Return(nil, errors.New("mocked focus error")).
			Times(1)

		root := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(root, "next")
		if err == nil {
			t.Fatal("expected focused window error")
		}
		if !strings.Contains(err.Error(), "unable to get focused window: mocked focus error") {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "" {
			t.Fatalf("expected no output, got %q", out)
		}
	})

	t.Run(
		"fails when getting focused workspace returns an error",
		func(t *testing.T) {
			command := "next"
			args := []string{command}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			gomock.InOrder(
				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(nil, errors.New("mocked_error")).
					Times(1),
			)

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

			cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
			testutils.MatchSnapshot(t, nil, cmdAsString, out, err)
		},
	)

	t.Run(
		"fails when no scratchpad windows available",
		func(t *testing.T) {
			command := "next"
			args := []string{command}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}
			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			gomock.InOrder(
				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(focusedWorkspace, nil).
					Times(1),
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindows().
					Return([]windows.Window{}, nil).
					Times(1),
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
					Return([]windows.Window{}, nil).
					Times(1),
			)

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

			cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
			testutils.MatchSnapshot(t, nil, cmdAsString, out, err)
		},
	)

	t.Run(
		"fails when moving window to workspace returns an error",
		func(t *testing.T) {
			command := "next"
			args := []string{command}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			focusedWorkspace := &workspaces.Workspace{Workspace: "ws1"}
			scratchpadWindows := []windows.Window{
				{
					AppName:  "Scratchpad Window",
					WindowID: 8888,
				},
			}
			windowID := 8888
			aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
			gomock.InOrder(
				aerospaceClient.GetWorkspacesMock().EXPECT().
					GetFocusedWorkspace().
					Return(focusedWorkspace, nil).
					Times(1),
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindows().
					Return([]windows.Window{}, nil).
					Times(1),
				aerospaceClient.GetWindowsMock().EXPECT().
					GetAllWindowsByWorkspace(constants.DefaultScratchpadWorkspaceName).
					Return(scratchpadWindows, nil).
					Times(1),
				aerospaceClient.GetWindowsMock().EXPECT().
					GetFocusedWindow().
					Return(&windows.Window{WindowID: 7777}, nil).
					Times(1),
				aerospaceClient.GetWorkspacesMock().EXPECT().
					MoveWindowToWorkspaceWithOpts(
						workspaces.MoveWindowToWorkspaceArgs{
							WorkspaceName: focusedWorkspace.Workspace,
						},
						workspaces.MoveWindowToWorkspaceOpts{
							WindowID: &windowID,
						},
					).
					Return(errors.New("mocked_move_error")).
					Times(1),
			)

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

			cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
			testutils.MatchSnapshot(t, nil, cmdAsString, out, err)
		},
	)
}
