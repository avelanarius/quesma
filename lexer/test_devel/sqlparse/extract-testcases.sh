#!/bin/bash -e

IMAGE="nixos/nix@sha256:3bb728719e2c4e478df4c50b80f93adbe27d5c561d1417c3a2306eb914d910da"

(
cat <<-"EOF"

cat <<EOF2 > shell.nix
let
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/a93b0806cc75ab6074764f86d7c145779625b189";
  pkgs = import nixpkgs { config = {}; overlays = []; };
in

pkgs.mkShellNoCC {
  packages = [
    (pkgs.python3.withPackages(python-pkgs: [ python-pkgs.pytest ]))
  ];
}
EOF2

cat <<-"EOF2" > sqlparse.patch
diff --git a/sqlparse/lexer.py b/sqlparse/lexer.py
index 8f88d17..92f04b5 100644
--- a/sqlparse/lexer.py
+++ b/sqlparse/lexer.py
@@ -134,6 +134,9 @@ class Lexer:
             raise TypeError("Expected text or file-like object, got {!r}".
                             format(type(text)))

+        with open("/mount/extracted-testcases.txt", "a") as file:
+            file.write(text + "\n<end_of_query/>\n")
+
         iterable = enumerate(text)
         for pos, char in iterable:
             for rexmatch, action in self._SQL_REGEX:
EOF2

nix-shell shell.nix --run "/bin/sh" <<-"EOF2"

git clone https://github.com/andialbrecht/sqlparse.git
cd sqlparse
git reset --hard 38c065b86ac43f76ffd319747e57096ed78bfa63
git apply ../sqlparse.patch

pytest
EOF2

EOF
) | docker run -v "$(dirname '$0')":/mount --rm -i "$IMAGE" /bin/sh