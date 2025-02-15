#!/bin/bash -e

CONTAINER_NAME="quesma_lexer_nix"
IMAGE="nixos/nix@sha256:3bb728719e2c4e478df4c50b80f93adbe27d5c561d1417c3a2306eb914d910da"
SHELL="/bin/sh"

export DOCKER_CLI_HINTS=false

# Check if the container exists
if docker ps -a --format '{{.Names}}' | grep -w "$CONTAINER_NAME" > /dev/null; then
  # Container exists; check if it's running
  if docker ps --format '{{.Names}}' | grep -w "$CONTAINER_NAME" > /dev/null; then
    docker exec -it "$CONTAINER_NAME" "$SHELL"
  else
    docker start -i "$CONTAINER_NAME"
  fi
else
  docker run --name "$CONTAINER_NAME" -it "$IMAGE" "$SHELL"
fi
