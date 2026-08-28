#!/bin/sh
# rootfsInitExtra: appended to the guest rootfs init before sandboxd starts.
# Runs on every boot, before any exec commands — more reliable than
# /etc/profile.d (which only sources in login shells).
#
# Gondolin MITMs HTTPS with its own CA (/etc/gondolin/mitm/ca.crt). curl and
# Go's crypto/x509 both trust the system CA bundle
# (/etc/ssl/certs/ca-certificates.crt); curl via SSL_CERT_FILE, Go via
# SystemCertPool. Append the MITM CA to the system bundle (idempotent) so
# `go mod` fetches and other TLS traffic trust Gondolin's MITM.

if [ -f /etc/gondolin/mitm/ca.crt ] && [ -f /etc/ssl/certs/ca-certificates.crt ]; then
  _marker=$(sed -n '2p' /etc/gondolin/mitm/ca.crt 2>/dev/null)
  if [ -n "$_marker" ] && ! grep -qF "$_marker" /etc/ssl/certs/ca-certificates.crt 2>/dev/null; then
    cat /etc/gondolin/mitm/ca.crt >> /etc/ssl/certs/ca-certificates.crt
  fi
  unset _marker
fi
