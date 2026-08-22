#!/bin/bash
# 评测专用：Go-only 打包（无前端构建步骤，web 由 Go 同进程托管）
set -euo pipefail
IMG="${1:?image}"
ARCH="${2:?arch}"
WORKDIR="$(cd "$(dirname "$0")" && pwd)"
cd "$WORKDIR"
case "$ARCH" in
  linux/arm64) export GOARCH=arm64 ;;
  linux/amd64) export GOARCH=amd64 ;;
  *) echo "unsupported arch $ARCH" >&2; exit 1 ;;
esac
docker buildx build --platform "$ARCH" -f benzhi.Dockerfile -t "$IMG:latest" .
