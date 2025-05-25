{ pkgs, ... }:
  pkgs.buildGoModule rec {
    # name of our derivation
    name = "aerospace-scratchpad";
    # FIXME: once we have the first release, we can use the version
    version = "v0.0.1";
    # version = "64ebad853115b0565f284354efb833bf6c045c33";

    # sources that will be used for our derivation.
    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-hkPNzJaAOsaKJXZ/C4sFuMfKTFQxCSIk2Z1th9Hsha0=";
    };

    vendorHash = "sha256-h/GtBDJOOyVXfSv/o4hozZAcHPlg2uJLApy5r3WP9aE=";

    ldflags = [
      "-s" "-w"
      # "-mod=mod" # FIXME: go forces vendoring some dependencies
      "-X main.VERSION=${version}"
    ];

    meta = with pkgs.lib; {
      description = "aerospace-scratchpad: Scratchpad for AeroSpaceWM";
      homepage = "https://github.com/cristianoliveira/aerospace-scratchpad";
      license = licenses.mit;
      maintainers = with maintainers; [ cristianoliveira ];
    };
  }
