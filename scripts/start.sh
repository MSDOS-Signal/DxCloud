#!/usr/bin/env sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$ROOT"
REGION="${1:-${DX_REGION:-cn}}"

if [ "$REGION" != "cn" ] && [ "$REGION" != "global" ]; then
  echo "区域参数只支持 cn 或 global，当前使用 cn"
  REGION="cn"
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "未检测到 Docker。请先安装 Docker Desktop: https://www.docker.com/products/docker-desktop/"
  echo "安装后等待 Docker Desktop 的 Engine running，再重新运行本脚本。"
  exit 1
fi

if ! docker version >/dev/null 2>&1; then
  echo "Docker 已安装但服务未启动，请打开 Docker Desktop 并等待 Engine running。"
  exit 1
fi

if [ ! -f .env ]; then
  cp .env.example .env
fi

set_env() {
  name="$1"
  value="$2"
  if grep -Eq "^${name}=" .env; then
    tmp="$(mktemp)"
    sed "s|^${name}=.*|${name}=${value}|" .env >"$tmp"
    mv "$tmp" .env
  else
    printf '\n%s=%s\n' "$name" "$value" >>.env
  fi
}

if [ "$REGION" = "cn" ]; then
  MIRROR="hub.rat.dev"
  set_env MYSQL_IMAGE "$MIRROR/library/mysql:8.0"
  set_env REDIS_IMAGE "$MIRROR/library/redis:7-alpine"
  set_env REGISTRY_IMAGE "$MIRROR/library/registry:2.8"
  set_env TRAEFIK_IMAGE "$MIRROR/library/traefik:v3"
  set_env GO_BUILD_IMAGE "$MIRROR/library/golang:1.25-alpine"
  set_env ALPINE_RUNTIME_IMAGE "$MIRROR/library/alpine:3.20"
  set_env NODE_IMAGE "$MIRROR/library/node:20-alpine"
  echo "当前区域：中国大陆，已写入国内镜像加速前缀。"
else
  set_env MYSQL_IMAGE "mysql:8.0"
  set_env REDIS_IMAGE "redis:7-alpine"
  set_env REGISTRY_IMAGE "registry:2.8"
  set_env TRAEFIK_IMAGE "traefik:v3"
  set_env GO_BUILD_IMAGE "golang:1.25-alpine"
  set_env ALPINE_RUNTIME_IMAGE "alpine:3.20"
  set_env NODE_IMAGE "node:20-alpine"
  echo "当前区域：非中国大陆，使用 Docker Hub 官方基础镜像。"
fi

docker compose up -d --build
echo ""
echo "启动完成，请在浏览器打开 http://localhost"
echo "初始管理员：admin / Admin@123456（首次登录后请立即修改）"
