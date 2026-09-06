# Purpose

Owns Nix packaging definitions for the CLI's supported build variants.

# Boundaries

This module translates repository source and release metadata into reproducible Nix packages and flake outputs. It does not own application behavior or duplicate runtime configuration.

# Connections

- [Repository architecture](../AGENTS.md): Supplies the source project and package ownership context.
- [Maintenance scripts](../scripts/AGENTS.md): Provides packaging maintenance automation such as source-hash updates.

# Placement

Put build inputs, package variants, and Nix-specific distribution policy here. Keep source maintenance automation in scripts and runtime behavior in the application modules. Add a packaging variant only when its source or release semantics differ from an existing variant.
