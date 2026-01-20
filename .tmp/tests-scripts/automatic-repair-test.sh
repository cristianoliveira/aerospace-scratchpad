#!/bin/bash
#
# Manual test script for automatic repair functionality of per-monitor scratchpad workspaces.
# This script verifies that mismatched scratchpad workspaces are automatically repaired
# when detected by ResolveScratchpadWorkspaceNameForMonitor.
#
# The script:
# 1. Sets up debug logging to see repair actions
# 2. Creates mismatched scratchpad workspaces (e.g., .scratchpad.1 attached to monitor 2)
# 3. Triggers automatic repair by calling aerospace-scratchpad move
# 4. Verifies repair happened (workspace moved to correct monitor or renamed)
# 5. Tests edge cases (orphaned monitors, default .scratchpad workspace)
#
# Requirements:
# - AeroSpace WM must be running
# - aerospace-scratchpad CLI must be installed (make install)
# - At least one monitor (multi-monitor setup for full testing)
# - A window to manipulate (Finder window used by default)
#
# Usage:
#   cd /path/to/aerospace-scratchpad
#   chmod +x .tmp/tests-scripts/automatic-repair-test.sh
#   .tmp/tests-scripts/automatic-repair-test.sh [--yes] [--app APP_NAME]
#
# Options:
#   --yes       Skip confirmation prompt
#   --app APP   Use windows from specified application (default: Finder)
#
# The script is safe to run: it manipulates a single window (by default Finder)
# and cleans up after itself.

set -euo pipefail
cd "$(dirname "$0")/../.."  # Move to project root

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Parse command line arguments
SKIP_CONFIRM=false
TEST_APP="Finder"
while [[ $# -gt 0 ]]; do
    case $1 in
        --yes)
            SKIP_CONFIRM=true
            shift
            ;;
        --app)
            TEST_APP="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [--yes] [--app APP_NAME]"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Enable debug logging for aerospace-scratchpad
export AEROSPACE_SCRATCHPAD_LOGS_LEVEL=DEBUG
export AEROSPACE_SCRATCHPAD_LOGS_PATH="/tmp/aerospace-scratchpad-repair-test.log"

# Clear previous log file
> "$AEROSPACE_SCRATCHPAD_LOGS_PATH"

# Check prerequisites
command -v aerospace >/dev/null 2>&1 || { log_error "aerospace CLI not found. Is AeroSpace WM installed?"; exit 1; }
command -v aerospace-scratchpad >/dev/null 2>&1 || { log_error "aerospace-scratchpad CLI not found. Run 'make install'."; exit 1; }
command -v jq >/dev/null 2>&1 || { log_error "jq not found. Please install jq (brew install jq)."; exit 1; }

# Safety confirmation
if [[ "$SKIP_CONFIRM" = false ]]; then
    echo "This test will manipulate a '$TEST_APP' window:"
    echo "  - Move it to scratchpad workspaces"
    echo "  - Rename and move workspaces"
    echo "  - Finally restore it to its original workspace"
    echo ""
    echo "Make sure you have a '$TEST_APP' window open and you're okay with it being moved temporarily."
    echo -n "Proceed? (y/N) "
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_info "Aborted by user."
        exit 0
    fi
fi

# Get monitor information
log_info "Fetching monitor information..."
MONITOR_LIST=$(aerospace list-monitors --json)
MONITOR_IDS=$(echo "$MONITOR_LIST" | jq -r '.[].monitorID')
MONITOR_COUNT=$(echo "$MONITOR_IDS" | wc -l | tr -d ' ')

log_info "Found $MONITOR_COUNT monitor(s): $(echo "$MONITOR_IDS" | tr '\n' ', ')"

if [[ $MONITOR_COUNT -eq 0 ]]; then
    log_error "No monitors detected. Cannot proceed."
    exit 1
fi

# Pick a window to manipulate
WINDOW_APP="$TEST_APP"
WINDOW_ID=$(aerospace list-windows --all --format '%{window-id} | %{app-name}' | grep "$WINDOW_APP" | head -1 | awk '{print $1}')
if [[ -z "$WINDOW_ID" ]]; then
    log_error "No '$WINDOW_APP' window found. Please open a window or specify a different app with --app."
    exit 1
fi
log_info "Using window ID $WINDOW_ID ($WINDOW_APP) for testing"

# Function to get workspace of a window
get_window_workspace() {
    aerospace list-windows --all --format '%{window-id} | %{workspace}' | grep "^$1" | awk '{print $3}'
}

# Function to get monitor of a workspace
get_workspace_monitor() {
    aerospace list-workspaces --all --format '%{workspace} | %{monitor-id}' | grep "^$1" | awk '{print $3}'
}

# Function to move window to a specific monitor
move_window_to_monitor() {
    local monitor_id="$1"
    log_info "Moving window to monitor $monitor_id"
    aerospace move-node-to-monitor "$monitor_id" --window-id "$WINDOW_ID"
    sleep 0.5
}

# Capture initial workspace for restoration
INITIAL_WORKSPACE=$(get_window_workspace "$WINDOW_ID")
log_info "Initial window workspace: $INITIAL_WORKSPACE"

# Global cleanup function
final_cleanup() {
    log_info "Restoring window to original workspace..."
    aerospace move-node-to-workspace "$INITIAL_WORKSPACE" --window-id "$WINDOW_ID" 2>/dev/null || true
}
trap final_cleanup EXIT

