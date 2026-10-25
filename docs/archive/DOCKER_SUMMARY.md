# Docker 部署完成总结

> 历史记录：本文描述过去的实现或验证结果，不代表当前运行状态。当前操作方式见[项目 README](../../README.md)。

## 已创建的文件

### 核心文件
- ✅ `docker-compose.yml` - Docker Compose 配置文件
- ✅ `.dockerignore` - Docker 构建忽略文件
- ✅ `.env.example` - 环境变量示例文件
- ✅ `Makefile` - Make 命令文件

### 配置文件
- ✅ `config/config.json.template` - 配置文件模板
- ✅ `config/config.json` - 实际配置文件（自动创建）

### 启动脚本
- ✅ `start.sh` - Linux/Mac 启动脚本
- ✅ `start.bat` - Windows 启动脚本
- ✅ `check.sh` - 部署检查脚本

### 文档
- ✅ `README_DOCKER.md` - Docker 快速部署指南
- ✅ `DOCKER_DEPLOY.md` - Docker 详细部署文档

## 快速开始

### Windows 用户

1. 双击运行 `start.bat`
2. 编辑 `config/config.json` 添加设备
3. 编辑 `.env` 设置广播地址
4. 重新运行 `start.bat`

### Linux/Mac 用户

```bash
# 方式1：使用启动脚本
chmod +x start.sh
./start.sh

# 方式2：使用 Makefile
make init      # 初始化配置
make up        # 启动服务
make logs      # 查看日志
```

## 配置文件挂载

配置文件已通过 volume 挂载到容器：

```yaml
volumes:
  - ./config:/home/config
```

这样可以在宿主机直接编辑配置文件，重启容器后生效。

## 目录结构

```
D:\code\home-wol\
├── docker-compose.yml          # Docker Compose 配置
├── .dockerignore              # Docker 忽略文件
├── .env.example               # 环境变量示例
├── .env                       # 环境变量（需创建）
├── start.sh                   # Linux/Mac 启动脚本
├── start.bat                  # Windows 启动脚本
├── check.sh                   # 部署检查脚本
├── Makefile                   # Make 命令
├── config/                    # 配置文件目录
│   ├── config.json.template   # 配置模板
│   └── config.json            # 实际配置（挂载到容器）
├── README_DOCKER.md           # Docker 快速部署
└── DOCKER_DEPLOY.md           # Docker 详细文档
```

## 环境变量

在 `.env` 文件中配置：

```env
# 广播地址（重要！）
BROADCAST_ADDRESS=192.168.1.255

# 服务端口
PORT=1323

# 时区
TZ=Asia/Shanghai
```

## 网络模式

### Bridge 模式（默认）
- 端口映射：`1323:1323`
- 适用场景：一般使用
- 限制：可能无法发送广播包

### Host 模式（推荐 WOL）
- 使用宿主机网络
- 适用场景：WOL 唤醒
- 配置：取消 `docker-compose.yml` 中的 `network_mode: "host"` 注释

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
make rebuild       # 重新构建
make update        # 更新并重启
```

## 配置文件管理

1. 编辑 `config/config.json`
2. 重启服务：`docker-compose restart`
3. 配置自动生效

配置文件内容示例：

```json
{
  "devices": [
    {
      "mac": "E0:E1:A9:10:87:06",
      "name": "电脑1",
      "secret": "JBSWY3DPEHPK3PXP"
    },
    {
      "mac": "112233445566",
      "name": "电脑2",
      "secret": "KRSXG5DSNFXAWZL"
    }
  ]
}
```

## 验证部署

运行检查脚本：

```bash
chmod +x check.sh
./check.sh
```

或手动测试：

```bash
# 测试服务
curl "http://localhost:1323"

# 生成验证码
curl "http://localhost:1323/generate_code?mac=YOUR_MAC"

# 唤醒设备
curl "http://localhost:1323/wol?mac=YOUR_MAC&auth_code=YOUR_CODE"
```

## 故障排查

### 1. 容器启动失败
```bash
docker-compose logs home-wol
```

### 2. 无法唤醒设备
- 检查广播地址配置
- 尝试使用 host 网络模式
- 检查设备是否支持 WOL

### 3. 配置不生效
```bash
docker-compose restart
```

## 下一步

1. 根据你的网络环境修改 `.env` 中的 `BROADCAST_ADDRESS`
2. 编辑 `config/config.json` 添加你的设备信息
3. 运行 `start.bat` (Windows) 或 `./start.sh` (Linux/Mac)
4. 使用 Google Authenticator 扫描二维码或输入密钥
5. 测试唤醒功能

## 文档链接

- [Docker 快速部署](../README_DOCKER.md)
- [Docker 详细部署](../DOCKER_DEPLOY.md)
- [快速开始指南](../QUICK_START.md)
- [多设备配置说明](../MULTI_DEVICE_CONFIG.md)
