# Hook Integration Guide to enhance Scratchpad's UX

**MINIMUM VERSION**: 0.3.0

This document dives deeper into the `hook`s and how to use then to enhance the UX. Hooks are thin commands you can tie into AeroSpace WM events so that scratchpad windows behave predictably when third-party tools (launchers, notifications, scripts) try to focus windows inside the scratchpad or otherwise interact with them.

## Why hooks exist? And why to use them?

The origin of this feature was this issue:

  - https://github.com/cristianoliveira/aerospace-scratchpad/issues/53

In short:

The scratchpad windows live on dedicated workspaces (default: `.scratchpad`, or `.scratchpad.<monitor-id>` for multi-monitor setups). When an action targets a hidden scratchpad window directly, AeroSpace WM focuses the scratchpad workspace instead of “summoning” the window into your current workspace. This is a bit annoying, you land on a workspace you were never supposed to see where is probably very messy.

## Hook: pull-window - To keep focus out of scratchpad workspace

As noted above, when a notification or launcher targets a scratchpad window, AeroSpace WM shifts focus to the scratchpad workspace. Windows managers like Sway and I3 solves that by attaching the special workspace
to hidden Output/Monitor which never takes focus, unfortunately macOS doesn't allow us to do that without the need of tools like https://github.com/Stengo/DeskPad

The `hook pull-window` subcommand fixes that by:

1. Detecting whenever the scratchpad workspace becomes focused.
2. Immediatly changing focus back to the prev workspace.
3. Pulling the focused window back into the current workspace.
4. Ensuring focus follows the window. *There is a bit of fleakering during this process*

### Usage

```bash
aerospace-scratchpad hook pull-window <previous-workspace> <focused-workspace>
```

- `previous-workspace`: the workspace the user was on before scratchpad stole focus (`$AEROSPACE_PREV_WORKSPACE` inside AeroSpace hooks).
- `focused-workspace`: the workspace that’s currently focused (`$AEROSPACE_FOCUSED_WORKSPACE` in AeroSpace hooks).

### Integrating it into AeroSpace WM configuration

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

1. **Install/Update** `aerospace-scratchpad` ≥= v0.3.0 so the `hook` command exists.
2. **Confirm** AeroSpace exports the workspace env vars (AeroSpace ≥ 0.15.x).
3. **Drop** the `exec-on-workspace-change` snippet above into your config.
4. **Reload** AeroSpace (`aerospace reload-config`) or restart it.
5. **Trigger** a notification/launcher that previously jumped you into scratchpad. The window should now be summoned in your current workspace.
6. **Verify** that no scratchpad workspace (`.scratchpad` or `.scratchpad.<monitor-id>`) is focused at any point.

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

If something fails, the hook writes to stderr and exits non-zero; AeroSpace shows that error in its log.
