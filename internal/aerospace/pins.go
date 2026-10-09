package aerospace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
)

const pinsPathEnv = "AEROSPACE_SCRATCHPAD_PINS_PATH"

type Pin struct {
	MonitorID int `json:"monitor_id"`
}

type PinRule struct {
	Pattern   string   `json:"pattern"`
	Filters   []string `json:"filters,omitempty"`
	MonitorID int      `json:"monitor_id"`
	Enabled   *bool    `json:"enabled,omitempty"`
}

type PinsState struct {
	Pinned map[int]Pin `json:"pinned"`
	Rules  []PinRule   `json:"rules,omitempty"`
}

func pinsStatePath() string {
	if path := os.Getenv(pinsPathEnv); path != "" {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "aerospace-scratchpad-pins.json")
	}
	return filepath.Join(homeDir, ".config", "aerospace-scratchpad", "pinned.json")
}

func LoadPins() (PinsState, error) {
	path := pinsStatePath()
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return PinsState{Pinned: map[int]Pin{}}, nil
	}
	if err != nil {
		return PinsState{}, fmt.Errorf("unable to read pins state: %w", err)
	}

	var state PinsState
	if unmarshalErr := json.Unmarshal(contents, &state); unmarshalErr != nil {
		return PinsState{}, fmt.Errorf("unable to parse pins state: %w", unmarshalErr)
	}
	if state.Pinned == nil {
		state.Pinned = map[int]Pin{}
	}
	return state, nil
}

func SavePins(state PinsState) error {
	path := pinsStatePath()
	if mkdirErr := os.MkdirAll(filepath.Dir(path), 0o750); mkdirErr != nil {
		return fmt.Errorf("unable to create pins state directory: %w", mkdirErr)
	}
	if state.Pinned == nil {
		state.Pinned = map[int]Pin{}
	}
	contents, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to encode pins state: %w", err)
	}
	if writeErr := os.WriteFile(path, contents, 0o600); writeErr != nil {
		return fmt.Errorf("unable to write pins state: %w", writeErr)
	}
	return nil
}

func PinWindow(windowID int, monitorID int) error {
	state, err := LoadPins()
	if err != nil {
		return err
	}
	state.Pinned[windowID] = Pin{MonitorID: monitorID}
	return SavePins(state)
}

func PinRuleForPattern(pattern string, filters []string, monitorID int) error {
	state, err := LoadPins()
	if err != nil {
		return err
	}
	enabled := true
	state.Rules = upsertPinRule(state.Rules, PinRule{
		Pattern:   pattern,
		Filters:   filters,
		MonitorID: monitorID,
		Enabled:   &enabled,
	})
	return SavePins(state)
}

func upsertPinRule(rules []PinRule, rule PinRule) []PinRule {
	for index, existingRule := range rules {
		if existingRule.Pattern == rule.Pattern && sameStrings(existingRule.Filters, rule.Filters) {
			rules[index] = rule
			return rules
		}
	}
	return append(rules, rule)
}

func sameStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func UnpinWindow(windowID int) error {
	state, err := LoadPins()
	if err != nil {
		return err
	}
	delete(state.Pinned, windowID)
	return SavePins(state)
}

func UnpinRuleForPattern(pattern string, filters []string) error {
	state, err := LoadPins()
	if err != nil {
		return err
	}

	var remainingRules []PinRule
	for _, rule := range state.Rules {
		if rule.Pattern == pattern && sameStrings(rule.Filters, filters) {
			continue
		}
		remainingRules = append(remainingRules, rule)
	}
	state.Rules = remainingRules
	return SavePins(state)
}

func SetPinRuleEnabled(pattern string, filters []string, enabled bool) error {
	state, err := LoadPins()
	if err != nil {
		return err
	}

	for index, rule := range state.Rules {
		if rule.Pattern == pattern && sameStrings(rule.Filters, filters) {
			state.Rules[index].Enabled = &enabled
		}
	}
	return SavePins(state)
}

func PinnedMonitorID(windowID int) (int, bool, error) {
	state, err := LoadPins()
	if err != nil {
		return 0, false, err
	}
	pin, ok := state.Pinned[windowID]
	return pin.MonitorID, ok, nil
}

func PinnedMonitorIDForWindow(window windows.Window) (int, bool, error) {
	state, err := LoadPins()
	if err != nil {
		return 0, false, err
	}
	if pin, ok := state.Pinned[window.WindowID]; ok {
		return pin.MonitorID, true, nil
	}

	for _, rule := range state.Rules {
		if !pinRuleEnabled(rule) {
			continue
		}
		matches, matchErr := pinRuleMatchesWindow(rule, window)
		if matchErr != nil {
			return 0, false, matchErr
		}
		if matches {
			return rule.MonitorID, true, nil
		}
	}
	return 0, false, nil
}

func pinRuleEnabled(rule PinRule) bool {
	return rule.Enabled == nil || *rule.Enabled
}

func pinRuleMatchesWindow(rule PinRule, window windows.Window) (bool, error) {
	pattern, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return false, fmt.Errorf("invalid pin pattern '%s': %w", rule.Pattern, err)
	}
	if !pattern.MatchString(window.AppName) {
		return false, nil
	}

	filters, err := ParseFilters(rule.Filters)
	if err != nil {
		return false, err
	}
	return ApplyFilters(window, filters)
}

func ResolveWindowMonitorID(cli AeroSpaceWMClient, window windows.Window) (int, error) {
	workspaces, err := ListWorkspacesWithMonitors(cli)
	if err == nil {
		for _, workspace := range workspaces {
			if workspace.Workspace == window.Workspace {
				return workspace.MonitorID, nil
			}
		}
	}

	monitor, monitorErr := GetFocusedMonitor(cli)
	if monitorErr != nil {
		if err != nil {
			return 0, fmt.Errorf("unable to resolve window monitor: %w", err)
		}
		return 0, monitorErr
	}
	return monitor.MonitorID, nil
}
