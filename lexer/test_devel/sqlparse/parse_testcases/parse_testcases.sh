#!/bin/bash -e

ROOTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"/../../../
IMAGE="nixos/nix@sha256:3bb728719e2c4e478df4c50b80f93adbe27d5c561d1417c3a2306eb914d910da"

docker run -v quesma_lexer_nix:/nix -v quesma_lexer_nix_cache:/root/.cache/ -v "$ROOTPATH":/mount --rm -i "$IMAGE" /bin/sh -c "cd /mount/test_devel/sqlparse/parse_testcases; nix-shell shell.nix --run 'python3 parse_testcases.py'"