
#!/bin/bash

cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." || exit 1

echo "=================================="
echo "  Home WOL 部署检查脚本"
echo "=================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_passed=true

# 检查 Docker
echo "🔍 检查 Docker..."
if command -v docker &> /dev/null; then
    echo -e "${GREEN}✅ Docker 已安装${NC}"
    docker --version
else
    echo -e "${RED}❌ Docker 未安装${NC}"
    check_passed=false
fi
echo ""

# 检查 Docker Compose
echo "🔍 检查 Docker Compose..."
if docker compose version &> /dev/null; then
    echo -e "${GREEN}✅ Docker Compose 已安装${NC}"
    docker compose version
else
    echo -e "${RED}❌ Docker Compose 未安装${NC}"
    check_passed=false
fi
echo ""

# 检查配置文件
echo "🔍 检查配置文件..."
if [ -f "config.json" ]; then
    echo -e "${GREEN}✅ 配置文件存在${NC}"
    # 验证 JSON 格式
    if command -v jq &> /dev/null; then
        if jq empty config.json 2>/dev/null; then
            echo -e "${GREEN}✅ 配置文件格式正确${NC}"
        else
            echo -e "${RED}❌ 配置文件格式错误${NC}"
            check_passed=false
        fi
    else
        echo -e "${YELLOW}⚠️  jq 未安装，跳过格式验证${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  配置文件不存在，运行 'make init' 创建${NC}"
fi
echo ""

# 检查环境变量文件
echo "🔍 检查环境变量文件..."
if [ -f ".env" ]; then
    echo -e "${GREEN}✅ 环境变量文件存在${NC}"
else
    echo -e "${YELLOW}⚠️  环境变量文件不存在，运行 'make init' 创建${NC}"
fi
echo ""

# 广播地址由 Web 管理页面配置，不再要求 .env 提供。
echo "广播地址请在 Web 管理页面的网络设置中查看和修改。"

# 检查端口占用
echo "🔍 检查端口占用..."
PORT=1323
if netstat -tuln 2>/dev/null | grep -q ":$PORT "; then
    echo -e "${YELLOW}⚠️  端口 $PORT 已被占用${NC}"
else
    echo -e "${GREEN}✅ 端口 $PORT 可用${NC}"
fi
echo ""

# 检查 Docker 容器状态
echo "🔍 检查容器状态..."
if docker ps --format '{{.Names}}' | grep -q "home-wol"; then
    echo -e "${GREEN}✅ 容器正在运行${NC}"
    docker ps --filter "name=home-wol" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
else
    echo -e "${YELLOW}⚠️  容器未运行${NC}"
    echo "运行 'docker compose up -d' 启动服务"
fi
echo ""

# 测试服务
echo "🔍 测试服务..."
if command -v curl &> /dev/null; then
    if curl -fsS http://localhost:1323/healthz > /dev/null; then
        echo -e "${GREEN}✅ 服务可访问${NC}"
    else
        echo -e "${YELLOW}⚠️  服务不可访问${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  curl 未安装，跳过服务测试${NC}"
fi
echo ""

# 总结
echo "=================================="
if [ "$check_passed" = true ]; then
    echo -e "${GREEN}✅ 所有检查通过！${NC}"
    echo ""
    echo "启动服务："
    echo "  docker compose up -d"
    echo ""
    echo "或使用 Makefile："
    echo "  make up"
else
    echo -e "${RED}❌ 检查未通过，请解决上述问题${NC}"
    exit 1
fi
echo "=================================="
