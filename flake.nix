{
  description = "PAM module example";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachSystem [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" ] (system:
      let
        pkgs = import nixpkgs { inherit system; };

        makeGatekeeper = { go ? pkgs.go }:
          let
            buildGoModule = pkgs.buildGoModule.override { inherit go; };
          in
          buildGoModule {
            pname = "jit-db-gatekeeper";
            version = "1.0.0";
            src = ./.;

            vendorHash = "sha256-gffPp1tNVzWusk5XNsYLkn9x+TD6sELlkik9m+ejNPY=";

            buildInputs = [ pkgs.pam ] ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [
              pkgs.darwin.apple_sdk.frameworks.Security
            ];

            buildPhase = ''
              runHook preBuild
              go build -buildmode=c-shared -o pam_jit_pg.so
              runHook postBuild
            '';

            checkPhase = ''
                runHook preCheck
                go test -v ./...
                runHook postCheck
            '';

            installPhase = ''
              runHook preInstall
              mkdir -p $out/lib/security
              cp pam_jit_pg.so $out/lib/security/
              runHook postInstall
            '';
          };
      in {
        packages.default = makeGatekeeper { };

        lib.makeGatekeeper = makeGatekeeper;
      });
}
