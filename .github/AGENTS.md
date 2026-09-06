# Purpose

Owns GitHub-based repository automation and issue intake.

# Boundaries

This module defines when automation runs, how repository checks and release tasks are composed, and how issue templates collect user input. It must preserve least-privilege access and should delegate project behavior to repository-owned commands rather than duplicating it in workflow YAML.

# Connections

- [Maintenance scripts](../scripts/AGENTS.md): Provides repository automation that workflows may invoke.
- [Packaging](../nix/AGENTS.md): Provides build and release outputs used by distribution automation.
- [Repository architecture](../AGENTS.md): Defines the project boundaries that automation validates.

# Placement

Put GitHub-specific triggers, permissions, and workflow composition here. Keep reusable commands in scripts or build definitions and keep application behavior in runtime modules. Add a workflow only when it represents a distinct automation lifecycle or trigger boundary.
