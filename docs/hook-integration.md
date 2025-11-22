# Hook Integration Guide to enhance UX

**MINIMUM VERSION**: 0.3.0

This document dives deeper into the `hook`s and how to use then to enhance the UX. Hooks are thin commands you can tie into AeroSpace WM events so that scratchpad windows behave predictably when third-party tools (launchers, notifications, scripts) try to focus 
windows inside the scratchpad or otherwise interact with them.

## Why hooks exist? And why to use them?

The origin of this feature was this issue:

  - https://github.com/cristianoliveira/aerospace-scratchpad/issues/53

In short:

The scratchpad windows lives on a dedicated workspace (default: `.scratchpad`). When an action targets a hidden scratchpad window directly, AeroSpace WM focuses the scratchpad workspace instead of “summoning” the window into your current workspace. This feels jarring—you land on a workspace you were never supposed to see.  

## Hook: pull-window - keep focus out of scratchpad workspace

As noted above, when a notification or launcher targets a scratchpad window, AeroSpace WM shifts focus to the scratchpad workspace. This breaks the illusion that the window is being summoned into your current workspace.

The `hook pull-window` subcommand fixes that by:

1. Detecting whenever the scratchpad workspace becomes focused.
2. Fetching the currently focused window.
3. Pulling that window back into the previously-focused workspace.
4. Ensuring focus follows the window so the user never notices the detour.

### Usage

```bash
aerospace-scratchpad hook pull-window <previous-workspace> <focused-workspace>
```

- `previous-workspace`: the workspace the user was on before scratchpad stole focus (`$AEROSPACE_PREV_WORKSPACE` inside AeroSpace hooks).
- `focused-workspace`: the workspace that’s currently focused (`$AEROSPACE_FOCUSED_WORKSPACE` in AeroSpace hooks).

Internally the command:

- Aborts immediately if the “previous” workspace was already the scratchpad (nothing to pull back to).
- Verifies the current focus is the scratchpad before acting.
- Reads the currently focused window via aircraft IPC.
- Uses a temp marker (`/tmp/.aerospace-scratchpad-moving`) to avoid loops when AeroSpace reports rapid focus changes during the move.
- Sends `move-node-to-workspace <previous-workspace> --window-id <id> --focus-follows-window`.

### Wiring it into AeroSpace WM

Add this snippet to your `~/.aerospace.toml` (or `~/.config/aerospace/config.toml`) to run the hook automatically whenever the focused workspace changes:

```toml
exec-on-workspace-change = ["/bin/bash", "-c",
  "aerospace-scratchpad hook pull-window $AEROSPACE_PREV_WORKSPACE $AEROSPACE_FOCUSED_WORKSPACE"
]
```

This leverages AeroSpace’s environment variables:

- `AEROSPACE_PREV_WORKSPACE` → where you *should* remain.
- `AEROSPACE_FOCUSED_WORKSPACE` → what unexpectedly got focus (the scratchpad).

### Step-by-step integration checklist

1. **Install/Update** `aerospace-scratchpad` ≥ v0.3.0 so the `hook` command exists.
2. **Confirm** AeroSpace exports the workspace env vars (AeroSpace ≥ 0.15.x).
3. **Drop** the `exec-on-workspace-change` snippet above into your config.
4. **Reload** AeroSpace (`aerospace reload-config`) or restart it.
5. **Trigger** a notification/launcher that previously jumped you into scratchpad. The window should now be summoned in your current workspace.
6. **Verify** The .scratchpad workspace is not focused at any point.

### Troubleshooting

Hooks share the same logging pattern as other commands. Set:

```bash
export AEROSPACE_SCRATCHPAD_LOGS_LEVEL=DEBUG
export AEROSPACE_SCRATCHPAD_LOGS_PATH=/tmp/aerospace-scratchpad.log
```

Then reproduce the issue. You’ll see messages like:

```
HOOK: pull-window invoked previous-workspace=1 focused-workspace=.scratchpad
HOOK: [final] moved window to new focused workspace workspace=1 window={...}
```

If something fails (e.g., AeroSpace IPC unavailable), the hook writes to stderr and exits non-zero; AeroSpace shows that error in its log.

### Future hooks

The `hook` command can grow more subcommands. Each subcommand should receive state from AeroSpace env vars so it remains declarative and easy to use in other contexts. When adding new hooks:

1. Keep arguments explicit—avoid global state when a notifier can pass context.
2. Document the hook here so users know which AeroSpace events to target.
3. Provide dry-run or verbose flags if the hook has destructive side effects.

Suggestions for expanding this guide? Drop an issue in the repo so we can document additional workflows. This file can be renamed to `hooks-and-integration.md` if you prefer that wording—feel free to adjust when reorganizing docs.
