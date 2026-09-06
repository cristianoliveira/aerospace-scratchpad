# Purpose

AeroSpace Scratchpad provides an i3/Sway-style scratchpad experience for AeroSpace through a scriptable command-line interface.

# Architecture

The composition root owns dependency wiring and process lifetime. The application layer owns command orchestration and user-facing command contracts. Private packages separate AeroSpace state translation, output policy, shared configuration names, logging, and error presentation. Packaging, documentation, examples, maintenance automation, and repository automation remain outside the runtime path.

Runtime dependencies point inward: composition wiring calls the application layer; the application layer consumes private contracts; the AeroSpace boundary owns communication with the window manager. Shared support packages must not pull command orchestration back into infrastructure.

# Modules

- [Application commands](cmd/AGENTS.md): Defines the CLI surface and coordinates use cases.
- [Private runtime packages](internal/AGENTS.md): Contains domain adapters and shared infrastructure.
- [Documentation](docs/AGENTS.md): Maintains user-facing product guidance.
- [Examples](examples/AGENTS.md): Demonstrates portable integrations with the CLI.
- [Packaging](nix/AGENTS.md): Defines Nix build and distribution variants.
- [Maintenance scripts](scripts/AGENTS.md): Owns repository maintenance automation.
- [Repository automation](.github/AGENTS.md): Owns GitHub workflows and issue templates.

# Placement

Place behavior that coordinates a user command in the application layer, and place window-manager translation behind the private AeroSpace boundary. Keep formatting, configuration names, logging, and error presentation in their focused support modules. A new module is justified when a responsibility has a cohesive owner and a distinct dependency direction; otherwise extend the nearest existing module. Keep packaging and documentation changes outside runtime packages.
