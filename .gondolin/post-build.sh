#!/bin/sh
# Build-time script run as a single postBuild command inside the rootfs
# chroot (invoked by build.sh via `gondolin build`). Installs the Go toolchain
# so the VM is ready to build barista on first boot.
#
# Gondolin runs each postBuild.commands array element as a separate
# /bin/sh -lc invocation, so this script keeps every export in one process.
#
# We deliberately do NOT pre-fetch modules or warm the build cache here:
# GOCACHE and GOMODCACHE are host-mounted (see bash.sh) and start empty,
# then fill at runtime and persist across VM restarts. Baking them into
# the image rootfs would be wasted work — the host mount shadows the
# rootfs on every boot.
set -eux

# Install Go into /opt/go (GOROOT), the only Go on the image. The version is
# pinned here, in this script. To bump Go, change GO_VERSION and rebuild.
# (mise is not used: mise's Go plugin has a cold-install bug in this build
# chroot — its post-extract `go version` verify execs the final install path
# before renaming temp -> final, ENOENTs, and rolls back. A direct tarball
# extract into /opt/go has no such moving parts and keeps the version pin in
# exactly one place.)
GO_VERSION=1.27.0
DEST=/opt/go
TARBALL_URL=https://dl.google.com/go/go${GO_VERSION}.linux-arm64.tar.gz

if [ ! -x "$DEST/bin/go" ]; then
  echo "post-build: installing Go ${GO_VERSION} into $DEST..."
  rm -rf "$DEST"
  mkdir -p "$DEST"
  curl -fsSLo /tmp/go.tar.gz "${TARBALL_URL}"
  curl -fsSLo /tmp/go.tar.gz.sha256 "${TARBALL_URL}.sha256"
  _expected=$(awk '{print $1}' /tmp/go.tar.gz.sha256)
  _actual=$(sha256sum /tmp/go.tar.gz | awk '{print $1}')
  test "$_expected" = "$_actual"
  # Tarball has a top-level go/ dir; strip it so $DEST is GOROOT.
  tar -xf /tmp/go.tar.gz -C "$DEST" --strip-components=1
  rm -f /tmp/go.tar.gz /tmp/go.tar.gz.sha256
fi

export PATH="$DEST/bin:$PATH"
go version

# Clean up build-only state to slim the image. build-base is KEPT so future
# cgo deps can compile at runtime.
rm -rf /tmp/post-build.sh /var/cache/apk/* /root/.cache
