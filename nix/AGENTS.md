# nix

Nix package definitions for default, source, and nightly builds.

- Keep shared package logic centralized; variants should only override variant-specific source/version inputs.
- Update hashes through `scripts/update-nix-hash.sh` where applicable.
- Preserve flake outputs consumed by `make nix-build`, `make nix-build-source`, and `make nix-build-nightly`.
- Validate the affected target; use `make nix-build-all` when shared packaging changes.
