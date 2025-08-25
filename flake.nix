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
            pname = "gatekeeper";
            version = "0.1.0";
            src = ./.;

            vendorHash = "sha256-pdF+bhvZQwd2iSEHVtDAGihkYZGSaQaFdsF8MSrWuKQ=";

            buildInputs = [ pkgs.pam ] ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [
              pkgs.darwin.apple_sdk.frameworks.Security
            ];

            buildPhase = ''
              runHook preBuild
              go build -buildmode=c-shared -o pam_jwt_pg.so
              runHook postBuild
            '';

            installPhase = ''
              runHook preInstall
              mkdir -p $out/lib/security
              cp pam_jwt_pg.so $out/lib/security/
              runHook postInstall
            '';
          };
      in {
        packages.default = makeGatekeeper { };
        
        lib.makeGatekeeper = makeGatekeeper;
      });
}
