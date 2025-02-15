#!/bin/bash -e

IMAGE="nixos/nix@sha256:3bb728719e2c4e478df4c50b80f93adbe27d5c561d1417c3a2306eb914d910da"

rm -f extracted-testcases.txt

docker run -v "$(dirname "$0")":/mount --rm -i "$IMAGE" /bin/sh -c "cd /mount; nix-shell shell.nix --run 'bash run-testcases.sh'"