{
  description = "NixXray — TUI app for xray subscription management and TUN connection";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        inherit (pkgs) lib;

        isLinux = lib.hasSuffix "-linux" system;
        isDarwin = lib.hasSuffix "-darwin" system;

        linuxDeps = lib.optionals isLinux (with pkgs; [
          iproute2
          iptables
          nftables
        ]);

        darwinDeps = lib.optionals isDarwin (with pkgs; [
          # TUN on macOS is via utun — no extra packages needed
        ]);
      in
      {
        devShells.default = pkgs.mkShell {
          name = "nixxray";

          packages = with pkgs; [
            # Go toolchain
            go
            gopls
            gotools       # goimports, godoc, etc.
            golangci-lint
            delve         # debugger

            # xray proxy core
            xray

            # Convenience
            just          # Justfile task runner
            git
          ] ++ linuxDeps ++ darwinDeps;

          shellHook = ''
            export GOPATH="$PWD/.gopath"
            export PATH="$GOPATH/bin:$PATH"
            echo "Go $(go version | awk '{print $3}') | xray $(xray version 2>/dev/null | head -1 || echo 'n/a')"
          '';
        };

        # Placeholder package — fill in once go.mod exists
        packages.default = pkgs.buildGoModule {
          pname = "nixxray";
          version = "0.1.0";
          src = ./.;
          vendorHash = null; # replace with real hash after go mod vendor
          meta = {
            description = "TUI for xray subscription and TUN management";
            license = lib.licenses.mit;
            mainProgram = "nixxray";
          };
        };
      }
    );
}
