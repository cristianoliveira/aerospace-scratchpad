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
      sha256 = "sha256-L72ckR0VqmoMeSXJI965ldg3FT/nh0EnHKHO0PsiLVs=";
    };

    vendorHash = "sha256-u2tlLqfcKG8JbzdAA5RnH8m0yzjCGKxQv0FsgxZKDYI=";

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
