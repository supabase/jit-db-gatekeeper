{
  description = "JIT Database Gatekeeper - PAM module for JWT authentication with PostgreSQL";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.callPackage ./default.nix { };

        packages.jit-db-gatekeeper = self.packages.${system}.default;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            pam
            pkg-config
          ];

          shellHook = ''
            echo "JIT DB Gatekeeper development environment"
            echo "Go version: $(go version)"
          '';
        };
      });
}