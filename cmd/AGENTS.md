# cmd

Cobra presentation and application-orchestration layer.

## Responsibilities

- Define command arguments, flags, help, validation, and output events.
- Coordinate `internal/aerospace` query and move operations.
- Keep command semantics distinct: `move`, `show`, `summon`, `next`, `list`, `hook`, and `info`.
- Wire shared flags in `root.go`; do not duplicate flag definitions across commands.

## Boundaries

- Put reusable AeroSpace querying or movement rules in `internal/aerospace`, not command files.
- Put reusable serialization in `internal/cli`.
- Commands receive interfaces/clients from `RootCmd`; do not construct IPC clients here.
- Send scriptable results through the output formatter and errors through the established stderr behavior.

## Tests

- Add or change command tests first in matching `*_test.go` files.
- Test success and failure behavior using `internal/testutils` and mocks.
- Snapshot user-visible output. Update snapshots only when the contract intentionally changes.
- Verify with `go test ./cmd -v`; for output changes also run `make update-snap-all` and inspect diffs.
