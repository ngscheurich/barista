#!/usr/bin/env bash
# Build the custom Gondolin guest image for the barista project.
#
# Run this on your host (macOS/Linux), NOT inside a Gondolin VM:
#
#   .gondolin/build.sh
#
# Requirements on the host: `gondolin` CLI + the build tools listed at
# https://earendil-works.github.io/gondolin/custom-images/ (cpio, lz4,
# e2fsprogs). Because postBuild.commands runs in a Linux chroot, building on
# macOS automatically uses a container — so Docker or Podman must be available
# (and running). Building on Linux needs root (chroot) instead.
#
# This build installs the Go toolchain into the image by running
# .gondolin/post-build.sh inside the chroot. The Go version is pinned in
# that script; rebuild only when it or the Alpine package list changes.
set -eu

cd "$(dirname "$0")/.."

# Gondolin puts its build work dir in os.tmpdir(), which on macOS is
# /var/folders/.../T/. Docker Desktop on macOS frequently fails to share
# /var/folders with containers, which surfaces as:
#   /bin/sh: can't open '/work/build-in-container.sh': No such file or directory
# because the mounted temp dir appears empty. Point TMPDIR at a project-local
# dir (under /Users/..., which Docker shares reliably) so the work dir is
# visible inside the container.
# Gondolin reuses its work dir across builds without cleaning, so stale
# postBuild scripts from a prior run can shadow edited source files. Wipe
# it before each build. (.gitignore already ignores /.tmpwork/.)
TMPDIR="$PWD/.gondolin/.tmpwork"
rm -rf "$TMPDIR"
mkdir -p "$TMPDIR"
export TMPDIR

echo "==> Building Gondolin guest image..."
gondolin build \
  --config .gondolin/build-config.json \
  --output .gondolin/assets

echo "==> Verifying built assets..."
gondolin build --verify .gondolin/assets

echo "==> Done."
