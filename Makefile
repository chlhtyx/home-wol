.PHONY: help build up down restart logs ps test clean install-config install-env init dev rebuild shell update

help: ## 显示帮助信息
	@echo "Home WOL - Makefile 命令"
	@echo ""
	@echo "使用方法: make [目标]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## 构建 Docker 镜像
	docker compose build

up: ## 启动服务
	docker compose up -d

down: ## 停止服务
	docker compose down

restart: ## 重启服务
	docker compose restart

logs: ## 查看日志（实时）
	docker compose logs -f

ps: ## 查看服务状态
	docker compose ps

test: ## 测试服务
	@echo "测试服务健康检查..."
	@curl -fsS "http://localhost:1323/healthz"
	@echo ""
	@echo "请登录管理端后测试设备唤醒"

clean: ## 清理容器和镜像
	docker compose down
	docker rmi home-wol_home-wol

install-config: ## 安装配置文件
	@if [ ! -f "config.json" ]; then \
		cp config.example.json config.json && \
		echo "✅ 已创建配置文件: config.json"; \
	else \
		echo "⚠️  配置文件已存在"; \
	fi

install-env: ## 安装环境变量文件
	@if [ ! -f ".env" ]; then \
		cp .env.example .env; \
		echo "✅ 已创建环境变量文件: .env"; \
	else \
		echo "⚠️  环境变量文件已存在"; \
	fi

init: install-config install-env ## 初始化配置
	@echo "✅ 配置文件已准备就绪"
	@echo "⚠️  请编辑以下文件："
	@echo "   - config.json (设备配置)"
	@echo "   - .env (环境变量)"

dev: ## 开发模式启动
	docker compose up

rebuild: ## 重新构建并启动
	docker compose up -d --build

shell: ## 进入容器 Shell
	docker exec -it home-wol sh

update: ## 更新并重启
	git pull
	docker compose build
	docker compose up -d
