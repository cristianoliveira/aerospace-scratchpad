package aerospace

import (
	"errors"
	"fmt"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

type Mover interface {
	// MoveWindowToScratchpad sends a window to the scratchpad attached to the
	// window's own source monitor.
	MoveWindowToScratchpad(window windows.Window) (string, error)

	// MoveWindowToScratchpadForMonitor sends a window to a scratchpad workspace
	// attached to the given source monitor. The monitor invariant is enforced
	// best-effort (check-then-move over separate IPC calls, not race-free): if a
	// same-monitor target cannot be established (unknown monitor, or target
	// attached elsewhere), it returns an error and leaves the window where it
	// is. A missing target is provisioned on the source monitor: focus moves
	// to the source window, the empty workspace is summoned, its placement is
	// verified, and the previous focus state is restored best-effort.
	MoveWindowToScratchpadForMonitor(window windows.Window, monitorID int) (string, error)

	// MoveWindowToWorkspace sends a window to a workspace and set focus
	MoveWindowToWorkspace(
		window *windows.Window,
		workspace *workspaces.Workspace,
		shouldSetFocus bool,
	) error
}

type MoverAeroSpace struct {
	aerospace AeroSpaceWMClient
}

func NewAeroSpaceMover(aerospace AeroSpaceWMClient) MoverAeroSpace {
	return MoverAeroSpace{
		aerospace: aerospace,
	}
}

// MoveWindowToScratchpad sends the window to the scratchpad attached to the
// window's own source monitor, resolved from the workspace-to-monitor mapping.
// It fails closed when the source monitor or a safe same-monitor target cannot
// be established.
func (a *MoverAeroSpace) MoveWindowToScratchpad(
	window windows.Window,
) (string, error) {
	logger := logger.GetDefaultLogger()
	logger.LogDebug("MOVING: MoveWindowToScratchpad", "window", window)

	sourceMonitorID, err := ResolveSourceMonitorForWorkspace(
		a.aerospace,
		window.Workspace,
	)
	if err != nil {
		logger.LogError(
			"MOVER: unable to determine source monitor",
			"window",
			window,
			"error",
			err,
		)
		return "", fmt.Errorf(
			"unable to determine source monitor for window '%+v': %w",
			window,
			err,
		)
	}

	return a.MoveWindowToScratchpadForMonitor(window, sourceMonitorID)
}

func (a *MoverAeroSpace) MoveWindowToScratchpadForMonitor(
	window windows.Window,
	monitorID int,
) (string, error) {
	logger := logger.GetDefaultLogger()
	logger.LogDebug(
		"MOVING: MoveWindowToScratchpadForMonitor",
		"window",
		window,
		"monitorID",
		monitorID,
	)

	if monitorID <= 0 {
		return "", fmt.Errorf(
			"source monitor for window '%+v' is unknown; refusing to move to scratchpad",
			window,
		)
	}

	targetWorkspace, resolveErr := ResolveScratchpadWorkspaceNameForMonitor(
		a.aerospace,
		monitorID,
	)
	if resolveErr != nil {
		logger.LogError(
			"MOVER: unable to resolve scratchpad workspace for monitor",
			"monitorID",
			monitorID,
			"error",
			resolveErr,
		)
		return "", fmt.Errorf(
			"unable to resolve scratchpad workspace for monitor %d: %w",
			monitorID,
			resolveErr,
		)
	}

	provisioned, prevFocused, affinityErr := a.ensureScratchpadTarget(
		window,
		targetWorkspace,
		monitorID,
	)
	if affinityErr != nil {
		logger.LogError(
			"MOVER: scratchpad target is not attached to the window's source monitor",
			"window",
			window,
			"targetWorkspace",
			targetWorkspace,
			"monitorID",
			monitorID,
			"error",
			affinityErr,
		)
		return targetWorkspace, affinityErr
	}

	if err := a.moveWindowToScratchpadWorkspace(window, targetWorkspace); err != nil {
		return targetWorkspace, err
	}

	if provisioned {
		a.restoreSourceMonitorAndFocus(window, prevFocused)
	}
	return targetWorkspace, nil
}

// ensureScratchpadTarget guarantees the invariant: the window must never cross
// monitors. An existing target must be attached to the source monitor. An
// absent target is provisioned on the source monitor (multi-monitor) or left
// to implicit creation on the single move (proven single-monitor setups).
// This is a check-then-act guard over separate IPC calls: AeroSpace state can
// change between validation, provisioning, and move, so the invariant is
// enforced best-effort, not race-free.
func (a *MoverAeroSpace) ensureScratchpadTarget(
	window windows.Window,
	targetWorkspace string,
	sourceMonitorID int,
) (bool, *windows.Window, error) {
	workspaces, err := ListWorkspacesWithMonitors(a.aerospace)
	if err != nil {
		return false, nil, fmt.Errorf(
			"unable to verify scratchpad target '%s' attachment: %w",
			targetWorkspace,
			err,
		)
	}

	for _, workspaceMonitor := range workspaces {
		if workspaceMonitor.Workspace != targetWorkspace {
			continue
		}
		if workspaceMonitor.MonitorID != sourceMonitorID {
			return false, nil, fmt.Errorf(
				"scratchpad workspace '%s' is attached to monitor %d, but the window source monitor is %d; refusing cross-monitor move",
				targetWorkspace,
				workspaceMonitor.MonitorID,
				sourceMonitorID,
			)
		}
		return false, nil, nil
	}

	// Target does not exist yet. On single-monitor setups there is no other
	// monitor to cross, so the move itself can create it (compatibility case).
	// On multi-monitor setups, provision it explicitly on the source monitor:
	// where AeroSpace places a newly created workspace cannot be verified
	// atomically, so we focus the source window first and verify the result.
	if countUniqueMonitors(workspaces) == 1 {
		return false, nil, nil
	}

	prevFocused, provisionErr := a.provisionScratchpadOnSourceMonitor(
		window,
		targetWorkspace,
		sourceMonitorID,
	)
	return true, prevFocused, provisionErr
}

// provisionScratchpadOnSourceMonitor creates an absent scratchpad on the
// window's source monitor. It returns the previously focused window so the
// caller can restore focus state after the move (restoreSourceMonitorAndFocus).
func (a *MoverAeroSpace) provisionScratchpadOnSourceMonitor(
	window windows.Window,
	targetWorkspace string,
	sourceMonitorID int,
) (*windows.Window, error) {
	wrapper, isWrapper := a.aerospace.(*AeroSpaceClient)
	if isWrapper && wrapper.IsDryRun() {
		// Dry-run simulates provisioning without state, so the placement
		// verification below would fail against a workspace that was never
		// created. Write steps still log through the wrapper.
		if err := a.focusWindowForProvisioning(window); err != nil {
			return nil, err
		}
		if err := wrapper.SummonWorkspace(targetWorkspace); err != nil {
			return nil, fmt.Errorf(
				"unable to provision scratchpad workspace '%s': %w",
				targetWorkspace,
				err,
			)
		}
		if err := wrapper.WorkspaceBackAndForth(); err != nil {
			return nil, err
		}
		// Dry-run captures no focus to restore, so nil prevFocused is correct.
		return nil, nil //nolint:nilnil // no captured window in dry-run
	}

	prevFocused, err := a.aerospace.Windows().GetFocusedWindow()
	if err != nil {
		return nil, fmt.Errorf(
			"unable to capture focused window before provisioning '%s': %w",
			targetWorkspace,
			err,
		)
	}

	// Focus a window on the source monitor so the summoned workspace is
	// created there.
	if focusErr := a.focusWindowForProvisioning(window); focusErr != nil {
		a.restoreFocus(prevFocused)
		return nil, fmt.Errorf(
			"unable to focus source window for provisioning '%s': %w",
			targetWorkspace,
			focusErr,
		)
	}

	// The target is known to be absent, so this creates an empty workspace on
	// the focused (source) monitor; never summon an existing name.
	if summonErr := a.summonWorkspace(targetWorkspace); summonErr != nil {
		a.restoreFocus(prevFocused)
		return nil, fmt.Errorf(
			"unable to provision scratchpad workspace '%s': %w",
			targetWorkspace,
			summonErr,
		)
	}

	// Verify the provisioned workspace actually landed on the source monitor.
	workspaces, verifyErr := ListWorkspacesWithMonitors(a.aerospace)
	verifyErr = a.verifyProvisionedPlacement(
		workspaces,
		targetWorkspace,
		sourceMonitorID,
		verifyErr,
	)
	if verifyErr != nil {
		a.restoreFocus(prevFocused)
		return prevFocused, fmt.Errorf(
			"unable to provision scratchpad workspace '%s': %w",
			targetWorkspace,
			verifyErr,
		)
	}

	return prevFocused, nil
}

// verifyProvisionedPlacement checks that the provisioned workspace exists and
// is attached to the source monitor.
func (a *MoverAeroSpace) verifyProvisionedPlacement(
	workspaces []WorkspaceMonitor,
	targetWorkspace string,
	sourceMonitorID int,
	priorErr error,
) error {
	if priorErr != nil {
		return fmt.Errorf(
			"unable to verify provisioned workspace placement: %w",
			priorErr,
		)
	}

	for _, workspaceMonitor := range workspaces {
		if workspaceMonitor.Workspace != targetWorkspace {
			continue
		}
		if workspaceMonitor.MonitorID != sourceMonitorID {
			return fmt.Errorf(
				"provisioned workspace '%s' landed on monitor %d, not %d",
				targetWorkspace,
				workspaceMonitor.MonitorID,
				sourceMonitorID,
			)
		}
		return nil
	}
	return fmt.Errorf(
		"provisioned workspace '%s' was not found in the monitor mapping",
		targetWorkspace,
	)
}

// restoreSourceMonitorAndFocus runs after a successful provisioned move: it
// returns the source monitor's visible workspace to the pre-summon one and
// gives focus back to the user's window unless it is the moved window.
func (a *MoverAeroSpace) restoreSourceMonitorAndFocus(
	window windows.Window,
	prevFocused *windows.Window,
) {
	logger := logger.GetDefaultLogger()

	// Leave the empty scratchpad: return the source monitor's visible
	// workspace to the one focused before the summon.
	if err := a.workspaceBackAndForth(); err != nil {
		logger.LogError(
			"MOVER: unable to restore source monitor workspace after provisioning",
			"error",
			err,
		)
	}

	// Refocusing the moved window would expose it inside the scratchpad.
	if prevFocused != nil && prevFocused.WindowID != window.WindowID {
		a.restoreFocus(prevFocused)
	}
}

func (a *MoverAeroSpace) focusWindowForProvisioning(window windows.Window) error {
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		return wrapper.SetFocusByWindowID(window.WindowID)
	}
	return a.aerospace.Focus().SetFocusByWindowID(window.WindowID)
}

