# Docker 部署指南

本指南提供两种部署方式：
1. **Docker Compose**（推荐）- 适合生产环境
2. **Docker Run** - 适合快速测试和简单部署

---

## 方式一：Docker Compose 部署（推荐）

### 快速开始

### 1. 准备配置文件

```bash
# 复制配置文件模板
cp config/config.json.template config/config.json

# 编辑配置文件，添加你的设备信息
# 编辑 config/config.json
```

### 2. 修改 docker-compose.yml（可选）

根据你的网络环境修改 `docker-compose.yml` 中的环境变量：

```yaml
environment:
  - BROADCAST_ADDRESS=192.168.1.255  # 修改为你的网络广播地址
```

### 3. 启动服务

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重启服务
docker-compose restart
```

### 4. 验证服务

```bash
# 测试生成验证码
curl "http://localhost:1323/generate_code?mac=E0:E1:A9:10:87:06"

# 测试唤醒设备
curl "http://localhost:1323/wol?mac=E0:E1:A9:10:87:06&auth_code=123456"
```

## 配置文件管理

### 配置文件位置

```
D:\code\home-wol\config\config.json
```

### 重新加载配置

修改配置文件后，只需重启容器：

```bash
docker-compose restart
```

### 配置文件格式

```json
{
  "devices": [
    {
      "mac": "E0:E1:A9:10:87:06",
      "name": "设备名称",
      "secret": "JBSWY3DPEHPK3PXP"
    }
  ]
}
```

## 环境变量

在 `docker-compose.yml` 中可以配置以下环境变量：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| CONFIG_PATH | 配置文件路径 | /home/config/config.json |
| BROADCAST_ADDRESS | 广播地址 | 192.168.1.255 |
| TZ | 时区 | Asia/Shanghai |

## 网络配置

### Bridge 模式（默认）

适用于大多数场景：

```yaml
services:
  home-wol:
    ports:
      - "1323:1323"
    networks:
      - wol-network
```

### Host 模式（推荐用于 WOL）

WOL 需要发送广播包，使用 host 模式可以避免网络限制：

```yaml
services:
  home-wol:
    network_mode: "host"
    # 不需要配置 ports
```

修改后重启：

```bash
docker-compose down
docker-compose up -d
```

## 目录结构

```
D:\code\home-wol\
├── docker-compose.yml           # Docker Compose 配置
├── Dockerfile                   # Docker 镜像构建文件
├── .dockerignore               # Docker 忽略文件
├── config/                     # 配置文件目录（挂载到容器）
│   ├── config.json.template    # 配置文件模板
│   └── config.json             # 实际配置文件（需要创建）
├── main.go                     # 主程序
├── common/                     # 公共模块
├── service/                    # 服务模块
└── ...
```

## 常见问题

### 1. 容器启动失败

检查日志：

```bash
docker-compose logs home-wol
```

常见原因：
- 配置文件不存在：确保 `config/config.json` 存在
- 配置文件格式错误：检查 JSON 格式是否正确
- 端口冲突：修改 `docker-compose.yml` 中的端口映射

### 2. 无法唤醒设备

检查项：
- [ ] 广播地址是否正确
- [ ] 容器网络模式（尝试使用 host 模式）
- [ ] 设备是否支持 WOL
- [ ] 设备 MAC 地址是否正确

### 3. 配置文件修改后不生效

重启容器：

```bash
docker-compose restart
```

### 4. 查看容器内配置文件

```bash
# 进入容器
docker exec -it home-wol sh

# 查看配置文件
cat /home/config/config.json

# 退出容器
exit
```

## 高级配置

### 使用自定义网络

```yaml
services:
  home-wol:
    networks:
      - custom-network

networks:
  custom-network:
    driver: bridge
    ipam:
      config:
        - subnet: 192.168.1.0/24
```

### 设置资源限制

```yaml
services:
  home-wol:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 128M
        reservations:
          cpus: '0.25'
          memory: 64M
```

### 日志配置

```yaml
services:
  home-wol:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## 生产环境建议

1. **使用 secrets 管理敏感信息**（Docker Swarm）
2. **配置 HTTPS**（使用反向代理如 Nginx）
3. **设置防火墙规则**
4. **定期更新镜像**
5. **备份配置文件**

