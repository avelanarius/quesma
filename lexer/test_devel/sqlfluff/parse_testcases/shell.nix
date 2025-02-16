let
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/a93b0806cc75ab6074764f86d7c145779625b189";
  pkgs = import nixpkgs { config = {}; overlays = []; };
in

pkgs.mkShellNoCC {
  packages = [
    (pkgs.python3.withPackages(python-pkgs: [
      (pkgs.python3Packages.toPythonModule(
        pkgs.sqlfluff.overrideAttrs (oldAttrs: {
          version = "0-unstable-2024-02-16";
          src = pkgs.fetchFromGitHub {
            owner = "sqlfluff";
            repo = "sqlfluff";
            rev = "6666db9ed97f45161fb318f901392d9a214808d2";
            hash = "sha256-hb/tbpD0Etf6Wvs4dUfjYvBBbfLquZx2DoN9X0NqwfM=";
          };
          patches = [
            ./sqlfluff.patch
          ];
          doCheck = false;
          doInstallCheck = false;
        })
      ))
    ]))
  ];
}