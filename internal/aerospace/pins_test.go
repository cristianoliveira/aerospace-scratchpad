package aerospace_test

import (
	"path/filepath"
	"testing"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
)

func TestPinsState(t *testing.T) {
	t.Setenv("AEROSPACE_SCRATCHPAD_PINS_PATH", filepath.Join(t.TempDir(), "pins.json"))

	_, ok, err := aerospace.PinnedMonitorID(123)
	if err != nil {
		t.Fatalf("expected missing pins file to load empty state: %v", err)
	}
	if ok {
		t.Fatal("expected window to start unpinned")
	}

	pinErr := aerospace.PinWindow(123, 2)
	if pinErr != nil {
		t.Fatalf("pin window: %v", pinErr)
	}

	monitorID, ok, err := aerospace.PinnedMonitorID(123)
	if err != nil {
		t.Fatalf("read pin: %v", err)
	}
	if !ok || monitorID != 2 {
		t.Fatalf("expected window pinned to monitor 2, got ok=%v monitor=%d", ok, monitorID)
	}

	unpinErr := aerospace.UnpinWindow(123)
	if unpinErr != nil {
		t.Fatalf("unpin window: %v", unpinErr)
	}

	_, ok, err = aerospace.PinnedMonitorID(123)
	if err != nil {
		t.Fatalf("read after unpin: %v", err)
	}
	if ok {
		t.Fatal("expected window to be unpinned")
	}
}

func TestPinRulesMatchFutureWindowsByPattern(t *testing.T) {
	t.Setenv("AEROSPACE_SCRATCHPAD_PINS_PATH", filepath.Join(t.TempDir(), "pins.json"))

	pinErr := aerospace.PinRuleForPattern(".*Brave.*", nil, 2)
	if pinErr != nil {
		t.Fatalf("pin rule: %v", pinErr)
	}

	monitorID, ok, err := aerospace.PinnedMonitorIDForWindow(windows.Window{
		WindowID: 9876,
		AppName:  "Brave Browser",
	})
	if err != nil {
		t.Fatalf("match pin rule: %v", err)
	}
	if !ok || monitorID != 2 {
		t.Fatalf("expected pattern-pinned window on monitor 2, got ok=%v monitor=%d", ok, monitorID)
	}
}
