# Purpose

Provides small, runnable integrations that demonstrate real CLI workflows.

# Boundaries

Examples should compose documented commands and machine-readable output rather than reimplement scratchpad or filtering behavior. They should make assumptions about external tools explicit and handle command failures at the integration boundary.

# Connections

- [Documentation](../docs/AGENTS.md): Defines the user-facing context and discoverability for examples.
- [Application commands](../cmd/AGENTS.md): Supplies the CLI contract demonstrated by examples.

# Placement

Put a responsibility here when it is a standalone integration or usage demonstration. Keep reusable runtime behavior in the application or private runtime modules. Add a separate example area only when an integration has a distinct dependency or lifecycle.
