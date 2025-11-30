{ pkgs, ... }:
  pkgs.buildGoModule rec {
    name = "aerospace-scratchpad";
    version = "v0.4.0";

    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-hZ9so/37++e6/b9pm6ji6wBzP2Z07siI0XHhAEhcun4=";
    };

    vendorHash = "sha256-1U7Z76k4EUvTSdkMZH8AU790A7K+ONhZizd6q5DAzRE=";

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
