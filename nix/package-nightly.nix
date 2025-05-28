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
      sha256 = "sha256-k2np+u+gTyl077KaqyHTQKO9Crmb2Yev+aiM1SVicM8=";
    };

    vendorHash = "sha256-h/GtBDJOOyVXfSv/o4hozZAcHPlg2uJLApy5r3WP9aE=";

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
