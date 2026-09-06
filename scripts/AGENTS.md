# Purpose

Owns repository maintenance automation that supports development, packaging, and generated artifacts.

# Boundaries

Scripts coordinate focused maintenance tasks and should remain independent of runtime business decisions. They may invoke repository tools and generators, but they should not become an alternative implementation of application behavior.

# Connections

- [Generated mocks](../internal/mocks/AGENTS.md): Receives generated output from mock-maintenance workflows.
- [Packaging](../nix/AGENTS.md): Supports reproducible packaging metadata maintenance.
- [Repository architecture](../AGENTS.md): Operates within repository-wide ownership and dependency boundaries.

# Placement

Put a responsibility here when it is repeatable repository maintenance rather than runtime behavior. Keep one concern per script and place reusable domain logic in the owning runtime module. Add a new script when the task has a distinct lifecycle or tool contract.
