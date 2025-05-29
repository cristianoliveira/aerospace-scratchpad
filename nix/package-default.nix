{ pkgs, ... }:
  pkgs.buildGoModule rec {
    # name of our derivation
    name = "aerospace-scratchpad";
    # FIXME: once we have the first release, we can use the version
    version = "v0.1.0";
    # version = "v0.1.0";

    # sources that will be used for our derivation.
    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-steDrVNKFdO5g/vLtjSULIHhSiDokK3WZIyTMLrw8zg=";
    };

    vendorHash = "sha256-yz5Zk7I9/5Q6KPkLemuiCk53vfcUn7QXUDl26bN9VgA=";

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
