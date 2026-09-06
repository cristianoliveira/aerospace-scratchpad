# Purpose

Owns process-wide diagnostic logging infrastructure.

# Boundaries

This module configures and exposes logging without changing command standard output or making business decisions. Lifecycle ownership remains with the composition root. Test callers may provide no-op or recording implementations through the logging contract.

# Connections

- [Constants](../constants/AGENTS.md): Provides stable logging configuration names.
- [Composition root](../../AGENTS.md): Owns logger construction, registration, and shutdown.
- [AeroSpace integration](../aerospace/AGENTS.md): Emits diagnostics for window-manager operations.
- [Application commands](../../cmd/AGENTS.md): Emits command diagnostics.
- [Error presentation](../stderr/AGENTS.md): Records failures before they are presented to users.

# Placement

Put process diagnostics and logger lifecycle primitives here. Keep user-facing error wording in error presentation and keep domain decisions in their owning module. Add another observability module only when it represents a distinct signal type with an independent contract.
