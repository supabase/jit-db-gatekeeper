{
  description = "PAM module example";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = f: builtins.listToAttrs (map (system: {
        name = system;
        value = f system;
      }) systems);
    in {
      packages = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in {
          default = pkgs.stdenv.mkDerivation {
            pname = "gatekeeper";
            version = "0.1.0";
            src = ./.;

            buildInputs = [ pkgs.pam pkgs.gcc ];

            # Assuming your pam module source is pam_foo.c
            buildPhase = ''
                go build -buildmode=c-shared -o pam_jwt_pg.so
            '';

            installPhase = ''
              mkdir -p $out/lib/security
              cp pam_jwt_pg.so $out/lib/security/
            '';
          };
        });
    };
}
