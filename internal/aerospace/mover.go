package aerospace

import (
	"errors"
	"fmt"
	"strconv"

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

	provisioning, affinityErr := a.ensureScratchpadTarget(
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
		// Provisioning changed focus and the source monitor's active
		// workspace; restore them even though the move failed. Restoration
		// errors are logged inside; the move error takes precedence.
		if provisioning != nil {
			_ = a.restoreAfterProvisionedMove(provisioning)
		}
		return targetWorkspace, err
	}

	if provisioning != nil {
		if restoreErr := a.restoreAfterProvisionedMove(provisioning); restoreErr != nil {
			return targetWorkspace, fmt.Errorf(
				"window moved to '%s' but focus restoration failed; check your visible workspaces: %w",
				targetWorkspace,
				restoreErr,
			)
		}
	}
	return targetWorkspace, nil
}

// provisioningState carries the focus state captured before provisioning so
// it can be restored after the move.
type provisioningState struct {
	prevFocused           *windows.Window
	movedWindowID         int
	sourceActiveWorkspace string
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
) (*provisioningState, error) {
	workspaces, err := ListWorkspacesWithMonitors(a.aerospace)
	if err != nil {
		return nil, fmt.Errorf(
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
			return nil, fmt.Errorf(
				"scratchpad workspace '%s' is attached to monitor %d, but the window source monitor is %d; refusing cross-monitor move",
				targetWorkspace,
				workspaceMonitor.MonitorID,
				sourceMonitorID,
			)
		}
		return nil, nil //nolint:nilnil // target exists, no provisioning needed
	}

	// Target does not exist yet. On single-monitor setups there is no other
	// monitor to cross, so the move itself can create it (compatibility case).
	// On multi-monitor setups, provision it explicitly on the source monitor:
	// where AeroSpace places a newly created workspace cannot be verified
	// atomically, so we focus the source monitor, capture its active
	// workspace, summon, and verify the result.
	if countUniqueMonitors(workspaces) == 1 {
		return nil, nil //nolint:nilnil // single-monitor, no cross risk
	}

	state, provisionErr := a.provisionScratchpadOnSourceMonitor(
		window,
		targetWorkspace,
		sourceMonitorID,
	)
	return state, provisionErr
}

