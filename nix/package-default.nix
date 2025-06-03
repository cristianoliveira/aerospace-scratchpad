{ pkgs, ... }:
  pkgs.buildGoModule rec {
    # name of our derivation
    name = "aerospace-scratchpad";
    # FIXME: once we have the first release, we can use the version
    version = "v0.1.2";
    # version = "v0.1.2";

    # sources that will be used for our derivation.
    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-MJd06jtXnyuULZETboVTg3DLISpZFGiWHNkpybmF0Cc=";
    };

    vendorHash = "sha256-CN/51oTiUm7oGuZY8eMb+8Dc4jA4B6Oqvu6khxOldlo=";

    ldflags = [
      "-s" "-w"
      # "-mod=mod" # FIXME: go forces vendoring some dependencies
      # Change the cli version
      "-X github.com/cristianoliveira/aerospace-scratchpad/cmd.VERSION=${version}"
    ];

    meta = with pkgs.lib; {
      description = "aerospace-scratchpad: Scratchpad for AeroSpaceWM";
      homepage = "https://github.com/cristianoliveira/aerospace-scratchpad";
      license = licenses.mit;
      maintainers = with maintainers; [ cristianoliveira ];
    };
  }
