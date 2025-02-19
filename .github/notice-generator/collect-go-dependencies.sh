#!/bin/bash -e

for gomod in **/go.mod ; do
  pushd "$(dirname "$gomod")"
  echo "Processing $gomod" 1>&2

  go mod tidy 1>&2
  go mod download all 1>&2

  go list -m -json all

  popd
done