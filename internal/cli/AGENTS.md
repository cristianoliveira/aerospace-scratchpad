# Purpose

Owns reusable CLI output and argument-validation policy.

# Boundaries

This module converts command results into the supported user-facing and machine-readable formats and validates shared command inputs. It remains independent of command orchestration and window-manager communication. It must not decide scratchpad behavior.

# Connections

- [Application commands](../../cmd/AGENTS.md): Consumes formatting and validation services while retaining command semantics.

# Placement

Put a responsibility here when it is format policy or reusable CLI validation shared by multiple commands. Keep command-specific validation and orchestration in the application layer. Add a new support module only when a policy has a cohesive contract that is independent of both commands and the window manager.
