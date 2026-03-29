#!/bin/sh
set -e

REPO="daghlny/scout_cli"
BINARY="scout_cli"
INSTALL_DIR="/usr/local/bin"

# Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS. For Windows, download from GitHub Releases."; exit 1 ;;
esac

# Get latest version
LATEST=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest version. Check https://github.com/${REPO}/releases"
  exit 1
fi

FILE="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILE}"

echo "Downloading ${BINARY} ${LATEST} for ${OS}/${ARCH}..."
TMP=$(mktemp -d)
curl -sL "$URL" -o "${TMP}/${FILE}"
tar -xzf "${TMP}/${FILE}" -C "$TMP"

echo "Installing to ${INSTALL_DIR}/${BINARY}..."
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi
chmod +x "${INSTALL_DIR}/${BINARY}"

rm -rf "$TMP"
echo "Done! Run '${BINARY}' to play."
