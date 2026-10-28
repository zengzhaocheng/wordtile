#!/usr/bin/env bash
# WordTile 生产部署脚本（在服务器上执行）
# 用法：./deploy.sh
set -euo pipefail

PROD_FILE="docker-compose.prod.yml"

echo "==> 检查 Docker ..."
if ! command -v docker >/dev/null 2>&1; then
  echo "未检测到 Docker，正在安装 ..."
  curl -fsSL https://get.docker.com | sh
fi
if ! docker compose version >/dev/null 2>&1; then
  echo "错误：未检测到 docker compose 插件，请先安装 Docker Compose v2。" >&2
  exit 1
fi

echo "==> 准备环境变量 ..."
if [ ! -f .env ]; then
  cp .env.example .env
  # 生成随机密钥
  gen() {
    if command -v openssl >/dev/null 2>&1; then
      openssl rand -base64 32 | tr -d '\n'
    else
      head -c 32 /dev/urandom | base64 | tr -d '\n'
    fi
  }
  if [[ "$(uname)" == "Darwin" ]]; then
    sed -i '' "s|请替换成随机32字节字符串|$(gen)|" .env
    sed -i '' "s|请改成随机强密码|$(gen | tr -d '/+=' | head -c 24)|" .env
  else
    sed -i "s|请替换成随机32字节字符串|$(gen)|" .env
    sed -i "s|请改成随机强密码|$(gen | tr -d '/+=' | head -c 24)|" .env
  fi
  echo "已生成 .env（含随机 JWT_SECRET 与数据库密码）"
else
  echo ".env 已存在，跳过生成"
fi

echo "==> 构建并启动服务 ..."
docker compose -f "$PROD_FILE" up -d --build

echo "==> 等待后端健康检查 ..."
for i in $(seq 1 30); do
  if docker compose -f "$PROD_FILE" exec -T backend wget -qO- http://localhost:8080/healthz >/dev/null 2>&1; then
    echo "后端已就绪"
    break
  fi
  sleep 2
done

DOMAIN=$(grep -E '^DOMAIN=' .env | cut -d '=' -f2-)
echo
echo "🎉 部署完成"
if [ -n "$DOMAIN" ] && [ "$DOMAIN" != "wordtile.example.com" ]; then
  echo "   访问地址：https://${DOMAIN}"
  echo "   （请确保域名已解析到本机公网 IP，且防火墙已放行 80/443）"
else
  echo "   请先把 .env 里的 DOMAIN 改成你的真实域名，然后重新执行 ./deploy.sh"
fi
echo "   演示账号：demo / demo1234"
echo
echo "常用命令："
echo "   查看状态：docker compose -f $PROD_FILE ps"
echo "   查看日志：docker compose -f $PROD_FILE logs -f backend"
echo "   停止服务：docker compose -f $PROD_FILE down"
