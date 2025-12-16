{ pkgs ? import <nixpkgs> {} }:

pkgs.buildGoModule {
  pname = "jit-db-gatekeeper";
  version = "1.0.1";

  src = ./.;

  vendorHash = null;

  nativeBuildInputs = with pkgs; [
    pkg-config
  ];

  buildInputs = with pkgs; [
    pam
  ];

  env.CGO_ENABLED = "1";

  # Build as shared library for PAM
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

  meta = with pkgs.lib; {
    description = "PAM module for JWT authentication with PostgreSQL backend";
    homepage = "https://github.com/supabase/jit-db-gatekeeper";
    license = licenses.mit;
    maintainers = [ ];
    platforms = platforms.unix;
  };
}
