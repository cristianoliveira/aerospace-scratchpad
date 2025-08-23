{ pkgs, ... }:
  pkgs.buildGoModule rec {
    # name of our derivation
    name = "aerospace-scratchpad";
    # FIXME: once we have the first release, we can use the version
    version = "v0.2.0";
    # version = "v0.2.0";

    # sources that will be used for our derivation.
    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-sjrqPBLfWVE9QIOF50GyeSaeU/3XrMrthqX/6zwsMvk=";
    };

    vendorHash = "sha256-u2tlLqfcKG8JbzdAA5RnH8m0yzjCGKxQv0FsgxZKDYI=";

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
