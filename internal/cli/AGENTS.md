# internal/cli

Reusable command-line output and validation helpers.

- Keep formatting independent of Cobra and AeroSpace IPC.
- Preserve output contracts for `text`, `json`, `tsv`, and `csv`.
- Add fields once in the shared event/row model; do not implement format-specific duplicate business rules.
- Quote separated values deterministically.
- Validate unsupported formats and invalid arguments explicitly.

Write tests first for every format and error path. Run `go test ./internal/cli -v`; intentional output changes may require command snapshot updates.
