{ pkgs, ... }:
  pkgs.buildGoModule rec {
    name = "aerospace-scratchpad";
    version = "v0.7.1";

    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-ujKRoiJjZe1Fmrtj6+JXfT7qlb84EuIkqWC2KTbu9P0=";
    };

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
