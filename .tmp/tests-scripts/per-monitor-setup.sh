#!/bin/bash

## IMPORTANT: All IDs and NAMEs here are just examples.

aerospace list-monitors --format '%{monitor-id} | %{monitor-name} | %{monitor-is-main}'
# output example:
# 1 | DELL BLA | true
# 2 | DELL FOOBAR | false

aerospace list-windows --all --format '%{window-id} | %{app-name} | %{monitor-id} | %{monitor-name} | %{monitor-is-main}' | grep Finder
# output example:
# 31 | Finder | 2 | DELL P2415H | true
WIN_ID=$(aerospace list-windows --all --format '%{window-id} | %{app-name} | %{monitor-id} | %{monitor-name} | %{monitor-is-main}' | grep Finder | awk '{ print $1 }')

# Focus on monitor 1
aerospace focus-monitor "$WIN_ID"

# Bring window to monitor 1
aerospace move-node-to-monitor 1 --window-id "$WIN_ID"

# Move Finder to scratchpad
aerospace-scratchpad move Finder

aerospace list-windows --all --format '%{window-id} | %{app-name} | %{monitor-id} -- %{workspace}' | grep Finder
# output example:
# 31 | Finder | 2 | 1 -- 1

# After fix:
# 31 | Finder | 2 -- .scratchpad.1

## Second test to ensure we are on the same page.
aerospace list-windows --all --format '%{window-id} | %{app-name} | %{monitor-id} | %{monitor-name} | %{monitor-is-main}' | grep Clock
# output example:
# 22 | Clock | 1 | DELL P2415H | true
WIN_2_ID=$(aerospace list-windows --all --format '%{window-id} | %{app-name} | %{monitor-id} | %{monitor-name} | %{monitor-is-main}' | grep Clock | awk '{ print $1 }')

# Focus on monitor 1
aerospace focus-monitor "$WIN_2_ID"

# Bring window to monitor 1
aerospace move-node-to-monitor 2 --window-id "$WIN_2_ID"

# Move Finder to scratchpad
aerospace-scratchpad move Clock

aerospace list-windows --all --format '%{window-id} | %{app-name} | %{monitor-id} -- %{workspace}' | grep -E 'Clock|Finder'
# Expected output:
# 31 | Finder | 2 -- .scratchpad.1
# 22 | Clock  | 2 -- .scratchpad.2

