# aerospace-scratchpad

Here you will find extensive documentation about the CLI.

## Command: `move`

Move the currently focused window to the scratchpad workspace (`.scratchpad` or `.scratchpad.<monitor-id>`). The window will be hidden until you show it again.
You can actually see this in your workspace list, but it can be ignored—it's just used to store windows that are "hidden".

When **sending a window into the scratchpad** (`move`, or the hide/toggle path in `show`), each window is routed to the scratchpad attached to **its own source monitor** — the monitor of the workspace the window currently lives in — never to the globally focused monitor. When a safe same-monitor scratchpad target cannot be established, the command fails for that window and leaves it where it is instead of moving it across monitors. This covers an unknown source monitor and a target scratchpad attached to another monitor. A missing target is provisioned automatically on the source monitor (empty workspace only); if provisioning or its placement verification fails, the send fails closed and the window stays. This never-cross constraint applies only to sending into the scratchpad: bringing a window out with `show`, `summon`, or `hook pull-window` may move it to another monitor by design. The check is performed just before the move over separate IPC calls and is not race-free.

### USAGE

`pattern` is a regex pattern to match the app name.

```bash
aerospace-scratchpad move <pattern>
```

For more details:
```bash
aerospace-scratchpad move --help
```

To move all windows that match the focused window's app name to the scratchpad, you can use:
```bash
aerospace-scratchpad move --all-matching 
```

To move all floating windows (scratchpad windows) to the scratchpad without requiring a pattern, you can use:
```bash
aerospace-scratchpad move --all-floating 
```

#### Option flags

##### All Matching `--all` (deprecated use `--all-matching`)

The `move` flag `--all` has been renamed to `--all-matching`. Update any configs, scripts, or keybindings that used `--all` to use `--all-matching` instead.

##### All Matching `--all-matching`

Move every window whose app name matches the focused window (or provided pattern) to the scratchpad in one command.

```bash
aerospace-scratchpad move --all-matching
```

### All Floating `--all-floating`

_min version: 0.5.0_

Move all floating windows (scratchpad windows) to the scratchpad workspace without requiring a pattern. This is useful when you want to hide all scratchpad windows at once.

```bash
aerospace-scratchpad move --all-floating
```

This command will:
- Find all windows with `WindowLayout == "floating"`
- Move each floating window to the scratchpad workspace attached to that window's own source monitor (`.scratchpad` or `.scratchpad.<monitor-id>`)
- Ensure they remain floating

