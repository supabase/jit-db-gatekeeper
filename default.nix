{ pkgs ? import <nixpkgs> {} }:

pkgs.stdenv.mkDerivation {
  pname = "pam_jwt_pg";
  version = "0.0.1";

  src = ./.;

  nativeBuildInputs = [
    pkgs.go
  ];

  buildInputs = [
    pkgs.pam
  ];

  CGO_ENABLED = "1";
  GOOS = "linux";
  GOARCH = "arm64";
  GOPATH = "${placeholder "out"}/go"; # Prevent writing to /homeless-shelter
  GOMODCACHE = "${placeholder "out"}/gomodcache";

  buildPhase = ''
    export HOME=$PWD
    go mod tidy
    go build -buildmode=c-shared -o pam_jwt_pg.so
  '';

  installPhase = ''
    mkdir -p $out/lib/security
    cp pam_jwt_pg.so $out/lib/security/
  '';
}