# Function to create a mismatched scratchpad workspace
# $1: desired workspace name (e.g., .scratchpad.1)
# $2: monitor ID to attach it to (must exist)
create_mismatched_workspace() {
    local workspace_name="$1"
    local target_monitor="$2"
    log_info "Creating mismatched workspace: $workspace_name attached to monitor $target_monitor"
    
    # Move window to target monitor first
    move_window_to_monitor "$target_monitor"
    
    # Move window to scratchpad (will create workspace on target monitor)
    aerospace-scratchpad move "$WINDOW_APP"
    sleep 0.5
    
    # Get current workspace name (should be .scratchpad or .scratchpad.<monitor>)
    local current_workspace
    current_workspace=$(get_window_workspace "$WINDOW_ID")
    log_info "Window moved to workspace: $current_workspace"
    
    # Rename workspace if needed
    if [[ "$current_workspace" != "$workspace_name" ]]; then
        aerospace rename-workspace "$current_workspace" "$workspace_name"
        sleep 0.5
    fi
    
    # Move workspace to target monitor (if not already there)
    local current_monitor
    current_monitor=$(get_workspace_monitor "$workspace_name" 2>/dev/null || echo "")
    if [[ "$current_monitor" != "$target_monitor" ]]; then
        aerospace move-workspace-to-monitor "$workspace_name" "$target_monitor"
        sleep 0.5
    fi
    
    log_info "Mismatched workspace created: $workspace_name on monitor $target_monitor"
}

# Function to trigger repair by moving a window to scratchpad
trigger_repair() {
    log_info "Triggering automatic repair via aerospace-scratchpad move..."
    aerospace-scratchpad move "$WINDOW_APP"
    sleep 0.5
}

# Function to verify workspace is correctly attached
# $1: expected workspace name pattern (can be regex)
# $2: expected monitor ID (optional, if empty just check name)
verify_workspace() {
    local expected_pattern="$1"
    local expected_monitor="$2"
    
    local current_workspace
    current_workspace=$(get_window_workspace "$WINDOW_ID")
    
    if [[ ! "$current_workspace" =~ $expected_pattern ]]; then
        log_error "Workspace mismatch. Expected pattern: $expected_pattern, actual: $current_workspace"
        return 1
    fi
    
    if [[ -n "$expected_monitor" ]]; then
        local current_monitor
        current_monitor=$(get_workspace_monitor "$current_workspace")
        if [[ "$current_monitor" != "$expected_monitor" ]]; then
            log_error "Monitor mismatch. Expected: $expected_monitor, actual: $current_monitor"
            return 1
        fi
    fi
    
    log_info "Verified workspace: $current_workspace (monitor ${current_monitor:-N/A}) matches expectations"
    return 0
}

# Function to clean up test workspace
cleanup() {
    log_info "Cleaning up test workspace..."
    # Move window out of scratchpad back to a regular workspace
    aerospace move-node-to-workspace 1 --window-id "$WINDOW_ID" 2>/dev/null || true
    # Workspace will be automatically deleted when empty
    sleep 0.5
}

# Main test execution
main() {
    log_info "Starting automatic repair test..."
    
    
    # Ensure we start with no scratchpad windows for our test window
    cleanup
    
    # Test 1: Default .scratchpad workspace (should NOT be repaired)
    log_info "=== Test 1: Default .scratchpad workspace (no repair expected) ==="
    create_mismatched_workspace ".scratchpad" "1"
    trigger_repair
    # Default .scratchpad should stay as is
    verify_workspace "^\.scratchpad$" "1"
    log_info "Test 1 passed: default .scratchpad workspace not repaired (as designed)"
    cleanup
    
    if [[ $MONITOR_COUNT -ge 2 ]]; then
        # Test 2: Mismatched workspace where expected monitor exists
        # Create .scratchpad.1 attached to monitor 2, monitor 1 exists
        log_info "=== Test 2: Mismatched workspace with expected monitor existing ==="
        create_mismatched_workspace ".scratchpad.1" "2"
        trigger_repair
        # Should move workspace to monitor 1 (expected monitor)
        verify_workspace "^\.scratchpad\.1$" "1"
        log_info "Test 2 passed: workspace moved to correct monitor"
        cleanup
        
        # Test 3: Mismatched workspace where expected monitor does NOT exist
        # Create .scratchpad.5 attached to monitor 2, monitor 5 doesn't exist
        log_info "=== Test 3: Mismatched workspace with orphaned expected monitor ==="
        create_mismatched_workspace ".scratchpad.5" "2"
        trigger_repair
        # Should rename workspace to match current monitor (monitor 2)
        # Expected name depends on monitor count
        local expected_name
        if [[ $MONITOR_COUNT -eq 1 ]]; then
            expected_name="^\\.scratchpad$"
        else
            expected_name="^\\.scratchpad\\.2$"
        fi
        verify_workspace "$expected_name" "2"
        log_info "Test 3 passed: workspace renamed to match current monitor"
        cleanup
    else
        log_warn "Only one monitor detected. Skipping multi-monitor tests 2 and 3."
    fi
    
    log_info "Test completed successfully!"
    
    # Show repair log snippets
    log_info "=== Debug log snippets ==="
    grep -E "(repair|mismatch|moving|renaming)" "$AEROSPACE_SCRATCHPAD_LOGS_PATH" | head -20 || true
    log_info "Full log available at $AEROSPACE_SCRATCHPAD_LOGS_PATH"
}

# Run main function and catch errors
if main; then
    log_info "All tests passed!"
    exit 0
else
    log_error "Test failed!"
    exit 1
fi