## 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建镜像
docker-compose build

# 重启服务
docker-compose up -d
```

## 完全卸载

```bash
# 停止并删除容器
docker-compose down

# 删除镜像
docker rmi home-wol_home-wol

# 删除配置文件（可选）
# rm -rf config/
```

---

## 方式二：Docker Run 部署

如果你不想使用 Docker Compose，可以直接使用 `docker run` 命令。

### 快速开始

#### 1. 构建镜像

```bash
# 克隆或进入项目目录
cd /path/to/home-wol

# 构建 Docker 镜像
docker build -t home-wol:latest .
```

#### 2. 准备配置文件

```bash
# 确保配置文件存在
cat config.json
```

配置文件示例：
```json
{
  "devices": [
    {
      "mac": "E0:E1:A9:10:87:06",
      "name": "我的电脑",
      "secret": "JBSWY3DPEHPK3PXP"
    }
  ]
}
```

#### 3. 运行容器

**基础运行命令**：

```bash
docker run -d \
  --name home-wol \
  -p 1323:1323 \
  -v $(pwd)/config.json:/home/config.json:ro \
  -e BROADCAST_ADDRESS=192.168.100.255 \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  home-wol:latest
```

**Windows PowerShell**：

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

**Windows CMD**：

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

### 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `-d` | 后台运行 | - |
| `--name` | 容器名称 | `home-wol` |
| `-p` | 端口映射 | `1323:1323` |
| `-v` | 挂载配置文件 | `./config.json:/home/config.json:ro` |
| `-e` | 环境变量 | `BROADCAST_ADDRESS=...` |
| `--restart` | 重启策略 | `unless-stopped` |
| `--network` | 网络模式 | `host`（可选） |

### 完整参数说明

#### 端口映射

```bash
# 映射到不同端口
-p 8080:1323

# 映射到所有网络接口
-p 0.0.0.0:1323:1323

# 只映射到 localhost
-p 127.0.0.1:1323:1323
```

#### 配置文件挂载

```bash
# 只读挂载（推荐）
-v $(pwd)/config.json:/home/config.json:ro

# 读写挂载（不推荐）
-v $(pwd)/config.json:/home/config.json:rw

# 使用绝对路径
-v /absolute/path/config.json:/home/config.json:ro
```

#### 环境变量

```bash
# 配置文件路径
-e CONFIG_PATH=/home/config.json

# 广播地址
-e BROADCAST_ADDRESS=192.168.100.255

# 时区
-e TZ=Asia/Shanghai

# Go 运行模式
-e GIN_MODE=release
```

#### 重启策略

```bash
# 总是重启
--restart always

# 退出时重启（默认）
--restart on-failure

# 除非手动停止才重启
--restart unless-stopped

# 不自动重启
--restart no
```

#### 网络模式

```bash
# Bridge 模式（默认）
# 无需额外参数

# Host 模式（推荐 WOL）
--network host

# 自定义网络
--network my-network

# 暴露所有端口
--network bridge --publish-all
```

#### 资源限制

```bash
# 限制 CPU 使用
--cpus="0.5"

# 限制内存
--memory="128m"

# 设置内存交换
--memory-swap="256m"

# 设置 CPU 权重（1024为基准）
--cpu-shares=512
```

### 高级用法

#### 1. 使用 Host 网络模式（推荐 WOL）

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

**注意**：使用 host 网络模式时，不需要 `-p` 参数。

#### 2. 健康检查

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

#### 3. 日志配置

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

#### 4. 资源限制

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

#### 5. 添加时区数据卷

```bash
docker run -d \
  --name home-wol \
  -p 1323:1323 \
  -v $(pwd)/config.json:/home/config.json:ro \
  -v /etc/timezone:/etc/timezone:ro \
  -v /etc/localtime:/etc/localtime:ro \
  -e BROADCAST_ADDRESS=192.168.100.255 \
  --restart unless-stopped \
  home-wol:latest
```

#### 6. 使用环境变量文件

```bash
# 创建环境变量文件
cat > wol.env << EOF
BROADCAST_ADDRESS=192.168.100.255
TZ=Asia/Shanghai
GIN_MODE=release
EOF

