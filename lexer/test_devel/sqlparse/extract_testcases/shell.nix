let
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/a93b0806cc75ab6074764f86d7c145779625b189";
  pkgs = import nixpkgs { config = {}; overlays = []; };
in

pkgs.mkShellNoCC {
  packages = [
    (pkgs.python3.withPackages(python-pkgs: [ python-pkgs.pytest ]))
    pkgs.git
  ];
}