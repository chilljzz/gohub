#!/usr/bin/env bash
set -euo pipefail

# 自动进入 GoHub 根目录
cd "$(dirname "$0")/.."

ENV_FILE=".env.test.local"

# 检查配置文件
if [ ! -f "$ENV_FILE" ]; then
    echo "Missing $ENV_FILE"
    exit 1
fi

# 读取并导出测试环境变量
set -a
source "$ENV_FILE"
set +a

# 检查 DSN 是否存在
: "${GOHUB_TEST_MYSQL_DSN:?missing test DSN}"

# 启动独立的测试 MySQL
docker compose \
    -p gohub-it \
    -f compose.test.yaml \
    --env-file "$ENV_FILE" \
    up -d --wait mysql-test

