let
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/a93b0806cc75ab6074764f86d7c145779625b189";
  pkgs = import nixpkgs { config = {}; overlays = []; };
in

pkgs.mkShellNoCC {
  packages = [
    (pkgs.python3.withPackages(python-pkgs: [
      (python-pkgs.sqlparse.overrideAttrs (oldAttrs: {
          version = "0-unstable-2024-02-17";
          src = pkgs.fetchFromGitHub {
            owner = "andialbrecht";
            repo = "sqlparse";
            rev = "38c065b86ac43f76ffd319747e57096ed78bfa63";
            hash = "sha256-YrzxL/uB8nOwU06qXesXEX93Y47R65PZx1xJ9EbhnGo=";
          };
          doCheck = false;
          doInstallCheck = false;
      }))
    ]))
  ];
}