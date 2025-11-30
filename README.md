# aerospace-scratchpad
[![Go project version](https://badge.fury.io/go/github.com%2Fcristianoliveira%2Faerospace-scratchpad.svg)](https://badge.fury.io/go/github.com%2Fcristianoliveira%2Faerospace-scratchpad)
[![CI](https://github.com/cristianoliveira/aerospace-scratchpad/actions/workflows/on-push.yml/badge.svg)](https://github.com/cristianoliveira/aerospace-scratchpad/actions/workflows/on-push.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cristianoliveira/aerospace-scratchpad.svg)](https://pkg.go.dev/github.com/cristianoliveira/aerospace-scratchpad)

A I3/Sway like scratchpad extension for [AeroSpace WM](https://github.com/nikitabobko/AeroSpace)

# IMPORTANT for AeroSpace v0.20.0

Currently incompatible with AeroSpace v0.20.0 work is being done in and will be release with v0.4.0

## Summary

- [Demo](#demo)
- [Basic Usage](#basic-usage)
- [Advanced Usage](#advanced-usage)
- [Installation](#installation)
- [How does it work?](#how-does-it-work)
- [Troubleshooting](#troubleshooting)
- [License](#license)

**Beta**: I use it daily, I'll try my best but there might be breaking changes till the 1.0.0 release. 

Please report any issues or ideas in the [issues](https://github.com/cristianoliveira/aerospace-scratchpad/issues) section.

## Demo

https://github.com/user-attachments/assets/48642cc7-3a5f-4037-863a-eaa493a7b10c

From [I3 User Guide](https://i3wm.org/docs/userguide.html#_scratchpad):
> 6.21. Scratchpad
> 
> There are two commands to use any existing window as scratchpad window. move scratchpad will move a window to the scratchpad workspace. This will make it invisible until you show it again. There is no way to open that workspace. Instead, when using scratchpad show, the window will be shown again, as a floating window, centered on your current workspace (using scratchpad show on a visible scratchpad window will make it hidden again, so you can have a keybinding to toggle). Note that this is just a normal floating window, so if you want to "remove it from scratchpad", you can simple make it tiling again (floating toggle).
>
> As the name indicates, this is useful for having a window with your favorite editor always at hand. However, you can also use this for other permanently running applications which you don’t want to see all the time: Your music player, alsamixer, maybe even your mail client…?

[What Are Scratchpads and WHY Are They Good?](https://youtu.be/72ccdlOWe20?si=tyYhdW6_BRJSCSHr)

[The origin of this project](https://github.com/nikitabobko/AeroSpace/issues/272)

## Basic Usage

Summon or move a window from the scratchpad workspace to the current workspace.
```text
aerospace-scratchpad show <pattern>
```

To find the correct `pattern` run `aerospace list-windows --all --json | grep app-name`

### Config Usage

```toml
# ~/.config/aerospace/config.toml
[mode.main.binding] 
# This toggle the scratchpad window show/hide
cmd-ctrl-1 = "exec-and-forget aerospace-scratchpad show Finder"

# Or using summon instead
cmd-ctrl-2 = """
exec-and-forget aerospace-scratchpad summon Finder || \
                aerospace-scratchpad move Finder
"""

# Bring windows one by one to current workspace
ctrl-minus = "exec-and-forget aerospace-scratchpad next"

# A terminal scratchpad a la Guake
ctrl-cmd-t = """
exec-and-forget aerospace-scratchpad show alacritty -F window-title='terminal-scratchpad' \
             || alacritty -t 'terminal-scratchpad'
"""
```

#### For better UX

_Minimum version: 0.3.0_

The scratchpad windows lives on a dedicated workspace (default: `.scratchpad`). In order to avoid that workspace taking focus add this to your config:

```toml
# ~/.config/aerospace/config.toml
# ...your configuration
exec-on-workspace-change = ["/bin/bash", "-c",
  "aerospace-scratchpad hook pull-window $AEROSPACE_PREV_WORKSPACE $AEROSPACE_FOCUSED_WORKSPACE"
]
```

For more details check [Hook commands](docs/hook-integration.md)

## Advanced Usage

See more in [documentation](docs/)

### Dinamic scratchpads mapping

By pairing with [aerospace-marks](https://github.com/cristianoliveira/aerospace-marks) you 
can turn any window into a scratchpad window and bind a toggle key on the fly!

```toml
# ~/.config/aerospace/config.toml
[mode.main.binding]
# Mark current window with sp-1 so you can use the shortcut
cmd-shift-ctrl-1 = [
    "exec-and-forget aerospace-marks mark sp-1"
]

# Toggle show/hide the window marked as 'sp-1' as scratchpad window
# If window is not marked does nothing
cmd-ctrl-1 = "exec-and-forget aerospace-scratchpad show \"$(aerospace-marks get sp-1 -a)\""

# Making a mark more specific
ctrl-cmd-t = """
exec-and-forget aerospace-scratchpad show \
               "$(aerospace-marks get term -a)" \
               -F window-title="$(aerospace-marks get term -t)"
"""
```

## Installation

**Min AeroSpace version**: 0.15.x

### Using Homebrew

If you have Homebrew installed, you can install `aerospace-scratchpad` using the following command:

```bash
brew install cristianoliveira/tap/aerospace-scratchpad
```
See the tap definition for other versions [https://github.com/cristianoliveira/homebrew-tap](https://github.com/cristianoliveira/homebrew-tap)

### Nix

If you have Nix installed, you can build and install `aerospace-scratchpad` using the following command:

```bash
nix profile install github:cristianoliveira/aerospace-scratchpad
```

You can also run without installing it by using:

```bash
nix run github:cristianoliveira/aerospace-scratchpad
```

This will build the default package defined in `flake.nix`.

### Go

If you have Go installed, you can install `aerospace-scratchpad` directly using:

```bash
go install github.com/cristianoliveira/aerospace-scratchpad@latest
```

This will download and install the latest version of `aerospace-scratchpad` to your `$GOPATH/bin`.

#### Post installation

After installing, start by checking compatibility and updating Aerospace accordingly:
```bash
aerospace-scratchpad info
```

Next, you may need to include aerospace-scratchpad in the Aerospace context.

To check where the binary is installed, run:
```bash
echo $(which aerospace-scratchpad) | sed 's/\/aerospace-scratchpad//g'
```

Then, in your configuration file, add:
```toml
[exec]
    inherit-env-vars = true
# OR
[exec.env-vars]
    # Replace 'aerospace-scratchpad/install/path' with the actual path from the command above
    PATH = 'aerospace-scratchpad/install/path/bin:${PATH}'
```

## How does it work?

This extension uses Inter-Process Communication (IPC), specifically through a Unix socket, to communicate directly with the AeroSpace service, just like the built-in AeroSpace CLI. By avoiding repeated process spawning, this approach offers lower latency and better efficiency, especially when one has to query AeroSpace many times.

See: https://github.com/cristianoliveira/aerospace-ipc

### Benchmarks

This CLI runs about *3x faster* than a bash script that does the same.

```bash
# time aerospace-scratchpad show Finder
# aerospace-scratchpad show Finder  0.01s user 0.01s system 12% cpu 0.125 total
# aerospace-scratchpad show Finder  0.01s user 0.01s system 14% cpu 0.097 total
# aerospace-scratchpad show Finder  0.01s user 0.01s system 12% cpu 0.140 total

# time bash scripts/scratchpad.sh Finder
# scripts/benchmark.sh Finder  0.13s user 0.08s system 75% cpu 0.281 total
# scripts/benchmark.sh Finder  0.14s user 0.09s system 82% cpu 0.276 total
# scripts/benchmark.sh Finder  0.14s user 0.09s system 80% cpu 0.289 total
```

See: `scripts/benchmarks.sh` for details, and test it yourself.

## Troubleshooting

If you encounter issues with `aerospace-scratchpad`, you can use the following environment variables to help diagnose and resolve problems:

- **AEROSPACE_SCRATCHPAD_LOGS_PATH**: Set this environment variable to specify the path for the AeroSpace scratchpad logs. By default, logs are stored at `/tmp/aerospace-scratchpad.log`.

- **AEROSPACE_SCRATCHPAD_LOGS_LEVEL**: Use this environment variable to set the logging level for the AeroSpace scratchpad. The default level is `DISABLED`. You can set it to other levels like `DEBUG` to get more detailed logs.

These environment variables can be set directly in the AeroSpace configuration file to ensure they are available whenever AeroSpace is running. Add the following to your [AeroSpace config](https://nikitabobko.github.io/AeroSpace/guide#config-location)

```toml
[exec.env-vars]
    AEROSPACE_SCRATCHPAD_LOGS_PATH = "/path/to/your/logfile.log"
    AEROSPACE_SCRATCHPAD_LOGS_LEVEL = "DEBUG"
```

Alternatively, you can export these environment variables in your shell configuration file (e.g., `.bashrc`, `.zshrc`):

```bash
  export AEROSPACE_SCRATCHPAD_LOGS_LEVEL="DEBUG"
  export AEROSPACE_SCRATCHPAD_LOGS_PATH="/path/to/your/logfile.log"
```

Replace the paths and values with your desired settings.

## License

This project is licensed under the terms of the LICENSE file.
