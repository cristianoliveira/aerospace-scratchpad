# internal

Private implementation packages for the CLI.

- `aerospace`: scratchpad domain adapter and IPC operations.
- `cli`: formatting and validation independent of Cobra command wiring.
- `constants`: shared stable names and environment keys.
- `logger`, `stderr`: process infrastructure.
- `mocks`, `testutils`: test-only support.

Keep dependencies directed toward small support packages. Avoid imports from `cmd` into `internal`. Production code must not depend on `mocks` or `testutils`.
