# Purpose

Owns the repository's stable configuration names and environment keys.

# Boundaries

This module is a behavior-free source of shared names. It may define stable identifiers for scratchpad state, runtime configuration, and environment integration, but it must not hold mutable state or make decisions.

# Connections

- [AeroSpace integration](../aerospace/AGENTS.md): Consumes names needed to map scratchpad behavior to window-manager state.
- [Logging](../logger/AGENTS.md): Consumes logging configuration names.
- [Application commands](../../cmd/AGENTS.md): Consumes names needed by command orchestration.

# Placement

Put a value here when multiple runtime boundaries must agree on one stable name. Keep behavior, parsing, and policy with the module that owns them. Do not create a separate constants area for names used by only one cohesive module.
