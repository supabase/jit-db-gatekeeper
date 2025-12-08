{ pkgs ? import <nixpkgs> {} }:

pkgs.buildGoModule {
  pname = "jit-db-gatekeeper";
  version = "1.0.0";

  src = ./.;

  vendorHash = "sha256-G9x2TARSJMn30R6ZOlsggxEtn5t2ezWz1YtkLXdYiAE=";

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