func (a *MoverAeroSpace) summonWorkspace(name string) error {
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		return wrapper.SummonWorkspace(name)
	}
	response, err := a.aerospace.Connection().SendCommand(
		"summon-workspace",
		[]string{name},
	)
	if err != nil {
		return fmt.Errorf("unable to summon workspace '%s': %w", name, err)
	}
	if response.ExitCode != 0 {
		return fmt.Errorf(
			"unable to summon workspace '%s': %s",
			name,
			response.StdErr,
		)
	}
	return nil
}

func (a *MoverAeroSpace) workspaceBackAndForth() error {
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		return wrapper.WorkspaceBackAndForth()
	}
	return a.aerospace.Workspaces().MoveBackAndForth()
}

func (a *MoverAeroSpace) restoreFocus(prevFocused *windows.Window) {
	if prevFocused == nil {
		return
	}
	if err := a.focusWindowForProvisioning(*prevFocused); err != nil {
		logger.GetDefaultLogger().LogError(
			"MOVER: unable to restore focus",
			"window",
			prevFocused,
			"error",
			err,
		)
	}
}

func (a *MoverAeroSpace) MoveWindowToWorkspace(
	window *windows.Window,
	workspace *workspaces.Workspace,
	shouldSetFocus bool,
) error {
	if window == nil {
		return errors.New("window is nil")
	}
	if workspace == nil {
		return errors.New("workspace is nil")
	}

	// Use wrapper's MoveWindowToWorkspace if available (for dry-run support)
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		if err := wrapper.MoveWindowToWorkspace(
			window.WindowID,
			workspace.Workspace,
		); err != nil {
			return fmt.Errorf(
				"unable to move window '%+v' to workspace '%s': %w",
				window,
				workspace.Workspace,
				err,
			)
		}
	} else {
		// Fallback to direct service call
		windowID := window.WindowID
		if err := a.aerospace.Workspaces().MoveWindowToWorkspaceWithOpts(
			workspaces.MoveWindowToWorkspaceArgs{
				WorkspaceName: workspace.Workspace,
			},
			workspaces.MoveWindowToWorkspaceOpts{
				WindowID: &windowID,
			},
		); err != nil {
			return fmt.Errorf(
				"unable to move window '%+v' to workspace '%s': %w",
				window,
				workspace.Workspace,
				err,
			)
		}
	}

	if !shouldSetFocus {
		return nil
	}

	// Use wrapper's SetFocusByWindowID if available (for dry-run support)
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		if err := wrapper.SetFocusByWindowID(window.WindowID); err != nil {
			return fmt.Errorf(
				"unable to set focus to window '%+v': %w",
				window,
				err,
			)
		}
	} else {
		// Fallback to direct service call
		if err := a.aerospace.Focus().SetFocusByWindowID(window.WindowID); err != nil {
			return fmt.Errorf(
				"unable to set focus to window '%+v': %w",
				window,
				err,
			)
		}
	}

	return nil
}

