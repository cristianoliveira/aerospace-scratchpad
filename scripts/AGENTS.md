# scripts

Repository maintenance automation.

- Scripts must be non-interactive when used by CI and fail on errors.
- Keep version validation, mock generation, Nix hash updates, and benchmarks focused in separate scripts.
- Resolve paths relative to repository root rather than caller working directory.
- Do not hand-edit outputs when an existing generator owns them.
- Test changed scripts with a timeout and run the related Make target.
