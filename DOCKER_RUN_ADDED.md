# Docker Run 命令已添加到部署文档

## ✅ 更新完成

已成功在 `DOCKER_DEPLOY.md` 中添加了完整的 Docker Run 部署指南！

## 📊 更新统计

- **原文件行数**: 266 行
- **更新后行数**: 873 行
- **新增内容**: 607 行
- **新增章节**: 2 个主要章节

## 📝 新增内容概览

### 方式一：Docker Compose 部署（推荐）
- 原有内容重组
- 标记为推荐方式

### 方式二：Docker Run 部署（新增）

#### 1. 快速开始
- ✅ 构建镜像命令
- ✅ 准备配置文件
- ✅ 基础运行命令
- ✅ Windows PowerShell/CMD 命令

#### 2. 参数详细说明
- ✅ 端口映射
- ✅ 配置文件挂载
- ✅ 环境变量
- ✅ 重启策略
- ✅ 网络模式
- ✅ 资源限制

#### 3. 高级用法
- ✅ Host 网络模式（推荐 WOL）
- ✅ 健康检查配置
- ✅ 日志管理
- ✅ 资源限制设置
- ✅ 时区配置
- ✅ 环境变量文件

#### 4. 常用管理命令
- ✅ 查看容器状态
- ✅ 容器控制命令
- ✅ 进入容器
- ✅ 更新容器

#### 5. 故障排查
- ✅ 容器启动失败
- ✅ 配置文件问题
- ✅ 网络问题
- ✅ 权限问题

#### 6. 完整示例脚本
- ✅ Linux/Mac Bash 脚本
- ✅ Windows PowerShell 脚本
- ✅ 自动化部署流程

#### 7. 对比分析
- ✅ Docker Run vs Docker Compose
- ✅ 使用场景推荐
- ✅ 优缺点对比

#### 8. 迁移指南
- ✅ 从 Docker Run 迁移到 Docker Compose

## 🎯 关键命令示例

### 基础命令

```bash
# 构建镜像
docker build -t home-wol:latest .

# 运行容器
docker run -d \
  --name home-wol \
  -p 1323:1323 \
  -v $(pwd)/config.json:/home/config.json:ro \
  -e BROADCAST_ADDRESS=192.168.100.255 \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  home-wol:latest
```

### Windows 命令

**PowerShell**:
```powershell
docker run -d `
  --name home-wol `
  -p 1323:1323 `
  -v "${PWD}/config.json:/home/config.json:ro" `
  -e BROADCAST_ADDRESS=192.168.100.255 `
  -e TZ=Asia/Shanghai `
  --restart unless-stopped `
  home-wol:latest
```

**CMD**:
```cmd
docker run -d ^
  --name home-wol ^
  -p 1323:1323 ^
  -v "%CD%/config.json:/home/config.json:ro" ^
  -e BROADCAST_ADDRESS=192.168.100.255 ^
  -e TZ=Asia/Shanghai ^
  --restart unless-stopped ^
  home-wol:latest
```

### Host 网络模式（推荐 WOL）

```bash
docker run -d \
  --name home-wol \
  --network host \
  -v $(pwd)/config.json:/home/config.json:ro \
  -e BROADCAST_ADDRESS=192.168.100.255 \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  home-wol:latest
```

### 带健康检查

```bash
docker run -d \
  --name home-wol \
  -p 1323:1323 \
  -v $(pwd)/config.json:/home/config.json:ro \
  -e BROADCAST_ADDRESS=192.168.100.255 \
  --health-cmd "wget --quiet --tries=1 --spider http://localhost:1323/ || exit 1" \
  --health-interval 30s \
  --health-timeout 10s \
  --health-retries 3 \
  --health-start-period 10s \
  --restart unless-stopped \
  home-wol:latest
```

### 资源限制

```bash
docker run -d \
  --name home-wol \
  -p 1323:1323 \
  -v $(pwd)/config.json:/home/config.json:ro \
  -e BROADCAST_ADDRESS=192.168.100.255 \
  --cpus="0.5" \
  --memory="128m" \
  --memory-swap="256m" \
  --restart unless-stopped \
  home-wol:latest
```

### 日志管理

```bash
docker run -d \
  --name home-wol \
  -p 1323:1323 \
  -v $(pwd)/config.json:/home/config.json:ro \
  -e BROADCAST_ADDRESS=192.168.100.255 \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  --restart unless-stopped \
  home-wol:latest
```

## 📚 完整目录结构

```
DOCKER_DEPLOY.md (873 行)
│
├── 方式一：Docker Compose 部署（推荐）
│   ├── 快速开始
│   ├── 配置文件管理
│   ├── 环境变量
│   ├── 网络配置
│   ├── 高级配置
│   ├── 生产环境建议
│   ├── 更新部署
│   └── 完全卸载
│
└── 方式二：Docker Run 部署 ⭐ 新增
    ├── 快速开始
    │   ├── 构建镜像
    │   ├── 准备配置文件
    │   └── 运行容器（Linux/Windows）
    │
    ├── 参数说明
    │   ├── 端口映射
    │   ├── 配置文件挂载
    │   ├── 环境变量
    │   ├── 重启策略
    │   ├── 网络模式
    │   └── 资源限制
    │
    ├── 高级用法
    │   ├── Host 网络模式
    │   ├── 健康检查
    │   ├── 日志配置
    │   ├── 资源限制
    │   └── 环境变量文件
    │
    ├── 常用管理命令
    │   ├── 查看容器状态
    │   ├── 容器控制
    │   ├── 进入容器
    │   └── 更新容器
    │
    ├── 故障排查
    │   ├── 容器启动失败
    │   ├── 配置文件问题
    │   ├── 网络问题
    │   └── 权限问题
    │
    ├── 完整示例脚本
    │   ├── Linux/Mac Bash 脚本
    │   └── Windows PowerShell 脚本
    │
    ├── Docker Run vs Docker Compose 对比
    └── 迁移指南
```

## 🎯 使用建议

### 选择 Docker Run 如果你：
- ✅ 只需要运行一个容器
- ✅ 想要快速测试
- ✅ 不需要复杂的配置
- ✅ 习惯使用命令行

### 选择 Docker Compose 如果你：
- ✅ 部署生产环境
- ✅ 需要多容器协作
- ✅ 团队协作开发
- ✅ 需要版本控制配置
- ✅ 需要自动化部署

## 🔧 快速参考

### Docker Run 常用命令

```bash
# 查看运行状态
docker ps | grep home-wol

# 查看日志
docker logs -f home-wol

# 停止容器
docker stop home-wol

# 启动容器
docker start home-wol

# 重启容器
docker restart home-wol

# 删除容器
docker rm -f home-wol

# 进入容器
docker exec -it home-wol sh

# 查看资源使用
docker stats home-wol

# 更新容器
docker stop home-wol && docker rm home-wol
docker run -d ... home-wol:latest
```

## 📖 完整文档

详细的 Docker Run 使用指南已添加到：
```
D:\code\home-wol\DOCKER_DEPLOY.md
```

**新增内容**：
- 第 266-873 行
- 完整的 Docker Run 部署指南
- 607 行新增内容
- 涵盖所有使用场景

**文档已更新完成！** 🎉
