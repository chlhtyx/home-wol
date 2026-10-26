# Docker 快速部署

> 命令均从项目根目录执行；当前目录布局和启动方式见[项目 README](../README.md)。旧目录示意保留作参考。

## 前置要求

- Docker
- Docker Compose

默认 Compose 使用 host 网络，面向 Linux 软路由。Windows Docker Desktop 本地开发需先删除 `network_mode: host` 并添加 `ports: ["1323:1323"]`；实际 WOL 广播需单独验证。

## 快速开始

### Windows 用户

双击运行 `scripts/start.bat` 脚本：

```batch
scripts/start.bat
```

### Linux/Mac 用户

使用启动脚本：

```bash
chmod +x scripts/start.sh
./scripts/start.sh
```

或使用 Makefile：

```bash
# 初始化配置
make init

# 编辑配置文件
vim config.json
vim .env

# 启动服务
make up

# 查看日志
make logs
```

## 配置说明

### 1. 编辑配置文件

**config.json** - 设备配置：

```json
{
  "devices": [
    {
      "mac": "E0:E1:A9:10:87:06",
      "name": "电脑1",
      "secret": "JBSWY3DPEHPK3PXP"
    }
  ]
}
```

**.env** - 环境变量：

```env
# 管理端账号，密码必须填写，没有默认密码。
ADMIN_USERNAME=admin
ADMIN_PASSWORD=请替换为你的强密码

# 广播地址请在启动后的 Web 管理页面“网络设置”中配置。

# 默认 host 模式固定监听宿主机 1323 端口，无需设置 PORT。

# 时区
TZ=Asia/Shanghai
```

### 2. 配置广播地址

启动后访问 <http://localhost:1323/admin/>，先用 `.env` 中配置的账号密码登录，再在“网络设置”中输入目标子网的 IPv4 广播地址并保存，下一次唤醒立即生效。配置写入 `config.json`，重启后保留。旧版 `BROADCAST_ADDRESS` 环境变量仅在文件未配置广播地址时生效。

广播地址由 IP 和子网掩码共同决定；下面 `192.168.1.255` 的示例假定子网为 `192.168.1.0/24`。

**Windows:**
```cmd
ipconfig
```
例如，设备 IP 为 `192.168.1.10`、子网掩码为 `255.255.255.0`，广播地址为 `192.168.1.255`。

**Linux/Mac:**
```bash
route -n | grep UG
```
或
```bash
ip route | grep default
```

### 3. 启动服务

```bash
docker-compose up -d
```

## 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 查看状态
docker-compose ps

# 进入容器
docker exec -it home-wol sh

# 重新构建
docker-compose build
```

## Makefile 命令

```bash
make help          # 显示帮助
make init          # 初始化配置
make up            # 启动服务
make down          # 停止服务
make restart       # 重启服务
make logs          # 查看日志
make ps            # 查看状态
make test          # 测试服务
make clean         # 清理
make shell         # 进入容器
make rebuild       # 重新构建并启动
```

## 测试服务

### 1. 健康检查与验证码

```bash
curl --fail "http://localhost:1323/healthz"
```

浏览器登录管理端后，在设备二维码中查看动态验证码。旧的 `/generate_code` 接口也要求登录 Cookie，未登录返回 401；程序化登录方法见 [Web 管理说明](WEB_MANUAL.md#api-接口)。

### 2. 唤醒设备

```bash
curl "http://localhost:1323/wol?mac=YOUR_MAC&auth_code=YOUR_CODE"
```

### 3. POST 方式

```bash
curl -X POST "http://localhost:1323/wolv2" \
  -d "mac=YOUR_MAC" \
  -d "auth_code=YOUR_CODE"
```

## 配置文件管理

配置文件位于 `config.json`，可以直接编辑：

1. 编辑配置文件
2. 重启服务：`docker-compose restart`
3. 或使用 Makefile：`make restart`

## 目录结构

```
D:\code\home-wol\
├── docker-compose.yml       # Docker Compose 配置
├── Dockerfile               # Docker 镜像构建
├── .dockerignore           # Docker 忽略文件
├── .env.example            # 环境变量示例
├── .env                    # 环境变量（需要创建）
├── scripts/start.sh                # Linux/Mac 启动脚本
├── scripts/start.bat               # Windows 启动脚本
├── Makefile                # Make 命令
├── config/                 # 配置文件目录（挂载）
│   ├── config.json.template  # 配置模板
│   └── config.json            # 实际配置（需要创建）
└── ...
```

## 故障排查

### 1. 容器启动失败

```bash
# 查看日志
docker-compose logs home-wol

# 检查配置文件
cat config.json

# 验证 JSON 格式
cat config.json | jq .
```

### 2. 无法唤醒设备

**检查广播地址：**

在 Web 管理页面的“网络设置”中确认广播地址为目标设备所在子网的广播地址。页面保存的设置优先于旧版 `BROADCAST_ADDRESS` 环境变量。

**确认使用 host 网络模式：**

当前 `docker-compose.yml` 和 `docker-compose.router.yml` 均已启用以下设置，且不配置 `ports`：
```yaml
network_mode: "host"
```

修改网络模式后必须重新创建容器，单纯 restart 不会应用网络变更。在原部署目录执行（使用软路由专用文件时添加 `-f docker-compose.router.yml`）：
```bash
docker compose up -d --no-build --force-recreate home-wol
docker inspect home-wol --format '{{.HostConfig.NetworkMode}}'
```

输出应为 `host`。管理地址为 `http://软路由的局域网IP:1323/admin/`。

### 3. 配置文件修改不生效

```bash
# 重启容器
docker-compose restart

# 或完全重启
docker-compose down
docker-compose up -d
```

### 4. 端口冲突

默认 host 模式直接使用宿主机 `1323` 端口，修改 `.env` 的 `PORT` 不生效。请先停止或调整占用该端口的其他服务，再重新创建容器。

## 网络模式

### Bridge 模式（本地开发可选）

删除 `network_mode: host` 后可使用以下端口映射；bridge 模式可能影响局域网广播，不推荐用于软路由 WOL 部署。

```yaml
ports:
  - "1323:1323"
```

### Host 模式（默认，推荐 Linux 软路由 WOL）

容器使用宿主机网络，可直接发送广播包。

```yaml
network_mode: "host"
```

## 安全建议

1. **不要提交配置文件到 Git**：已在 `.gitignore` 中排除
2. **使用强密钥**：使用工具生成随机密钥
3. **配置防火墙**：限制访问来源
4. **使用 HTTPS**：生产环境建议配置反向代理

## 更新

```bash
# 拉取最新代码
git pull

# 重新构建
docker-compose build

# 重启服务
docker-compose up -d
```

或使用 Makefile：

```bash
make update
```

## 卸载

```bash
# 停止并删除容器
docker-compose down

# 删除镜像
docker rmi home-wol_home-wol

# 删除配置（可选）
rm -rf config/ .env
```

## 更多信息

- [完整部署文档](./DOCKER_DEPLOY.md)
- [快速开始指南](./QUICK_START.md)
- [配置说明](./MULTI_DEVICE_CONFIG.md)
