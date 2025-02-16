let
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/a93b0806cc75ab6074764f86d7c145779625b189";
  pkgs = import nixpkgs { config = {}; overlays = []; };
in

pkgs.mkShellNoCC {
  packages = [
    (pkgs.python3.withPackages(python-pkgs: [
      (pkgs.python3Packages.toPythonModule(
        pkgs.sqlfluff.overrideAttrs {
          version = "0-unstable-2024-02-16";
          src = pkgs.fetchFromGitHub {
            owner = "sqlfluff";
            repo = "sqlfluff";
            tag = "6666db9ed97f45161fb318f901392d9a214808d2";
            hash = "sha256-PQSGB8723y0+cptoLHpXzXfSQFicf5tasbTEf0efA8c=";
          };
          doCheck = false;
          doInstallCheck = false;
        }
      ))
    ]))
  ];
}