# 使用环境变量文件
docker run -d \
  --name home-wol \
  -p 1323:1323 \
  -v $(pwd)/config.json:/home/config.json:ro \
  --env-file wol.env \
  --restart unless-stopped \
  home-wol:latest
```

### 常用管理命令

#### 查看容器状态

```bash
# 查看运行状态
docker ps | grep home-wol

# 查看详细信息
docker inspect home-wol

# 查看资源使用
docker stats home-wol

# 查看日志
docker logs -f home-wol

# 查看最近 100 行日志
docker logs --tail 100 home-wol
```

#### 容器控制

```bash
# 停止容器
docker stop home-wol

# 启动容器
docker start home-wol

# 重启容器
docker restart home-wol

# 删除容器（需要先停止）
docker stop home-wol && docker rm home-wol

# 强制删除运行中的容器
docker rm -f home-wol
```

#### 进入容器

```bash
# 使用 sh 进入
docker exec -it home-wol sh

# 查看环境变量
docker exec home-wol env

# 查看配置文件
docker exec home-wol cat /home/config.json

# 测试网络连接
docker exec home-wol wget -O- http://localhost:1323/
```

#### 更新容器

```bash
# 1. 停止并删除旧容器
docker stop home-wol && docker rm home-wol

# 2. 重新构建镜像（如果需要）
docker build -t home-wol:latest .

# 3. 运行新容器
docker run -d \
  --name home-wol \
  -p 1323:1323 \
  -v $(pwd)/config.json:/home/config.json:ro \
  -e BROADCAST_ADDRESS=192.168.100.255 \
  --restart unless-stopped \
  home-wol:latest
```

### 故障排查

#### 1. 容器启动失败

```bash
# 查看容器日志
docker logs home-wol

# 查看退出代码
docker ps -a | grep home-wol

# 查看详细错误
docker inspect home-wol | grep -A 10 State
```

#### 2. 配置文件问题

```bash
# 检查配置文件是否存在
docker exec home-wol ls -la /home/config.json

# 查看配置文件内容
docker exec home-wol cat /home/config.json

# 验证 JSON 格式
cat config.json | python -m json.tool
```

#### 3. 网络问题

```bash
# 检查端口映射
docker port home-wol

# 查看容器网络配置
docker inspect home-wol | grep -A 20 NetworkSettings

# 测试容器内服务
docker exec home-wol wget -O- http://localhost:1323/
```

#### 4. 权限问题

```bash
# 检查文件权限
ls -la config.json

# 修改文件权限（如果需要）
chmod 644 config.json

# 使用当前用户运行（Linux/Mac）
docker run -d \
  --name home-wol \
  -p 1323:1323 \
  -v $(pwd)/config.json:/home/config.json:ro \
  -e BROADCAST_ADDRESS=192.168.100.255 \
  -u $(id -u):$(id -g) \
  home-wol:latest
```

### 完整示例脚本

#### Linux/Mac 脚本

```bash
#!/bin/bash

# 配置变量
IMAGE_NAME="home-wol:latest"
CONTAINER_NAME="home-wol"
CONFIG_FILE="$(pwd)/config.json"
BROADCAST_ADDR="192.168.100.255"
PORT="1323"

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    exit 1
fi

# 停止并删除旧容器
echo "停止旧容器..."
docker stop $CONTAINER_NAME 2>/dev/null
docker rm $CONTAINER_NAME 2>/dev/null

# 运行新容器
echo "启动新容器..."
docker run -d \
  --name $CONTAINER_NAME \
  -p ${PORT}:1323 \
  -v ${CONFIG_FILE}:/home/config.json:ro \
  -e BROADCAST_ADDRESS=${BROADCAST_ADDR} \
  -e TZ=Asia/Shanghai \
  -e GIN_MODE=release \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  --restart unless-stopped \
  --health-cmd "wget --quiet --tries=1 --spider http://localhost:1323/ || exit 1" \
  --health-interval 30s \
  --health-timeout 10s \
  --health-retries 3 \
  $IMAGE_NAME

# 等待容器启动
echo "等待容器启动..."
sleep 3

