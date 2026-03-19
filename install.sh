#!/bin/sh
# wall-vault installer
# Usage: curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh

set -e

REPO="sookmook/wall-vault"
INSTALL_DIR="$HOME/.local/bin"
BINARY="wall-vault"

# --- detect OS and arch ---
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Error: unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux)  PLATFORM="linux-$ARCH" ;;
  darwin) PLATFORM="darwin-$ARCH" ;;
  *)
    echo "Error: unsupported OS: $OS"
    echo "For Windows, download wall-vault-windows-amd64.exe from:"
    echo "  https://github.com/$REPO/releases/latest"
    exit 1
    ;;
esac

# --- resolve latest version ---
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Error: could not determine latest version."
  echo "Check: https://github.com/$REPO/releases"
  exit 1
fi

URL="https://github.com/$REPO/releases/download/$LATEST/wall-vault-$PLATFORM"

echo "wall-vault $LATEST ($PLATFORM)"
echo "Downloading from: $URL"

# --- install ---
mkdir -p "$INSTALL_DIR"
curl -fSL "$URL" -o "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"

echo ""
echo "Installed: $INSTALL_DIR/$BINARY"

# --- PATH check ---
case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    ;;
  *)
    echo ""
    echo "Add to PATH (add this line to ~/.bashrc or ~/.zshrc):"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "Or run now:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    ;;
esac

echo ""
echo "Run setup wizard:  wall-vault setup"
echo "Start (proxy+vault): wall-vault start"
