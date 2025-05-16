#!/usr/bin/env bash
set -euo pipefail

# Install the Go toolchain version declared in go.mod and
# pre-download module dependencies. This script should be run
# while network access is available.

# Determine Go version from go.mod
GO_VERSION=$(awk '/^go /{print $2}' go.mod)

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64";;
    aarch64|arm64) ARCH="arm64";;
    *) echo "Unsupported architecture: $ARCH" >&2; exit 1;;
esac

TARBALL="go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
URL="https://go.dev/dl/${TARBALL}"

if ! command -v curl >/dev/null; then
    echo "curl is required" >&2
    exit 1
fi

# Download and install Go
curl -fsSL -o "$TARBALL" "$URL"
if [ "$EUID" -ne 0 ]; then
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$TARBALL"
else
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "$TARBALL"
fi
rm "$TARBALL"

export PATH="/usr/local/go/bin:$PATH"

go version

# Pre-fetch dependencies
if [ -f go.mod ]; then
    go mod download
fi

cat <<EOM

Go $GO_VERSION installed.
Dependencies downloaded.
EOM
