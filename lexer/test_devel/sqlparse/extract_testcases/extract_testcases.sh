#!/bin/bash -e

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

IMAGE="nixos/nix@sha256:3bb728719e2c4e478df4c50b80f93adbe27d5c561d1417c3a2306eb914d910da"

rm -f "$(dirname "$0")"/../../../dialect_sqlparse/test_files/extracted-sqlparse-testcases.txt

docker run -v "$SCRIPTPATH":/mount -v "$SCRIPTPATH"/../../../dialect_sqlparse/test_files:/output --rm -i "$IMAGE" /bin/sh -c "cd /mount; nix-shell shell.nix --run 'bash run_testcases.sh'"