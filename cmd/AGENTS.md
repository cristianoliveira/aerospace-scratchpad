# Purpose

Owns the CLI surface and application orchestration for scratchpad operations.

# Boundaries

This module defines command contracts, argument and flag handling, validation flow, use-case coordination, and user-visible command results. It must not own reusable window-manager translation, generic output serialization, or process-wide infrastructure. Dependencies are supplied by the composition root rather than constructed here.

# Connections

- [Private runtime packages](../internal/AGENTS.md): Provides the internal boundaries consumed by commands.
- [AeroSpace integration](../internal/aerospace/AGENTS.md): Supplies window-manager queries and state-changing operations.
- [CLI support](../internal/cli/AGENTS.md): Supplies output formatting and argument validation.
- [Logging](../internal/logger/AGENTS.md): Receives diagnostic events without changing scriptable output.
- [Error presentation](../internal/stderr/AGENTS.md): Presents command failures separately from standard output.
- [Composition root](../AGENTS.md): Supplies dependencies and owns process-level wiring.

# Placement

Put a responsibility here when it coordinates a user-facing command or defines its CLI contract. Put reusable domain decisions in the AeroSpace integration, and put cross-command formatting or validation in CLI support. Create a new command-focused module only when a separate command family has its own cohesive orchestration boundary.
