# internal/logger

Process-wide structured logging infrastructure.

- Configuration comes from the environment constants package.
- Keep logger initialization and closing in the composition root.
- Logging must not alter command stdout contracts.
- Preserve the no-op logger for deterministic tests.
- Never log as a substitute for returning an actionable error.

Run package tests when changing configuration, levels, file handling, or JSON helpers, followed by `make test`.
