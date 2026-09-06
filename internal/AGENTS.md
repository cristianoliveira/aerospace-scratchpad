# Purpose

Contains private runtime packages that implement the CLI's domain boundary and shared infrastructure.

# Boundaries

The package tree hides implementation details from external consumers. The AeroSpace integration translates between scratchpad behavior and the window manager; support packages provide focused policies for output, stable names, logging, and error presentation. Test doubles and test harnesses are test-only and must not become production dependencies.

# Connections

- [Application commands](../cmd/AGENTS.md): Consumes the private contracts exposed by this tree.
- [AeroSpace integration](aerospace/AGENTS.md): Owns window-manager communication and scratchpad state translation.
- [CLI support](cli/AGENTS.md): Owns reusable output and validation policy.
- [Constants](constants/AGENTS.md): Owns stable configuration and environment names.
- [Logging](logger/AGENTS.md): Owns process diagnostics.
- [Error presentation](stderr/AGENTS.md): Owns standard-error behavior.
- [Generated mocks](mocks/AGENTS.md): Provides test doubles for private contracts and external boundaries.
- [Test utilities](testutils/AGENTS.md): Provides deterministic test composition and fixtures.

# Placement

Place production behavior here only when it belongs behind the repository's private boundary. Keep each responsibility in the narrowest package that owns it. Add a child package when the responsibility is cohesive, has a distinct dependency direction, and would otherwise make an existing package coordinate unrelated concerns.
