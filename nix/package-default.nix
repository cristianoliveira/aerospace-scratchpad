{ pkgs, ... }:
  pkgs.buildGoModule rec {
    name = "aerospace-scratchpad";
    version = "v0.5.1";

    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-pjG+ITigcG7oaLi85H6JwmbUszLHJRlqcQZMXtnr+xc=";
    };

    vendorHash = "sha256-HGTE983ZK9jyfjslkMQfmuyngvedxEOg9qL6JxDec4M=";

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
