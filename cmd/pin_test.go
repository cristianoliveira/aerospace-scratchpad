package cmd_test

import (
	"path/filepath"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/stderr"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

func TestPinCmdPinsFocusedWindowToCurrentMonitor(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)
	t.Setenv("AEROSPACE_SCRATCHPAD_PINS_PATH", filepath.Join(t.TempDir(), "pins.json"))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	focusedWindow := &windows.Window{
		WindowID:  1234,
		AppName:   "Finder",
		Workspace: "ws2",
	}
	aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
	aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
		{Workspace: "ws2", MonitorID: 2},
	})
	aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 1})

	aerospaceClient.GetWindowsMock().EXPECT().
		GetFocusedWindow().
		Return(focusedWindow, nil).
		Times(1)

	rootCmd := cmd.RootCmd(aerospaceClient)
	_, err := testutils.CmdExecute(rootCmd, "pin")
	if err != nil {
		t.Fatalf("pin command failed: %v", err)
	}

	monitorID, ok, err := aerospace.PinnedMonitorID(focusedWindow.WindowID)
	if err != nil {
		t.Fatalf("read pin: %v", err)
	}
	if !ok || monitorID != 2 {
		t.Fatalf(
			"expected window pinned to its workspace monitor 2, got ok=%v monitor=%d",
			ok,
			monitorID,
		)
	}
}

func TestPinCmdStoresPatternRule(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)
	t.Setenv("AEROSPACE_SCRATCHPAD_PINS_PATH", filepath.Join(t.TempDir(), "pins.json"))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	braveWindow := windows.Window{
		WindowID:  1234,
		AppName:   "Brave Browser",
		Workspace: "ws2",
	}
	aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
	aerospaceClient.SetWorkspaceMonitors([]aerospace.WorkspaceMonitor{
		{Workspace: "ws2", MonitorID: 2},
	})
	aerospaceClient.GetWindowsMock().EXPECT().
		GetAllWindows().
		Return([]windows.Window{braveWindow}, nil).
		Times(1)

	rootCmd := cmd.RootCmd(aerospaceClient)
	_, err := testutils.CmdExecute(rootCmd, "pin", ".*Brave.*")
	if err != nil {
		t.Fatalf("pin command failed: %v", err)
	}

	monitorID, ok, err := aerospace.PinnedMonitorIDForWindow(windows.Window{
		WindowID: 9876,
		AppName:  "Brave Browser",
	})
	if err != nil {
		t.Fatalf("read pin rule: %v", err)
	}
	if !ok || monitorID != 2 {
		t.Fatalf(
			"expected future Brave window pinned to monitor 2, got ok=%v monitor=%d",
			ok,
			monitorID,
		)
	}
}

func TestPinCmdDisablesAndEnablesPatternRule(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)
	t.Setenv("AEROSPACE_SCRATCHPAD_PINS_PATH", filepath.Join(t.TempDir(), "pins.json"))

	if err := aerospace.PinRuleForPattern(".*Brave.*", nil, 2); err != nil {
		t.Fatalf("seed pin rule: %v", err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
	rootCmd := cmd.RootCmd(aerospaceClient)

	_, err := testutils.CmdExecute(rootCmd, "pin", "--disable", ".*Brave.*")
	if err != nil {
		t.Fatalf("disable pin rule command failed: %v", err)
	}
	_, ok, err := aerospace.PinnedMonitorIDForWindow(windows.Window{
		AppName: "Brave Browser",
	})
	if err != nil {
		t.Fatalf("read disabled pin rule: %v", err)
	}
	if ok {
		t.Fatal("expected disabled pin rule not to match")
	}

	rootCmd = cmd.RootCmd(aerospaceClient)
	_, err = testutils.CmdExecute(rootCmd, "pin", "--enable", ".*Brave.*")
	if err != nil {
		t.Fatalf("enable pin rule command failed: %v", err)
	}
	monitorID, ok, err := aerospace.PinnedMonitorIDForWindow(windows.Window{
		AppName: "Brave Browser",
	})
	if err != nil {
		t.Fatalf("read enabled pin rule: %v", err)
	}
	if !ok || monitorID != 2 {
		t.Fatalf("expected enabled pin rule on monitor 2, got ok=%v monitor=%d", ok, monitorID)
	}
}

func TestUnpinCmdRemovesFocusedWindowPin(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)
	t.Setenv("AEROSPACE_SCRATCHPAD_PINS_PATH", filepath.Join(t.TempDir(), "pins.json"))

	if err := aerospace.PinWindow(1234, 2); err != nil {
		t.Fatalf("seed pin: %v", err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	focusedWindow := &windows.Window{
		WindowID:  1234,
		AppName:   "Finder",
		Workspace: "ws2",
	}
	aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
	aerospaceClient.GetWindowsMock().EXPECT().
		GetFocusedWindow().
		Return(focusedWindow, nil).
		Times(1)

	rootCmd := cmd.RootCmd(aerospaceClient)
	_, err := testutils.CmdExecute(rootCmd, "unpin")
	if err != nil {
		t.Fatalf("unpin command failed: %v", err)
	}

	_, ok, err := aerospace.PinnedMonitorID(focusedWindow.WindowID)
	if err != nil {
		t.Fatalf("read pin: %v", err)
	}
	if ok {
		t.Fatal("expected window to be unpinned")
	}
}

func TestPinnedSummonFocusesWindowOnOtherMonitorWithoutMoving(t *testing.T) {
	logger.SetDefaultLogger(&logger.EmptyLogger{})
	stderr.SetBehavior(false)
	t.Setenv("AEROSPACE_SCRATCHPAD_PINS_PATH", filepath.Join(t.TempDir(), "pins.json"))

	seedErr := aerospace.PinWindow(1234, 2)
	if seedErr != nil {
		t.Fatalf("seed pin: %v", seedErr)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pinnedWindow := windows.Window{
		WindowID:  1234,
		AppName:   "Finder",
		Workspace: ".scratchpad.2",
	}
	aerospaceClient := testutils.NewMockAeroSpaceWM(ctrl)
	aerospaceClient.SetFocusedMonitor(aerospace.MonitorInfo{MonitorID: 1})

	gomock.InOrder(
		aerospaceClient.GetWorkspacesMock().EXPECT().
			GetFocusedWorkspace().
			Return(&workspaces.Workspace{Workspace: "ws1"}, nil).
			Times(1),
		aerospaceClient.GetWindowsMock().EXPECT().
			GetAllWindows().
			Return([]windows.Window{pinnedWindow}, nil).
			Times(1),
		aerospaceClient.GetFocusMock().EXPECT().
			SetFocusByWindowID(pinnedWindow.WindowID).
			Return(nil).
			Times(1),
	)

	rootCmd := cmd.RootCmd(aerospaceClient)
	_, err := testutils.CmdExecute(rootCmd, "summon", "Finder")
	if err != nil {
		t.Fatalf("summon command failed: %v", err)
	}
}
