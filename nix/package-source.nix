{ pkgs, ... }:
  pkgs.buildGoModule rec {
    # name of our derivation
    name = "aerospace-scratchpad";
    version = "source";

    # sources that will be used for our derivation.
    src = ../.;

    vendorHash = "sha256-7O1Yl9bSCy+cyeMy2KgMifXXyudNQk0HbYws8RyJJ5M=";

    ldflags = [
      "-s" "-w"
      "-X main.VERSION=${version}"
    ];

    meta = with pkgs.lib; {
      description = "aerospace-scratchpad: Scratchpad for AeroSpaceWM";
      homepage = "https://github.com/cristianoliveira/aerospace-scratchpad";
      license = licenses.mit;
      maintainers = with maintainers; [ cristianoliveira ];
    };
  }
