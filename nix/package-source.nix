{ pkgs, ... }:
  pkgs.buildGoModule rec {
    # name of our derivation
    name = "aerospace-scratchpad";
    version = "source";

    # sources that will be used for our derivation.
    src = ../.;

    vendorHash = "sha256-3GiKZiWUG6W+qbtGkDPjxRu2L/pZeqHkGtyWXMOyP9o=";

    ldflags = [
      "-s" "-w"
      "-X github.com/cristianoliveira/aerospace-scratchpad/cmd.VERSION=${version}"
    ];

    meta = with pkgs.lib; {
      description = "aerospace-scratchpad: Scratchpad for AeroSpaceWM";
      homepage = "https://github.com/cristianoliveira/aerospace-scratchpad";
      license = licenses.mit;
      maintainers = with maintainers; [ cristianoliveira ];
    };
  }
