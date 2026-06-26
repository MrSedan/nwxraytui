{
  description = "nwxraytui — TUI app for xray subscription management and TUN connection";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        inherit (pkgs) lib;

        isLinux = lib.hasSuffix "-linux" system;
        isDarwin = lib.hasSuffix "-darwin" system;

        linuxDeps = lib.optionals isLinux (
          with pkgs;
          [
            iproute2
            iptables
            nftables
          ]
        );
        darwinDeps = lib.optionals isDarwin [ ];
      in
      {
        devShells.default = pkgs.mkShell {
          name = "nwxraytui";
          packages =
            with pkgs;
            [
              go
              gopls
              gotools
              golangci-lint
              delve
              xray
              just
              git
            ]
            ++ linuxDeps
            ++ darwinDeps;
          shellHook = ''
            export GOPATH="$PWD/.gopath"
            export PATH="$GOPATH/bin:$PATH"
            echo "Go $(go version | awk '{print $3}') | xray $(xray version 2>/dev/null | head -1 || echo 'n/a')"
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "nwxraytui";
          version = "0.1.1";
          src = ./.;
          vendorHash = "sha256-qrX55UC7IMOZS8yDB+JIf5fAatfsRaMl38T1rDKHSAg=";
          nativeBuildInputs = with pkgs; [
            makeWrapper
          ];

          postInstall = ''
            wrapProgram $out/bin/nwxraytui \
              --prefix PATH : ${
                lib.makeBinPath (
                  [
                    pkgs.xray
                    pkgs.git
                  ]
                  ++ linuxDeps
                  ++ darwinDeps
                )
              }
          '';
          meta = {
            description = "TUI for xray subscription and TUN management";
            license = lib.licenses.mit;
            mainProgram = "nwxraytui";
          };
        };
      }
    )
    // {
      nixosModules.nwxraytui = import ./nix/module.nix;
      darwinModules.nwxraytui = import ./nix/darwin-module.nix;
    };
}