# 检查容器状态
if docker ps | grep -q $CONTAINER_NAME; then
    echo "✅ 容器启动成功"
    echo "Web 管理界面: http://localhost:${PORT}/admin/"
    echo "API 地址: http://localhost:${PORT}/api/devices"
else
    echo "❌ 容器启动失败"
    echo "查看日志: docker logs $CONTAINER_NAME"
    exit 1
fi
```

#### Windows PowerShell 脚本

```powershell
# 配置变量
$IMAGE_NAME = "home-wol:latest"
$CONTAINER_NAME = "home-wol"
$CONFIG_FILE = "$PWD\config.json"
$BROADCAST_ADDR = "192.168.100.255"
$PORT = "1323"

# 检查配置文件
if (-not (Test-Path $CONFIG_FILE)) {
    Write-Host "错误: 配置文件不存在: $CONFIG_FILE" -ForegroundColor Red
    exit 1
}

# 停止并删除旧容器
Write-Host "停止旧容器..." -ForegroundColor Yellow
docker stop $CONTAINER_NAME 2>$null
docker rm $CONTAINER_NAME 2>$null

# 运行新容器
Write-Host "启动新容器..." -ForegroundColor Yellow
docker run -d `
  --name $CONTAINER_NAME `
  -p "${PORT}:1323" `
  -v "${CONFIG_FILE}:/home/config.json:ro" `
  -e BROADCAST_ADDRESS=$BROADCAST_ADDR `
  -e TZ=Asia/Shanghai `
  -e GIN_MODE=release `
  --log-driver json-file `
  --log-opt max-size=10m `
  --log-opt max-file=3 `
  --restart unless-stopped `
  $IMAGE_NAME

# 等待容器启动
Write-Host "等待容器启动..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# 检查容器状态
$containerStatus = docker ps --filter "name=$CONTAINER_NAME" --format "{{.Status}}"
if ($containerStatus) {
    Write-Host "✅ 容器启动成功" -ForegroundColor Green
    Write-Host "Web 管理界面: http://localhost:${PORT}/admin/"
    Write-Host "API 地址: http://localhost:${PORT}/api/devices"
} else {
    Write-Host "❌ 容器启动失败" -ForegroundColor Red
    Write-Host "查看日志: docker logs $CONTAINER_NAME"
    exit 1
}
```

### Docker Run vs Docker Compose 对比

| 特性 | Docker Run | Docker Compose |
|------|------------|----------------|
| 复杂度 | 简单 | 中等 |
| 配置文件 | 单条命令 | YAML 文件 |
| 可维护性 | 较低 | 高 |
| 多容器管理 | 不方便 | 方便 |
| 环境管理 | 手动 | 自动化 |
| 适合场景 | 测试、单容器 | 生产、多容器 |
| 配置版本控制 | 脚本 | YAML（推荐） |

### 推荐选择

- **使用 Docker Run**：
  - 快速测试
  - 单机部署
  - 简单场景
  - 不需要频繁修改配置

- **使用 Docker Compose**：
  - 生产环境
  - 需要配置多个服务
  - 团队协作
  - 需要版本控制

### 迁移指南

#### 从 Docker Run 迁移到 Docker Compose

1. **创建 docker-compose.yml**：
```yaml
services:
  home-wol:
    image: home-wol:latest
    container_name: home-wol
    ports:
      - "1323:1323"
    volumes:
      - ./config.json:/home/config.json:ro
    environment:
      - BROADCAST_ADDRESS=192.168.100.255
      - TZ=Asia/Shanghai
    restart: unless-stopped
```

2. **迁移命令**：
```bash
# 导出当前容器配置
docker inspect home-wol > container.json

# 使用 Compose 启动
docker-compose up -d

# 删除旧容器
docker rm -f home-wol
```

---

## 参考文档

- [Docker Run 官方文档](https://docs.docker.com/engine/reference/commandline/run/)
- [Docker Compose 文档](https://docs.docker.com/compose/)
- [项目主页](../../README.md)

# 删除配置文件（可选）
# rm -rf config/
```

## 参考文档

- [Docker Compose 文档](https://docs.docker.com/compose/)
- [Dockerfile 最佳实践](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [项目主文档](./QUICK_START.md)
