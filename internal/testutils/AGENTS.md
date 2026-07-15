# internal/testutils

Shared deterministic test harness for commands and internal packages.

- Centralize reusable client state, IPC command simulation, CLI execution, logger setup, and snapshot helpers here.
- Model AeroSpace behavior needed by tests; avoid copying production algorithms into mocks.
- Keep fixtures explicit and deterministic: no real socket, window manager, clock, or user state.
- Add helpers only when multiple tests benefit; keep scenario-specific setup near its test.
- Snapshot names and output must remain stable unless behavior intentionally changes.
