#!/usr/bin/env bash
# Launch the image directly via the Gondolin CLI with the network policy
# barista needs (Go module proxy + GitHub HTTP allow-hosts).
#
#   .gondolin/bash.sh                # interactive shell at /workspace
#   .gondolin/bash.sh -- go test ./...
set -eu

cd "$(dirname "$0")/.."

ASSETS="$PWD/.gondolin/assets"
if [ ! -d "$ASSETS" ]; then
  echo "No built image found at $ASSETS." >&2
  echo "Run .gondolin/build.sh first." >&2
  exit 1
fi

# Persistent host dirs for the VM's linux/aarch64 Go build + module caches.
# Created on demand. The cow rootfs overlay is ephemeral, so without these
# mounts `go build` would re-download modules and re-compile the stdlib on
# every VM restart. They start empty and fill at runtime.
LINUX_BUILD="$PWD/.gondolin/linux-build"
mkdir -p "$LINUX_BUILD/go-build" "$LINUX_BUILD/gomodcache"

exec gondolin bash \
  --image "$ASSETS" \
  --mount-hostfs "$PWD:/workspace" \
  --mount-hostfs "$LINUX_BUILD/go-build:/opt/go-build" \
  --mount-hostfs "$LINUX_BUILD/gomodcache:/opt/gopath/pkg/mod" \
  --env GOCACHE=/opt/go-build \
  --env GOMODCACHE=/opt/gopath/pkg/mod \
  --env GOPATH=/opt/gopath \
  --env GOROOT=/opt/go \
  --dns-synthetic-host-mapping per-host \
  --allow-host 'proxy.golang.org' \
  --allow-host 'sum.golang.org' \
  --allow-host 'github.com' \
  --allow-host '*.github.com' \
  --allow-host 'objects.githubusercontent.com' \
  --allow-host 'codeload.github.com' \
  --allow-host 'dl.google.com' \
  "$@"
