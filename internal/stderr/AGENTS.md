# Purpose

Owns consistent presentation of command failures on standard error.

# Boundaries

This module separates actionable error presentation from scriptable standard output and applies the process's exit behavior. It does not decide domain outcomes or replace returned errors with logging.

# Connections

- [Application commands](../../cmd/AGENTS.md): Delegates command failure presentation here.
- [Logging](../logger/AGENTS.md): Records failures before presentation.

# Placement

Put shared standard-error wording and exit behavior here. Keep domain context and recovery decisions with the originating module. Create another presentation module only when a separate output channel has a distinct contract.
