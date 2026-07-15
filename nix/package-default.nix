{ pkgs, ... }:
  pkgs.buildGoModule rec {
    name = "aerospace-scratchpad";
    version = "v0.6.0";

    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-BOy4wTvqWCgNxqKybLAl3iVe0T+P9Y9JcbupXb3KhyM=";
    };

    vendorHash = "sha256-jYCA3DNoi9Mj55t6ru76voENM8VSJt7Gp74JV7Vq19k=";

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
