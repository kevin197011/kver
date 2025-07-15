#!/usr/bin/env bash

set -e

REPO="kevin197011/kver"  # 替换为你的 GitHub 仓库
INSTALL_DIR="${HOME}/.kver/bin"
BIN_NAME="kver"
VERSION="latest"

# 解析参数
while [[ $# -gt 0 ]]; do
  case $1 in
    --version)
      VERSION="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# 检测系统架构
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

PKG="${BIN_NAME}-${OS}-${ARCH}"

# 获取下载链接
if [[ "$VERSION" == "latest" ]]; then
  VERSION=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep tag_name | cut -d '"' -f4)
fi

URL="https://github.com/${REPO}/releases/download/${VERSION}/${PKG}"

echo "[kver] Downloading $PKG ($VERSION) from $URL"
mkdir -p "$INSTALL_DIR"
curl -fsSL "$URL" -o "$INSTALL_DIR/$BIN_NAME"
chmod +x "$INSTALL_DIR/$BIN_NAME"

# 检查 PATH
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo
  echo "[kver] Please add $INSTALL_DIR to your PATH:"
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
  echo "  # Add above line to your ~/.bashrc or ~/.zshrc"
fi

echo
echo "[kver] Install complete! Try: kver --help"