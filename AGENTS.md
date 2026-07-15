# aerospace-scratchpad

Go CLI providing i3/Sway-like scratchpads for AeroSpace WM through `aerospace-ipc`.

## Architecture

- `main.go`: composition root; create logger and IPC client, then call `cmd.Execute`.
- `cmd/`: Cobra commands and use-case orchestration. See `cmd/AGENTS.md`.
- `internal/aerospace/`: AeroSpace adapter and scratchpad query/move logic. See its guide.
- `internal/cli/`: output formatting and CLI validation.
- `internal/constants/`, `internal/logger/`, `internal/stderr/`: shared infrastructure.
- `internal/mocks/`, `internal/testutils/`: generated mocks and test support.
- `docs/`, `examples/`: user documentation and integrations.
- `nix/`, `scripts/`, `.github/`: packaging and automation.

Dependency direction is `main -> cmd -> internal packages`. Keep business decisions out of `main.go`. Production packages must not import test support.

## Domain rules

- Hidden windows live in `.scratchpad` or `.scratchpad.<monitor-id>`.
- Scratchpad windows include windows on scratchpad workspaces and floating windows.
- Preserve distinct semantics: `show` toggles, `summon` brings forward, `move` hides, `next` cycles.
- Use the injected `AeroSpaceWMClient`; do not spawn `aerospace` processes when IPC supports the operation.

## Workflow

- Follow test-first development. Cover successful and error paths.
- Match nearby table-driven and snapshot test style.
- Run `make test`, `make lint`, and `make build` before finishing.
- Update snapshots only for intentional output changes: `make update-snap-all`.
- Use `make fmt` for Go formatting and lint fixes.
- Generated files under `internal/mocks/` must be regenerated, not hand-edited.

External AeroSpace references are available under `.tmp/docs/`; runtime logs are under `.tmp/aerospace-scratchpad.log`.
