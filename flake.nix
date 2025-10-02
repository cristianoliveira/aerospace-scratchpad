{
  description = "aerospace-scratchpad: Scratchpad for AeroSpaceWM";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, utils, ... }: 
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go

            golangci-lint

            # Test runner with good output
            # USAGE: gotestsum --watch
            gotestsum

            # To create new subcommands, run:
            # cobra-cli add <subcommand-name>
            cobra-cli

            # To generate the mock for the interfaces, run:
            # mockgen -source=./pkg/cli/cli.go -destination=./pkg/cli/mock/mock_cli.go -package=mock
            mockgen

            # To proxy connections and debug
            # For testing Unix connections
            # USAGE:
            # sudo mv /tmp/bobko.aerospace-$USER.sock /tmp/bobko.aerospace-$USER.sock.real
            # socat -v UNIX-LISTEN:/tmp/bobko.aerospace-$USER.sock,fork UNIX-CONNECT:/tmp/bobko.aerospace-$USER.sock.real | tee /tmp/socket.log
            socat

            # File watcher 
            # USAGE: (check .watch.yaml for config)
            # fzz 
            funzzy
          ];
        };

        packages = import ./default.nix {
          inherit pkgs;
        };
    });
}
