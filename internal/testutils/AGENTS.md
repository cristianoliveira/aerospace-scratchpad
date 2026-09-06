# Purpose

Provides deterministic test composition, fixtures, and adapters shared by package tests.

# Boundaries

This module models the external boundary needed by tests, assembles mock services, and provides reusable command and output helpers. It must remain test-only and must not duplicate production algorithms or depend on real sockets, window-manager state, or user state.

# Connections

- [AeroSpace integration](../aerospace/AGENTS.md): Uses the domain boundary and its state types to describe scenarios.
- [Generated mocks](../mocks/AGENTS.md): Composes generated service doubles.
- [Logging](../logger/AGENTS.md): Provides deterministic logging implementations for tests.
- [Application commands](../../cmd/AGENTS.md): Supports isolated command execution and assertions.

# Placement

Put reusable test composition here when more than one package benefits from it. Keep scenario-specific setup beside the test that owns the scenario. Add a helper only when it reduces duplicated setup without reimplementing production behavior.
