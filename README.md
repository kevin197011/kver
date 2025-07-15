# kver

> Cross-language version manager (Go, Python, Node.js, Ruby, Java, ...)

## 项目简介

`kver` 是一个用 Go 编写的轻量级多语言版本管理工具，支持多版本共存、全局/项目级切换、插件扩展，适用于本地开发、CI/CD 和多平台环境。灵感来源于 asdf/nvm/pyenv，但支持更多语言和更强的可扩展性。

## 主要特性

- 跨语言、跨平台版本管理（Go, Python, Node.js, Ruby, Java 等）
- 多版本共存与快速切换
- 全局/项目级版本隔离（.kver 文件）
- 插件机制，易于扩展新语言
- 零依赖，无需 Python/bash
- 一键安装与自动环境激活

## 安装方法

### 方式一：一键脚本

```sh
curl -fsSL https://raw.githubusercontent.com/kevin197011/kver/main/deploy.sh | bash
```

### 方式二：手动下载

1. 前往 [Releases](https://github.com/kevin197011/kver/releases) 下载对应平台的二进制包
2. 放入 `~/.kver/bin` 并加入 PATH

## 常用命令

```sh
# 安装语言版本
kver install python 3.11.1
kver install ruby 3.2.2

# 查看已安装/可用版本
kver list python
kver list-remote ruby

# 切换版本
kver use nodejs 18.16.0
kver global go 1.21.0
kver local python 3.11.1

# 查看当前激活版本
kver current

# 激活环境变量（推荐在 shell 启动脚本中加入）
eval "$(kver activate)"
```

## 许可证

MIT License © 2025 kk
