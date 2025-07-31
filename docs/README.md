# aerospace-scratchpad

Here you will find extensive documentation about the CLI.

## Command: `move`

Move the currently focused window to the `.scratchpad` workspace. The window will be hidden until you show it again.
You can actually see this in your workspace list, but it can be ignored—it's just used to store windows that are "hidden".

### USAGE

`pattern` is a regex pattern to match the app name.

```bash
aerospace-scratchpad move <pattern>
```

For more details:
```bash
aerospace-scratchpad move --help
```

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

## Command: `summon`

Unlike the `show` command, this command will only summon the window to the current workspace and set focus on it.

### USAGE

The `pattern` is a regex pattern to match the "App Name".

```bash
aerospace-scratchpad summon <pattern>
```

## Command: `next`

This command will summon the next window from the scratchpad workspace until there are no more windows to summon.

### USAGE

```bash
aerospace-scratchpad next
```

## Command: `workspace-handler`

This command handles when the scratchpad workspace gets focused (which shouldn't happen). It will move focus back to the last focused workspace and take the focused window to that workspace too.
This allow you to use different tools to focus windows in scratchpad, like notifications, external launchers, etc., and behave as "summoning" the window to the current workspace instead of focusing the window
in the scratchpad workspace.

### USAGE

`aerospace-scratchpad workspace-handler <workspace>`

For more details:

```bash
aerospace-scratchpad workspace-handler --help
```

Add this snippet in your `~/.aerospace.toml` config:
```toml
exec-on-workspace-change = ['/bin/bash', '-c',
    'aerospace-scratchpad workspace-handler $AEROSPACE_FOCUSED_WORKSPACE'
]
```
You can also use the short alias `ws-handler` or `wsh`.

## Implementation details

### Scratchpad workspace

It will send the window to a "special" workspace called `.scratchpad`. This workspace is like any other workspace, but can be ignored. The window will be hidden until you show it again.

### Communication with AeroSpaceWM

The communication with AeroSpaceWM is done through an IPC socket client.
See: https://github.com/cristianoliveira/aerospace-ipc
