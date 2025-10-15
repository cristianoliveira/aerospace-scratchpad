package cmd_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"go.uber.org/mock/gomock"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	mock_aerospace "github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

//nolint:gocognit // Integration-style test exercises multiple window scenarios for coverage
func TestShowCmd(t *testing.T) {
	t.Skip("Skipping integration-style test that requires AeroSpace WM")
	logger.SetDefaultLogger(&logger.EmptyLogger{})

	t.Run("fails when pattern is empty", func(t *testing.T) {
		command := "show"
		args := []string{command, ""}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		stderr.SetBehavior(false)

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []aerospacecli.Window{
					{
						AppName:  "Notepad",
						WindowID: 1234,
					},
					{
						AppName:  "Finder",
						WindowID: 5678,
					},
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: "ws1",
				},

				FocusedWindowID: 5678,
			},
		}

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error, got %v", out)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})

	t.Run("fails when pattern doesn match any window", func(t *testing.T) {
		command := "show"
		args := []string{command, "foo"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []aerospacecli.Window{
					{
						AppName:  "Notepad",
						WindowID: 1234,
					},
					{
						AppName:  "Finder",
						WindowID: 5678,
					},
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: "ws1",
				},

				FocusedWindowID: 1234,
			},
		}

		allWindows := testutils.ExtractAllWindows(tree)
		focusedTree := testutils.ExtractFocusedTree(tree)

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		aerospaceClient.EXPECT().
			GetAllWindows().
			Return(allWindows, nil).
			Times(1)

		aerospaceClient.EXPECT().
			GetFocusedWorkspace().
			Return(focusedTree.Workspace, nil).
			Times(1)

		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error, got %v", out)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})

	t.Run(
		"set focus to window if already in the focused workspace but not focused",
		func(t *testing.T) {
			command := "show"
			args := []string{command, "Finder"}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{
				{
					Windows: []aerospacecli.Window{
						{
							AppName:  "Notepad",
							WindowID: 1234,
						},
						{
							AppName:  "Finder",
							WindowID: 5678,
						},
					},
					Workspace: &aerospacecli.Workspace{
						Workspace: "ws1",
					},

					FocusedWindowID: 1234,
				},
			}

			allWindows := testutils.ExtractAllWindows(tree)
			focusedTree := testutils.ExtractFocusedTree(tree)
			focusedWindow := testutils.ExtractFocusedWindow(tree)

			aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1)

			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedTree.Workspace, nil).
				Times(1)

			gomock.InOrder(
				aerospaceClient.EXPECT().
					GetAllWindowsByWorkspace(focusedTree.Workspace.Workspace).
					Return(focusedTree.Windows, nil).
					Times(1),

				aerospaceClient.EXPECT().
					GetFocusedWindow().
					Return(focusedWindow, nil).
					Times(1),

				aerospaceClient.EXPECT().
					SetFocusByWindowID(focusedTree.Windows[1].WindowID).
					Return(nil).
					Times(1),

				// DO NOT set the layout to floating
				aerospaceClient.EXPECT().
					SetLayout(gomock.Any(), "floating").
					Return(nil).
					Times(0),
			)

			cmd := cmd.RootCmd(aerospaceClient)
			out, err := testutils.CmdExecute(cmd, args...)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if out == "" {
				t.Errorf("Expected output, got empty string")
			}

			cmdAsString := "aerospace-scratchpad " + strings.Join(
				args,
				" ",
			) + "\n"
			expectedError := fmt.Sprintf("Error\n%+v", err)
			snaps.MatchSnapshot(
				t,
				tree,
				cmdAsString,
				"Output",
				out,
				expectedError,
			)
		},
	)

	t.Run("moves a window to scratchpad by pattern", func(t *testing.T) {
		command := "show"
		args := []string{command, "Finder"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tree := []testutils.AeroSpaceTree{
			{
				Windows: []aerospacecli.Window{
					{
						AppName:  "Notepad",
						WindowID: 1234,
					},
					{
						AppName:  "Finder",
						WindowID: 5678,
					},
				},
				Workspace: &aerospacecli.Workspace{
					Workspace: "ws1",
				},

				FocusedWindowID: 5678,
			},
		}

		allWindows := testutils.ExtractAllWindows(tree)
		focusedTree := testutils.ExtractFocusedTree(tree)
		focusedWindow := testutils.ExtractFocusedWindow(tree)

		aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
		aerospaceClient.EXPECT().
			GetAllWindows().
			Return(allWindows, nil).
			Times(1)

		aerospaceClient.EXPECT().
			GetFocusedWorkspace().
			Return(focusedTree.Workspace, nil).
			Times(1)

		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetAllWindowsByWorkspace("ws1").
				Return(focusedTree.Windows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				GetFocusedWindow().
				Return(focusedWindow, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(
					focusedWindow.WindowID,
					constants.DefaultScratchpadWorkspaceName).
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetFocusByWindowID(focusedWindow.WindowID).
				Return(nil).
				Times(0),

			// When moving to scratchpad, set the layout to floating
			aerospaceClient.EXPECT().
				SetLayout(focusedWindow.WindowID, "floating").
				Return(nil).
				Times(1),
		)

		cmd := cmd.RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if out == "" {
			t.Errorf("Expected output, got empty string")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		expectedError := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, tree, cmdAsString, "Output", out, expectedError)
	})

	t.Run(
		"summon the window to the current workspace if in another workspace",
		func(t *testing.T) {
			command := "show"
			args := []string{command, "Finder"}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{
				{
					Windows: []aerospacecli.Window{
						{
							AppName:  "Notepad",
							WindowID: 1234,
						},
						{
							AppName:  "Finder",
							WindowID: 5678,
						},
					},
					Workspace: &aerospacecli.Workspace{
						Workspace: "ws1",
					},
					FocusedWindowID: 0, // Not focused
				},
				{
					Windows: []aerospacecli.Window{
						{
							AppName:  "Terminal",
							WindowID: 91011,
						},
					},
					Workspace: &aerospacecli.Workspace{
						Workspace: "ws2",
					},
					FocusedWindowID: 91011,
				},
			}

			allWindows := testutils.ExtractAllWindows(tree)
			focusedTree := testutils.ExtractFocusedTree(tree)
			focusedWindow := testutils.ExtractFocusedWindow(tree)

			aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1)

			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedTree.Workspace, nil).
				Times(1)

			gomock.InOrder(
				aerospaceClient.EXPECT().
					GetAllWindowsByWorkspace(focusedTree.Workspace.Workspace).
					Return(focusedTree.Windows, nil).
					Times(1),

				aerospaceClient.EXPECT().
					MoveWindowToWorkspace(
						tree[0].Windows[1].WindowID,
						focusedTree.Workspace.Workspace).
					Return(nil).
					Times(1),

				aerospaceClient.EXPECT().
					SetFocusByWindowID(
						tree[0].Windows[1].WindowID).
					Return(nil).
					Times(1),

				// When moving to scratchpad, set the layout to floating
				aerospaceClient.EXPECT().
					SetLayout(focusedWindow.WindowID, "floating").
					Return(nil).
					Times(0),
			)

			cmd := cmd.RootCmd(aerospaceClient)
			out, err := testutils.CmdExecute(cmd, args...)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if out == "" {
				t.Errorf("Expected output, got empty string")
			}

			cmdAsString := "aerospace-scratchpad " + strings.Join(
				args,
				" ",
			) + "\n"
			expectedError := fmt.Sprintf("Error\n%+v", err)
			snaps.MatchSnapshot(
				t,
				tree,
				cmdAsString,
				"Output",
				out,
				expectedError,
			)
		},
	)

	t.Run("MultipleWindows", func(tt *testing.T) {
		tt.Run("brings all windows to focused workspace", func(t *testing.T) {
			command := "show"
			args := []string{command, "Finder"}

			ctrl := gomock.NewController(tt)
			defer ctrl.Finish()

			tree := []testutils.AeroSpaceTree{
				{
					Windows: []aerospacecli.Window{
						{
							AppName:   "Finder1",
							WindowID:  5678,
							Workspace: "ws1",
						},
						{
							AppName:   "Finder2",
							WindowID:  5679,
							Workspace: "ws1",
						},
					},
					Workspace: &aerospacecli.Workspace{
						Workspace: "ws1",
					},
					FocusedWindowID: 0, // Not focused
				},
				{
					Windows: []aerospacecli.Window{
						{
							AppName:   "Terminal",
							WindowID:  91011,
							Workspace: "ws2",
						},
					},
					Workspace: &aerospacecli.Workspace{
						Workspace: "ws2",
					},
					FocusedWindowID: 91011,
				},
			}

			allWindows := testutils.ExtractAllWindows(tree)
			focusedTree := testutils.ExtractFocusedTree(tree)
			// focusedWindow := testutils.ExtractFocusedWindow(tree)

			aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)

			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(allWindows, nil).
				Times(1)

			aerospaceClient.EXPECT().
				GetFocusedWorkspace().
				Return(focusedTree.Workspace, nil).
				Times(1)

			gomock.InOrder(
				// Send first window
				aerospaceClient.EXPECT().
					MoveWindowToWorkspace(
						tree[0].Windows[0].WindowID,
						focusedTree.Workspace.Workspace,
					).
					Return(nil).
					Times(1),
				aerospaceClient.EXPECT().
					SetFocusByWindowID(
						tree[0].Windows[0].WindowID,
					).
					Return(nil).
					Times(1),

				// Send 2nd window
				aerospaceClient.EXPECT().
					MoveWindowToWorkspace(
						tree[0].Windows[1].WindowID,
						focusedTree.Workspace.Workspace,
					).
					Return(nil).
					Times(1),
				aerospaceClient.EXPECT().
					SetFocusByWindowID(
						tree[0].Windows[1].WindowID,
					).
					Return(nil).
					Times(1),
			)

			cmd := cmd.RootCmd(aerospaceClient)
			out, err := testutils.CmdExecute(cmd, args...)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if out == "" {
				t.Errorf("Expected output, got empty string")
			}

			cmdAsString := "aerospace-scratchpad " + strings.Join(
				args,
				" ",
			) + "\n"
			expectedError := fmt.Sprintf("Error\n%+v", err)
			snaps.MatchSnapshot(
				t,
				tree,
				cmdAsString,
				"Output",
				out,
				expectedError,
			)
		})

		tt.Run(
			"sends all windows to scratchpad if at least one window is focused",
			func(t *testing.T) {
				command := "show"
				args := []string{command, "Finder"}

				ctrl := gomock.NewController(tt)
				defer ctrl.Finish()

				tree := []testutils.AeroSpaceTree{
					{
						Windows: []aerospacecli.Window{},
						Workspace: &aerospacecli.Workspace{
							Workspace: "ws1",
						},
						FocusedWindowID: 0, // Not focused
					},
					{
						Windows: []aerospacecli.Window{
							{
								AppName:   "Finder1",
								WindowID:  5678,
								Workspace: "ws2",
							},
							{
								AppName:   "Finder2",
								WindowID:  5679,
								Workspace: "ws2",
							},
							{
								AppName:   "Terminal",
								WindowID:  91011,
								Workspace: "ws2",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: "ws2",
						},
						FocusedWindowID: 5678,
					},
				}

				allWindows := testutils.ExtractAllWindows(tree)
				focusedTree := testutils.ExtractFocusedTree(tree)
				focusedWindow := testutils.ExtractFocusedWindow(tree)

				aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
				aerospaceClient.EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1)

				aerospaceClient.EXPECT().
					GetFocusedWorkspace().
					Return(focusedTree.Workspace, nil).
					Times(1)

				gomock.InOrder(
					aerospaceClient.EXPECT().
						GetFocusedWindow().
						Return(focusedWindow, nil).
						Times(2),

					aerospaceClient.EXPECT().
						MoveWindowToWorkspace(
							tree[1].Windows[0].WindowID,
							constants.DefaultScratchpadWorkspaceName,
						).
						Return(nil).
						Times(1),
					aerospaceClient.EXPECT().
						SetLayout(
							tree[1].Windows[0].WindowID,
							"floating",
						).
						Return(nil).
						Times(1),

					aerospaceClient.EXPECT().
						MoveWindowToWorkspace(
							tree[1].Windows[1].WindowID,
							constants.DefaultScratchpadWorkspaceName,
						).
						Return(nil).
						Times(1),
					aerospaceClient.EXPECT().
						SetLayout(
							tree[1].Windows[1].WindowID,
							"floating",
						).
						Return(nil).
						Times(1),
				)

				cmd := cmd.RootCmd(aerospaceClient)
				out, err := testutils.CmdExecute(cmd, args...)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if out == "" {
					t.Errorf("Expected output, got empty string")
				}

				cmdAsString := "aerospace-scratchpad " + strings.Join(
					args,
					" ",
				) + "\n"
				expectedError := fmt.Sprintf("Error\n%+v", err)
				snaps.MatchSnapshot(
					t,
					tree,
					cmdAsString,
					"Output",
					out,
					expectedError,
				)
			},
		)

		tt.Run(
			"gives priority to bringing scratchpads together",
			func(t *testing.T) {
				command := "show"
				args := []string{command, "Finder"}

				ctrl := gomock.NewController(tt)
				defer ctrl.Finish()

				tree := []testutils.AeroSpaceTree{
					{
						Windows: []aerospacecli.Window{
							{
								AppName:   "Finder1",
								WindowID:  5678,
								Workspace: "ws1",
							},
							{
								AppName:   "Browser",
								WindowID:  22,
								Workspace: "ws1",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: "ws1",
						},
						FocusedWindowID: 0, // Not focused
					},
					{
						Windows: []aerospacecli.Window{
							{
								AppName:   "Finder2",
								WindowID:  5679,
								Workspace: "ws2",
							},
							{
								AppName:   "Terminal",
								WindowID:  91011,
								Workspace: "ws2",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: "ws2",
						},
						FocusedWindowID: 91011,
					},
				}

				allWindows := testutils.ExtractAllWindows(tree)
				focusedTree := testutils.ExtractFocusedTree(tree)
				focusedWindow := testutils.ExtractFocusedWindow(tree)

				aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
				aerospaceClient.EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1)

				aerospaceClient.EXPECT().
					GetFocusedWorkspace().
					Return(focusedTree.Workspace, nil).
					Times(1)

				gomock.InOrder(
					aerospaceClient.EXPECT().
						GetFocusedWindow().
						Return(focusedWindow, nil).
						Times(1),

					aerospaceClient.EXPECT().
						MoveWindowToWorkspace(
							tree[0].Windows[0].WindowID,
							focusedTree.Workspace.Workspace,
						).
						Return(nil).
						Times(1),

					aerospaceClient.EXPECT().
						SetFocusByWindowID(
							tree[0].Windows[0].WindowID,
						).
						Return(nil).
						Times(1),

					aerospaceClient.EXPECT().
						SetFocusByWindowID(
							tree[1].Windows[0].WindowID,
						).
						Return(nil).
						Times(1),
				)

				cmd := cmd.RootCmd(aerospaceClient)
				out, err := testutils.CmdExecute(cmd, args...)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if out == "" {
					t.Errorf("Expected output, got empty string")
				}

				cmdAsString := "aerospace-scratchpad " + strings.Join(
					args,
					" ",
				) + "\n"
				expectedError := fmt.Sprintf("Error\n%+v", err)
				snaps.MatchSnapshot(
					t,
					tree,
					cmdAsString,
					"Output",
					out,
					expectedError,
				)
			},
		)

		tt.Run(
			"when bringing windows together, it doesnt change focus",
			func(t *testing.T) {
				command := "show"
				args := []string{command, "Finder"}

				ctrl := gomock.NewController(tt)
				defer ctrl.Finish()

				tree := []testutils.AeroSpaceTree{
					{
						Windows: []aerospacecli.Window{
							{
								AppName:   "Finder1",
								WindowID:  5678,
								Workspace: "ws1",
							},
							{
								AppName:   "Browser",
								WindowID:  22,
								Workspace: "ws1",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: "ws1",
						},
						FocusedWindowID: 0, // Not focused
					},
					{
						Windows: []aerospacecli.Window{
							{
								AppName:   "Finder2",
								WindowID:  5679,
								Workspace: "ws2",
							},
							{
								AppName:   "Terminal",
								WindowID:  91011,
								Workspace: "ws2",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: "ws2",
						},
						FocusedWindowID: 5679,
					},
				}

				allWindows := testutils.ExtractAllWindows(tree)
				focusedTree := testutils.ExtractFocusedTree(tree)
				focusedWindow := testutils.ExtractFocusedWindow(tree)

				aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
				aerospaceClient.EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1)

				aerospaceClient.EXPECT().
					GetFocusedWorkspace().
					Return(focusedTree.Workspace, nil).
					Times(1)

				gomock.InOrder(
					aerospaceClient.EXPECT().
						GetFocusedWindow().
						Return(focusedWindow, nil).
						Times(1),

					aerospaceClient.EXPECT().
						MoveWindowToWorkspace(
							tree[0].Windows[0].WindowID,
							focusedTree.Workspace.Workspace,
						).
						Return(nil).
						Times(1),

					aerospaceClient.EXPECT().
						SetFocusByWindowID(
							tree[1].Windows[0].WindowID,
						).
						Return(nil).
						Times(1),
				)

				cmd := cmd.RootCmd(aerospaceClient)
				out, err := testutils.CmdExecute(cmd, args...)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if out == "" {
					t.Errorf("Expected output, got empty string")
				}

				cmdAsString := "aerospace-scratchpad " + strings.Join(
					args,
					" ",
				) + "\n"
				expectedError := fmt.Sprintf("Error\n%+v", err)
				snaps.MatchSnapshot(
					t,
					tree,
					cmdAsString,
					"Output",
					out,
					expectedError,
				)
			},
		)

		tt.Run(
			"Filter flag: brings any windows that matches filter",
			func(ttt *testing.T) {
				command := "show"
				args := []string{
					command,
					"Finder",
					"--filter",
					"window-title=foo",
				}

				ctrl := gomock.NewController(ttt)
				defer ctrl.Finish()

				tree := []testutils.AeroSpaceTree{
					{
						Windows: []aerospacecli.Window{
							{
								AppName:     "Finder1",
								WindowID:    5678,
								WindowTitle: "Finder - foo and zas",
								Workspace:   "ws1",
							},
							{
								AppName:   "Finder2 - bar and baz",
								WindowID:  5679,
								Workspace: "ws1",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: ".scratchpad",
						},
						FocusedWindowID: 0, // Not focused
					},
					{
						Windows: []aerospacecli.Window{
							{
								AppName:   "Terminal",
								WindowID:  91011,
								Workspace: "ws2",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: "ws2",
						},
						FocusedWindowID: 91011,
					},
				}

				allWindows := testutils.ExtractAllWindows(tree)
				focusedTree := testutils.ExtractFocusedTree(tree)
				// focusedWindow := testutils.ExtractFocusedWindow(tree)

				aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
				aerospaceClient.EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1)

				aerospaceClient.EXPECT().
					GetFocusedWorkspace().
					Return(focusedTree.Workspace, nil).
					Times(1)

				gomock.InOrder(
					// Send first window
					aerospaceClient.EXPECT().
						MoveWindowToWorkspace(
							tree[0].Windows[0].WindowID,
							focusedTree.Workspace.Workspace,
						).
						Return(nil).
						Times(1),
					aerospaceClient.EXPECT().
						SetFocusByWindowID(
							tree[0].Windows[0].WindowID,
						).
						Return(nil).
						Times(1),
				)

				cmd := cmd.RootCmd(aerospaceClient)
				out, err := testutils.CmdExecute(cmd, args...)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if out == "" {
					t.Errorf("Expected output, got empty string")
				}

				cmdAsString := "aerospace-scratchpad " + strings.Join(
					args,
					" ",
				) + "\n"
				expectedError := fmt.Sprintf("Error\n%+v", err)
				snaps.MatchSnapshot(
					ttt,
					tree,
					cmdAsString,
					"Output",
					out,
					expectedError,
				)
			},
		)

		tt.Run(
			"Filter flag: brings any windows that matches filter - allow multiple",
			func(ttt *testing.T) {
				command := "show"
				args := []string{command,
					"Finder",
					"-F", "window-title=foo",
					"-F", "app-bundle-id=linux",
				}

				ctrl := gomock.NewController(ttt)
				defer ctrl.Finish()

				tree := []testutils.AeroSpaceTree{
					{
						Windows: []aerospacecli.Window{
							{
								AppName:     "Finder1",
								WindowID:    5678,
								WindowTitle: "Finder - foo and zas",
								AppBundleID: "com.linux.finder",
								Workspace:   "ws1",
							},
							{
								AppName:     "Finder2",
								WindowID:    5679,
								WindowTitle: "Finder2 - foo and baz",
								AppBundleID: "com.apple.finder",
								Workspace:   "ws1",
							},
							{
								AppName:     "Finder2",
								WindowID:    5680,
								WindowTitle: "Finder2 - bar and baz",
								AppBundleID: "com.apple.finder",
								Workspace:   "ws1",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: ".scratchpad",
						},
						FocusedWindowID: 0, // Not focused
					},
					{
						Windows: []aerospacecli.Window{
							{
								AppName:   "Terminal",
								WindowID:  91011,
								Workspace: "ws2",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: "ws2",
						},
						FocusedWindowID: 91011,
					},
				}

				allWindows := testutils.ExtractAllWindows(tree)
				focusedTree := testutils.ExtractFocusedTree(tree)
				// focusedWindow := testutils.ExtractFocusedWindow(tree)

				aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
				aerospaceClient.EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1)

				aerospaceClient.EXPECT().
					GetFocusedWorkspace().
					Return(focusedTree.Workspace, nil).
					Times(1)

				gomock.InOrder(
					// Send first window
					aerospaceClient.EXPECT().
						MoveWindowToWorkspace(
							tree[0].Windows[0].WindowID,
							focusedTree.Workspace.Workspace,
						).
						Return(nil).
						Times(1),
					aerospaceClient.EXPECT().
						SetFocusByWindowID(
							tree[0].Windows[0].WindowID,
						).
						Return(nil).
						Times(1),
				)

				cmd := cmd.RootCmd(aerospaceClient)
				out, err := testutils.CmdExecute(cmd, args...)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if out == "" {
					t.Errorf("Expected output, got empty string")
				}

				cmdAsString := "aerospace-scratchpad " + strings.Join(
					args,
					" ",
				) + "\n"
				expectedError := fmt.Sprintf("Error\n%+v", err)
				snaps.MatchSnapshot(
					ttt,
					tree,
					cmdAsString,
					"Output",
					out,
					expectedError,
				)
			},
		)
		// Test fail unkonwn filter property
		tt.Run(
			"Filter flag: fails when unknown filter property is used",
			func(ttt *testing.T) {
				command := "show"
				args := []string{command, "Finder", "--filter", "unknown=foo"}

				ctrl := gomock.NewController(ttt)
				defer ctrl.Finish()

				tree := []testutils.AeroSpaceTree{
					{
						Windows: []aerospacecli.Window{
							{
								AppName:   "Finder1",
								WindowID:  5678,
								Workspace: "ws1",
							},
							{
								AppName:   "Finder2",
								WindowID:  5670,
								Workspace: "ws1",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: ".scratchpad",
						},
						FocusedWindowID: 5670, // Not focused
					},
				}

				allWindows := testutils.ExtractAllWindows(tree)
				focusedTree := testutils.ExtractFocusedTree(tree)

				aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
				aerospaceClient.EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1)

				aerospaceClient.EXPECT().
					GetFocusedWorkspace().
					Return(focusedTree.Workspace, nil).
					Times(1)

				cmd := cmd.RootCmd(aerospaceClient)
				out, err := testutils.CmdExecute(cmd, args...)
				if err == nil {
					t.Errorf("Expected error, got %v", out)
				}

				cmdAsString := "aerospace-scratchpad " + strings.Join(
					args,
					" ",
				) + "\n"
				expectedError := fmt.Sprintf("Error\n%+v", err)
				snaps.MatchSnapshot(
					ttt,
					tree,
					cmdAsString,
					"Output",
					out,
					expectedError,
				)
			},
		)

		tt.Run(
			"Filter flag: shows error messages when no match",
			func(ttt *testing.T) {
				command := "show"
				args := []string{
					command,
					"Finder",
					"--filter",
					"window-title=cantfindme",
				}

				ctrl := gomock.NewController(ttt)
				defer ctrl.Finish()

				tree := []testutils.AeroSpaceTree{
					{
						Windows: []aerospacecli.Window{
							{
								AppName:     "Finder1",
								WindowID:    5678,
								WindowTitle: "Finder - foo and zas",
								Workspace:   "ws1",
							},
							{
								AppName:   "Finder2 - bar and baz",
								WindowID:  5679,
								Workspace: "ws1",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: ".scratchpad",
						},
						FocusedWindowID: 0, // Not focused
					},
					{
						Windows: []aerospacecli.Window{
							{
								AppName:   "Terminal",
								WindowID:  91011,
								Workspace: "ws2",
							},
						},
						Workspace: &aerospacecli.Workspace{
							Workspace: "ws2",
						},
						FocusedWindowID: 91011,
					},
				}

				allWindows := testutils.ExtractAllWindows(tree)
				focusedTree := testutils.ExtractFocusedTree(tree)
				// focusedWindow := testutils.ExtractFocusedWindow(tree)

				aerospaceClient := mock_aerospace.NewMockAeroSpaceClient(ctrl)
				aerospaceClient.EXPECT().
					GetAllWindows().
					Return(allWindows, nil).
					Times(1)

				aerospaceClient.EXPECT().
					GetFocusedWorkspace().
					Return(focusedTree.Workspace, nil).
					Times(1)

				cmd := cmd.RootCmd(aerospaceClient)
				out, err := testutils.CmdExecute(cmd, args...)
				if err == nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if out != "" {
					t.Errorf("Expected output, got empty string")
				}

				cmdAsString := "aerospace-scratchpad " + strings.Join(
					args,
					" ",
				) + "\n"
				expectedError := fmt.Sprintf("Error\n%+v", err)
				snaps.MatchSnapshot(
					ttt,
					tree,
					cmdAsString,
					"Output",
					out,
					expectedError,
				)
			},
		)
	})
}
