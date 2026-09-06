# Purpose

Provides generated test doubles for private contracts and external service boundaries.

# Boundaries

This module is test-only. Generated artifacts are owned by their generator and must not contain hand-written production behavior. Production packages must not depend on this tree.

# Connections

- [AeroSpace integration](../aerospace/AGENTS.md): Defines the boundary contracts that require isolated test doubles.
- [Test utilities](../testutils/AGENTS.md): Composes generated doubles into deterministic scenarios.

# Placement

Put generated doubles here when a contract needs isolation in tests. Change the contract or generator input rather than editing generated output. Keep scenario setup and higher-level fixtures in test utilities.
