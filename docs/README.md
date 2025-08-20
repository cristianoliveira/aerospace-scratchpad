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

## Global flags

### `--dry-run|-d` Dry Run

This flag will not execute the command, but will print what would be done. Very handy to test your command before adding to your
config file.

Usage:
```bash
aerospace-scratchpad --dry-run show <pattern>
```

It will print the actions that would be taken, but will not execute them.

## Implementation details

### Scratchpad workspace

It will send the window to a "special" workspace called `.scratchpad`. This workspace is like any other workspace, but can be ignored. The window will be hidden until you show it again.

### Communication with AeroSpaceWM

The communication with AeroSpaceWM is done through an IPC socket client.
See: https://github.com/cristianoliveira/aerospace-ipc
