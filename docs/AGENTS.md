# Purpose

Owns user-facing documentation for installing, configuring, and using the scratchpad CLI.

# Boundaries

Documentation describes supported CLI contracts, integrations, and user-visible behavior. It should not become a second implementation of runtime decisions or preserve details that are not part of the public contract.

# Connections

- [Application commands](../cmd/AGENTS.md): Defines the CLI behavior that documentation explains.
- [Examples](../examples/AGENTS.md): Provides runnable integration material referenced by user guidance.

# Placement

Put user-facing conceptual guidance and integration instructions here. Keep executable demonstrations in examples and implementation rationale in the owning module guide. Add a new documentation area only when a distinct audience or product surface needs an independent information boundary.