See more [flags](#flags).

## Command: `show`

Similar to Sway's `show`, this command will:

 - Show a window that was previously moved to the scratchpad workspace.
 - Move the window to the scratchpad if it is focused and matches the `<pattern>`.
 - If a scratchpad window is in another workspace, it will move it to the current workspace.
 - If a scratchpad window is already in the current workspace, it will set focus on it.
 - If multiple windows match a pattern, it will bring all of them to the current workspace.
 - If no window matches a pattern, it will do nothing.

The `pattern` is a regex pattern to match the "App Name".

USAGE: `aerospace-scratchpad show <pattern>`

For more details:
```bash
aerospace-scratchpad show --help
```

See also [flags](#flags).

## Command: `summon`

Like `show`, this command requires a pattern. Unlike `show`, it only brings matching windows to the current workspace and focuses them; it does not toggle visible windows back to scratchpad.

Use `next` when you want to cycle through scratchpad windows without specifying a pattern.

### USAGE

The `pattern` is a regex pattern to match the "App Name".

```bash
aerospace-scratchpad summon <pattern>
```

See also [flags](#flags).

## Command: `next`

This command cycles through scratchpad windows across all monitors by default. Use `--monitor current`
or `--monitor <ID>` to restrict the candidate set. Windows are ordered by window ID. If the focused
window belongs to that set, `next` selects its successor and wraps after the last window. Without a
usable focus cursor, it prefers the first candidate outside the current workspace to avoid repeating
a no-op; if all candidates are already there, it selects the first one.

The selected window moves to the current workspace and receives focus. That focus acts as the cursor
for the next invocation, so cycling works across separate CLI processes without persisted state.
The candidate set uses the same scratchpad definition as `list`: windows in a scratchpad workspace
or floating windows.

### USAGE

```bash
# Cycle scratchpad windows across all monitors (default)
aerospace-scratchpad next

# Restrict cycling to the current monitor
aerospace-scratchpad next --monitor current

# Restrict cycling to monitor 2
aerospace-scratchpad next --monitor 2
```

## Command: `list` / `ls`

_Min version: 0.5.0_

List all scratchpad windows. A scratchpad window is defined as:
- A window in a scratchpad workspace (`.scratchpad` or `.scratchpad.<monitor-id>`), OR
- A floating window (WindowLayout == "floating")

The output is scriptable and supports multiple formats (text, json, tsv, csv).

### USAGE

```bash
# List scratchpad windows across all monitors (text format; default)
aerospace-scratchpad list

# Using the alias
aerospace-scratchpad ls

# List in JSON format for scripting
aerospace-scratchpad list --output json

# List with filters
aerospace-scratchpad list --filter app-name=^Terminal
```

See more [flags](#flags).

## Options flag

### Filter `--filter|-F <property>=<regex>` 

_min version: 0.2.0_

The filter flag helps to narrow down the windows that will be shown. It accepts a property and a regex pattern to match against that property. It can be used multiple time with different properties to narrow down the window matching.

For example, to filter by class and title, you can use:

```bash
aerospace-scratchpad show Brave -F window-title=Gmail -F window-title="personal"
# Bring all Brave windows with title containing "Gmail" AND "personal" to the current workspace.

aerospace-scratchpad show Terminal --filter window-title=kitty
# Bring all Terminal windows with title containing "kitty" to the current workspace.

aerospace-scratchpad show Kitty -F window-title='(?i)kitty.*work'
# Bring all windows with title matching the regex (Case insensitive) "kitty.*work" to the current workspace. Eg. "kitty work", "kitty work project", "KITTY more WORK", etc

## Example on how to use only window filter (We may allow empty patterns in the future)
aerospace-scratchpad show . --filter window-title=kitty
# Match all windows and filter the ones with title containing "kitty" bringing to the current workspace.
```

Current allowed properties for filtering are:

    - *window-id*: The ID of the window.
    - *window-title*: The title of the window. 
    - *app-name*: The name of the application. E.g. `Terminal`, `Brave`, etc.
    - *app-bundle-id*: The bundle ID of the application. E.g. `com.apple.Terminal`.

It fails if the property is not recognized or if the regex pattern is invalid.

For more advanced regex patterns check [Google re2 syntax](https://github.com/google/re2/wiki/Syntax)

### Dry Run `--dry-run|-n`

_min version: 0.2.0_

This flag will not execute the command, but will print what would be done. Very handy to test your command before adding to your
config file.

Usage:
```bash
aerospace-scratchpad --dry-run show <pattern>
```

It will print the actions that would be taken, but will not execute them.

### Output format `--output|-o`

_min version: 0.5.0_

Available on commands that emit structured results (`move`, `show`, `summon`, `next`). Choose between:
- `text` (default): single-line key=value pairs (quote-aware)
- `json`: one JSON object per line
- `tsv`: tab-separated with header
- `csv`: comma-separated with header

Examples:
- Text (default): `aerospace-scratchpad move --all-matching | rg 'result=ok'`
- JSON: `aerospace-scratchpad move --output=json | jq -r 'select(.result==\"ok\") | .window_id'`
- TSV: `aerospace-scratchpad move --output=tsv | awk 'NR>1 {print $3}'`  # 3rd column is window_id
- CSV: `aerospace-scratchpad show foo --output=csv | csvcut -c window_id`  # requires csvkit
- Next: `aerospace-scratchpad next --output=json | jq -r '.target_workspace'`

Fields (in order): `command action window_id app_name workspace target_workspace result message`

#### Scripting tips
- Filter successes: `aerospace-scratchpad move --output=text | rg 'result=ok'`
- Collect window IDs: `aerospace-scratchpad show chatgpt --output=json | jq -r 'select(.action==\"focus\") | .window_id'`
- Pipe to awk: `aerospace-scratchpad next --output=tsv | awk 'NR>1 {print $3}'` # window_id
- CSV tooling: `aerospace-scratchpad move --output=csv | csvcut -c window_id,app_name` (requires csvkit)

## Auxiliar Commands for integrations

### Command: `hook`

_min version: 0.3.0_

This is a set of subcommands to be used to handle specific events from AeroSpaceWM related to the scratchpad workspace.
Use this to enhance the UX of the scratchpad and the interaction with the macOS environment.

Usually these commands are used by hooking to AeroSpaceWM events using the `exec-on-*` hooks in your `aerospace.toml` config file.

### Command: `hook pull-window`

This subcommand handles when the scratchpad workspace gets focused, which shouldn't happen. It will move focus back to the last focused workspace and pull the focused window from scratchpad.
This allows you to use different tools to focus windows in scratchpad, like notifications, external launchers, etc., and behave as "summoning" the window to the current workspace instead of focusing the window in the scratchpad workspace.
For a deeper walkthrough (architecture, logging, troubleshooting) see [`docs/hook-integration.md`](./hook-integration.md).

#### USAGE

`aerospace-scratchpad hook pull-window <previous-workspace> <focused-workspace>`

Add this snippet in your `~/.aerospace.toml` config:
```toml
# Ensure when scratchpad windows take focus, they are summoned to the previous workspace instead.
exec-on-workspace-change = ["/bin/bash", "-c",
  "aerospace-scratchpad hook pull-window $AEROSPACE_PREV_WORKSPACE $AEROSPACE_FOCUSED_WORKSPACE"
]
```

For more details:
```bash
aerospace-scratchpad hook pull-window --help
```

## Implementation details

### Scratchpad workspace

It will send the window to a "special" workspace called `.scratchpad` (or `.scratchpad.<monitor-id>` for multi-monitor setups). This workspace is like any other workspace, but can be ignored. The window will be hidden until you show it again.

When you have multiple monitors, each monitor can have its own scratchpad workspace (e.g., `.scratchpad.1`, `.scratchpad.2`). When **sending a window into the scratchpad** with `move` or the hide/toggle path in `show`, the destination is the scratchpad attached to that window's source monitor (the monitor of the workspace it currently lives in), resolved independently per window. If a same-monitor destination cannot be verified — unknown source monitor, destination attached to another monitor — the send fails for that window and leaves it where it is. A destination that does not exist yet is provisioned automatically on the source monitor (see the provisioning flow in the Multi-Monitor Configuration section); if provisioning or its placement verification fails, the send fails closed and the window stays. This no-cross-monitor guarantee applies only when sending into the scratchpad; bringing a window out with `show`, `summon`, or `hook pull-window` may move it to a workspace on another monitor. Note the guard is check-then-move validation over separate IPC calls: it is enforced at validation time and is not race-free against concurrent workspace changes.

For single-monitor setups, the default `.scratchpad` workspace is used for backward compatibility.

#### Limitations

The same-monitor guard is a preflight check followed by a separate move command (two IPC calls). Workspaces can change in between — focus switches, `workspace-to-monitor-force-assignment` edits, or a workspace being removed — so affinity is best-effort against the state observed at validation time, not an atomic guarantee.

Provisioning an absent scratchpad temporarily switches focus (to the sending window's monitor, then back) and runs `summon-workspace` followed by a placement verification; provisioning and verification are separate IPC calls and are not race-free. If focus restoration fails, your focus may be left on the source monitor, or — in the failure case — the newly created empty scratchpad may remain the visible workspace there (AeroSpace has no workspace delete).

### Multi-Monitor Configuration

For optimal multi-monitor scratchpad experience:

1. **Prevent workspace drift**: Add to your `aerospace.toml`:
   ```toml
   workspace-to-monitor-force-assignment = {
       ".scratchpad.1" = 1,
       ".scratchpad.2" = 2
   }
   ```

2. **Pre-create per-monitor scratchpads (or let `move` do it)**: when sending a window to a scratchpad that does not exist yet in a multi-monitor setup, `move` provisions it automatically: it focuses the sending window's monitor, creates the empty scratchpad there (`summon-workspace`), verifies the placement, moves the window, then restores the previous focus and the monitor's visible workspace. It never summons an existing scratchpad (that would relocate it with all its windows), and it fails closed — without moving the window — if creation or placement verification fails. Manual pre-creation is still possible: verify absence with `aerospace list-workspaces --all --format '%{workspace} %{monitor-id}'`, focus a window on the intended monitor, run `aerospace summon-workspace .scratchpad.2`, and re-verify the attachment. **Never** run `summon-workspace` on a scratchpad name that already exists with windows in it — it relocates the entire workspace, windows included, to the focused monitor.

   Then pin it with `workspace-to-monitor-force-assignment` above.

3. **Monitor-aware commands**: `list` and `next` default to all monitors; use `--monitor` to narrow their candidates:
   ```bash
   # List scratchpad windows on the current monitor
   aerospace-scratchpad list --monitor current

   # Cycle scratchpad windows across all monitors (default)
   aerospace-scratchpad next

   # Cycle scratchpad windows on monitor 2
   aerospace-scratchpad next --monitor 2
   ```

4. **Hook integration**: The `hook pull-window` command automatically handles both `.scratchpad` and `.scratchpad.<monitor-id>` workspaces.

### Communication with AeroSpaceWM

The communication with AeroSpaceWM is done through an IPC socket client.
See: https://github.com/cristianoliveira/aerospace-ipc