func (a *MoverAeroSpace) moveWindowToScratchpadWorkspace(
	window windows.Window,
	targetWorkspace string,
) error {
	logger := logger.GetDefaultLogger()
	// Use wrapper's MoveWindowToWorkspace if available (for dry-run support)
	var err error
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		err = wrapper.MoveWindowToWorkspace(
			window.WindowID,
			targetWorkspace,
		)
	} else {
		windowID := window.WindowID
		err = a.aerospace.Workspaces().MoveWindowToWorkspaceWithOpts(
			workspaces.MoveWindowToWorkspaceArgs{
				WorkspaceName: targetWorkspace,
			},
			workspaces.MoveWindowToWorkspaceOpts{
				WindowID: &windowID,
			},
		)
	}
	logger.LogDebug(
		"MOVING: after MoveWindowToWorkspace",
		"window", window,
		"to-workspace", targetWorkspace,
		"error", err,
	)
	if err != nil {
		return err
	}

	// Use wrapper's SetLayout if available (for dry-run support)
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		err = wrapper.SetLayout(window.WindowID, floatingLayout)
	} else {
		err = a.aerospace.Layout().SetLayout([]string{floatingLayout}, layout.SetLayoutOpts{
			WindowID: layout.IntPtr(window.WindowID),
		})
	}
	if err != nil {
		logger.LogDebug(
			"MOVER: unable to set layout to floating",
			"window", window,
			"error", err,
		)
	}
	return nil
}
