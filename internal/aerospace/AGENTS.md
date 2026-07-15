# internal/aerospace

Owns the translation between scratchpad behavior and AeroSpace IPC.

## Files

- `client.go`: wraps upstream client, exposes common operations, connection lifetime, and dry-run behavior.
- `querier.go`: finds/filter windows, resolves monitor/workspace names, and persists `next` cycling state.
- `mover.go`: performs workspace, focus, and floating-layout transitions.

## Rules

- Depend on the injected `AeroSpaceWMClient`; keep IPC details behind this boundary.
- Query code discovers state; mover code changes state.
- Preserve `.scratchpad` for single-monitor behavior and `.scratchpad.<monitor-id>` for monitor-specific behavior.
- Use raw `Connection().SendCommand` only when the upstream typed API does not expose the operation.
- Dry-run must avoid mutations while preserving useful output.
- Wrap errors with operation context; do not silently discard IPC failures.

## Tests

- Write focused tests before behavior changes, including no-window/error paths and multi-monitor cases.
- Use deterministic mock monitor/window/workspace state.
- Run `go test ./internal/aerospace -v` and then `make test`.
