{ pkgs, ... }:
  pkgs.buildGoModule rec {
    # name of our derivation
    name = "aerospace-scratchpad";
    version = "nightly"; # Branch 

    # sources that will be used for our derivation.
    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-BMN6MgkQAXeHudjrxcfcL9Av2WqjZd+3qtbTKZG8Jyk=";
    };

    vendorHash = "sha256-5yyvDOkXYhw0UAPKa3EVLnzLC0idAfb/BgJAJH86k/w=";

    ldflags = [
      "-s" "-w"
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