// provisionScratchpadOnSourceMonitor creates an absent scratchpad on the
// source monitor following the source-backed sequence: focus-monitor the
// source, capture its active workspace, summon-workspace the absent name,
// and verify the attachment. It returns the state needed to restore focus
// after the move (restoreAfterProvisionedMove). Every failure path restores
// what changed so far and leaves the window where it is.
func (a *MoverAeroSpace) provisionScratchpadOnSourceMonitor(
	window windows.Window,
	targetWorkspace string,
	sourceMonitorID int,
) (*provisioningState, error) {
	wrapper, isWrapper := a.aerospace.(*AeroSpaceClient)
	if isWrapper && wrapper.IsDryRun() {
		// Dry-run simulates provisioning without state, so the placement
		// verification below would fail against a workspace that was never
		// created. Write steps still log through the wrapper.
		if err := wrapper.FocusMonitor(sourceMonitorID); err != nil {
			return nil, err
		}
		if err := wrapper.SummonWorkspace(targetWorkspace); err != nil {
			return nil, fmt.Errorf(
				"unable to provision scratchpad workspace '%s': %w",
				targetWorkspace,
				err,
			)
		}
		// Dry-run captures no focus state to restore; movedWindowID is set
		// solely so the caller can skip refocusing the moved window.
		return &provisioningState{
			movedWindowID: window.WindowID,
		}, nil
	}

	prevFocused, err := a.aerospace.Windows().GetFocusedWindow()
	if err != nil {
		return nil, fmt.Errorf(
			"unable to capture focused window before provisioning '%s': %w",
			targetWorkspace,
			err,
		)
	}

	// Make the source monitor the focused one so the summoned workspace is
	// created there, then capture its currently active workspace.
	if focusErr := a.focusMonitor(sourceMonitorID); focusErr != nil {
		a.restoreFocus(prevFocused)
		return nil, fmt.Errorf(
			"unable to focus source monitor for provisioning '%s': %w",
			targetWorkspace,
			focusErr,
		)
	}

	sourceActive := a.sourceActiveWorkspace()
	if sourceActive == "" {
		a.restoreFocus(prevFocused)
		return nil, fmt.Errorf(
			"unable to capture active workspace of monitor %d for provisioning '%s'",
			sourceMonitorID,
			targetWorkspace,
		)
	}

	// The target is known to be absent, so this creates an empty workspace on
	// the focused (source) monitor; never summon an existing name.
	if summonErr := a.summonWorkspace(targetWorkspace); summonErr != nil {
		a.restoreAfterProvisioningFailure(sourceActive, prevFocused)
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
		a.restoreAfterProvisioningFailure(sourceActive, prevFocused)
		return nil, fmt.Errorf(
			"unable to provision scratchpad workspace '%s': %w",
			targetWorkspace,
			verifyErr,
		)
	}

	return &provisioningState{
		prevFocused:           prevFocused,
		movedWindowID:         window.WindowID,
		sourceActiveWorkspace: sourceActive,
	}, nil
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

// restoreAfterProvisioningFailure returns the source monitor to its captured
// active workspace and the user to their focused window. Best-effort; failures
// are logged.
func (a *MoverAeroSpace) restoreAfterProvisioningFailure(
	sourceActiveWorkspace string,
	prevFocused *windows.Window,
) {
	logger := logger.GetDefaultLogger()

	if sourceActiveWorkspace != "" {
		if err := a.switchToWorkspace(sourceActiveWorkspace); err != nil {
			logger.LogError(
				"MOVER: unable to restore source monitor workspace after failed provisioning",
				"error",
				err,
			)
		}
	}
	a.restoreFocus(prevFocused)
}

// restoreAfterProvisionedMove runs after a provisioned move: it returns the
// source monitor's active workspace to the captured one and gives focus back
// to the user's window unless it is the moved window (refocusing it would
// expose it inside the scratchpad). Both steps are attempted; failures are
// joined and surfaced to the caller.
func (a *MoverAeroSpace) restoreAfterProvisionedMove(
	state *provisioningState,
) error {
	logger := logger.GetDefaultLogger()

	var restoreErrs []error

	if state.sourceActiveWorkspace != "" {
		if err := a.switchToWorkspace(state.sourceActiveWorkspace); err != nil {
			logger.LogError(
				"MOVER: unable to restore source monitor workspace after provisioning",
				"error",
				err,
			)
			restoreErrs = append(restoreErrs, fmt.Errorf(
				"unable to restore the source monitor workspace: %w",
				err,
			))
		}
	}

	// Refocusing the moved window would expose it inside the scratchpad.
	if state.prevFocused != nil && state.prevFocused.WindowID != state.movedWindowID {
		if err := a.focusWindow(state.prevFocused.WindowID); err != nil {
			logger.LogError(
				"MOVER: unable to restore focus after provisioning",
				"window",
				state.prevFocused,
				"error",
				err,
			)
			restoreErrs = append(restoreErrs, fmt.Errorf(
				"unable to restore focus to window %d: %w",
				state.prevFocused.WindowID,
				err,
			))
		}
	}

	return errors.Join(restoreErrs...)
}

// sourceActiveWorkspace returns the globally focused workspace, which after
// focus-monitor is the source monitor's active workspace. Empty on error.
func (a *MoverAeroSpace) sourceActiveWorkspace() string {
	focused, err := a.aerospace.Workspaces().GetFocusedWorkspace()
	if err != nil || focused == nil {
		return ""
	}
	return focused.Workspace
}

func (a *MoverAeroSpace) focusMonitor(monitorID int) error {
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		return wrapper.FocusMonitor(monitorID)
	}
	response, err := a.aerospace.Connection().SendCommand(
		"focus-monitor",
		[]string{strconv.Itoa(monitorID)},
	)
	if err != nil {
		return fmt.Errorf("unable to focus monitor %d: %w", monitorID, err)
	}
	if response.ExitCode != 0 {
		return fmt.Errorf(
			"unable to focus monitor %d: %s",
			monitorID,
			response.StdErr,
		)
	}
	return nil
}

func (a *MoverAeroSpace) switchToWorkspace(name string) error {
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		return wrapper.SwitchWorkspace(name)
	}
	response, err := a.aerospace.Connection().SendCommand(
		"workspace",
		[]string{name},
	)
	if err != nil {
		return fmt.Errorf("unable to switch to workspace '%s': %w", name, err)
	}
	if response.ExitCode != 0 {
		return fmt.Errorf(
			"unable to switch to workspace '%s': %s",
			name,
			response.StdErr,
		)
	}
	return nil
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

// focusWindow focuses a window; used to restore the user's focus.
func (a *MoverAeroSpace) focusWindow(windowID int) error {
	if wrapper, ok := a.aerospace.(*AeroSpaceClient); ok {
		return wrapper.SetFocusByWindowID(windowID)
	}
	return a.aerospace.Focus().SetFocusByWindowID(windowID)
}

func (a *MoverAeroSpace) restoreFocus(prevFocused *windows.Window) {
	if prevFocused == nil {
		return
	}
	if err := a.focusWindow(prevFocused.WindowID); err != nil {
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
