# Purpose

Owns the translation between scratchpad behavior and the AeroSpace window-manager boundary.

# Boundaries

This module queries window, workspace, and monitor state and applies scratchpad movements, focus changes, and layout transitions through the injected window-manager client. It owns domain decisions that require AeroSpace state. It does not define CLI syntax, serialize command output, or own process-wide lifecycle.

# Connections

- [Application commands](../../cmd/AGENTS.md): Consumes this module's query and movement contracts to implement user-facing operations.
- [Constants](../constants/AGENTS.md): Provides stable scratchpad and state names used when translating window-manager state.
- [Logging](../logger/AGENTS.md): Receives diagnostics for boundary operations and recoverable fallbacks.
- [Test utilities](../testutils/AGENTS.md): Supplies deterministic window-manager scenarios around this boundary.

# Placement

Put a responsibility here when it interprets or changes AeroSpace state for a scratchpad use case. Keep command-specific sequencing in the application layer and generic policies in support modules. Add another adapter module only for a separate external system with its own contract and lifecycle.
