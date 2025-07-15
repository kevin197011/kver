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

# 确保 env.d 目录存在
mkdir -p "$HOME/.kver/env.d"

# 写入 env.sh 统一 source 机制
cat > "$HOME/.kver/env.sh" <<'EOF'
# kver multi-language env loader
env_d="$HOME/.kver/env.d"
if [ -d "$env_d" ]; then
  for f in "$env_d"/*.sh; do
    [ -f "$f" ] && source "$f"
  done
fi
EOF

echo "[kver] ~/.kver/env.sh updated to load all ~/.kver/env.d/*.sh"

# 自动写入 PATH 到 shell 配置
SHELL_NAME="$(basename "$SHELL")"
if [[ "$SHELL_NAME" == "zsh" ]]; then
  RC_FILE="$HOME/.zshrc"
else
  RC_FILE="$HOME/.bashrc"
fi

EXPORT_LINE="[ -f \"$HOME/.kver/env.sh\" ] && source \"$HOME/.kver/env.sh\""
if ! grep -Fxq "$EXPORT_LINE" "$RC_FILE" 2>/dev/null; then
  echo "$EXPORT_LINE" >> "$RC_FILE"
  echo "[kver] Added env.sh source to $RC_FILE"
else
  echo "[kver] env.sh already sourced in $RC_FILE"
fi

# 自动写入 kver shell function，彻底防重复且可升级
KVER_FUNC_MARK_START='# >>> kver shell function >>>'
KVER_FUNC_MARK_END='# <<< kver shell function <<<'

# 删除旧区块
if grep -Fq "$KVER_FUNC_MARK_START" "$RC_FILE" 2>/dev/null; then
  sed -i "/$KVER_FUNC_MARK_START/,/$KVER_FUNC_MARK_END/d" "$RC_FILE"
fi

# 追加最新 function
cat <<'EOF' >> "$RC_FILE"
# >>> kver shell function >>>
kver() {
  command kver "$@"
  if [[ "$1" == "use" || "$1" == "global" ]]; then
    if [ -f "$HOME/.kver/env.sh" ]; then
      source "$HOME/.kver/env.sh"
      echo -e "\033[1;32m[kver] 环境已自动激活 (当前 shell)\033[0m"
    fi
  fi
}
# <<< kver shell function <<<
EOF

echo "[kver] kver shell function updated in $RC_FILE"

echo "\n[kver] Install complete! Try: kver --help"