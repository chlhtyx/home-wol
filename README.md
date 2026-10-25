# Home WOL

基于 Go 的局域网 Wake-on-LAN 服务，提供 Web 设备管理、二维码和动态验证码。

## Docker 启动

以下命令均在项目根目录执行。首次使用时复制配置模板；已有配置请直接保留。

```powershell
# Windows PowerShell（仅在文件不存在时复制）
if (!(Test-Path config.json)) { Copy-Item config.example.json config.json }
if (!(Test-Path .env)) { Copy-Item .env.example .env }
```

```bash
# Linux / macOS（仅在文件不存在时复制）
[ -f config.json ] || cp config.example.json config.json
[ -f .env ] || cp .env.example .env
```

编辑 `config.json` 中的设备信息，然后启动：

```bash
docker compose up -d --build
docker compose ps
docker compose logs -f
```

Compose 从本地源码构建 `home-wol:local`，无需登录阿里云仓库或执行 `docker pull`。请在包含 `Dockerfile` 和源码的项目根目录启动；构建仍需能够获取 Dockerfile 中的基础镜像和 Go 依赖。

默认管理地址：<http://localhost:1323/admin/>。宿主机端口由 `.env` 的 `PORT` 设置，默认为 `1323`。

在管理页面顶部的“网络设置”中填写目标设备所在子网的 IPv4 广播地址并保存。配置会写入 `config.json` 的 `broadcast_address` 字段，下一次唤醒立即使用新地址，重启后保留，无需修改 `.env`。设备的新增、编辑和删除会保留网络设置。

旧版配置没有 `broadcast_address` 时，兼容读取 `BROADCAST_ADDRESS` 环境变量；未设置时使用 `192.168.100.255`。页面保存后的配置优先于环境变量。

默认使用 bridge 网络和端口映射，兼容 Windows Docker Desktop。Linux 宿主机需要直接发送局域网广播时，可以按照 `docker-compose.yml` 注释切换为 host 网络；host 模式使用应用固定监听的 `1323` 端口。Windows 上的 WOL 广播需要根据实际网络环境验证。

也可在 Windows 双击 `scripts/start.bat`，或在 Linux / macOS 执行 `bash scripts/start.sh`。脚本会自动切换到项目根目录并初始化缺失的配置。部署检查命令为 `bash scripts/check.sh`。

## 本地开发

需要 Go 1.23 或更高版本。在项目根目录运行，以便程序读取配置和 `static/`：

```bash
go run .
```

生成设备密钥：

```bash
go run ./cmd/generate_secret
```

Windows 编译和启动：

```powershell
go build -o bin/wol.exe .
go build -o bin/generate_secret.exe ./cmd/generate_secret
.\bin\wol.exe
```

## 目录说明

```text
home-wol/
├── main.go                  # 服务入口
├── common/                  # 配置、认证和 WOL 工具
├── service/                 # HTTP 接口
├── cmd/generate_secret/     # 密钥生成工具源码
├── static/                  # Web 管理界面
├── scripts/                 # 启动和部署检查脚本
├── docs/                    # 使用和部署文档
│   └── archive/             # 历史实现总结与测试记录
├── bin/                     # 本地可执行文件（忽略）
├── artifacts/               # Docker 镜像归档（忽略）
├── logs/                    # 本地历史日志（忽略）
├── config.example.json      # 设备配置模板
├── config.json              # 当前设备配置
├── config.backup.json       # 本地设备配置备份
├── .env.example             # 环境变量模板
├── .env                     # 当前环境变量
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

Docker 挂载根目录的 `config.json`，Web 管理界面可直接保存设备变更。`CONFIG_PATH` 指定容器内挂载路径，本地运行时则指定实际配置文件路径。

`bin/`、`artifacts/` 和 `logs/` 保存本地文件，不参与 Docker 构建；首次克隆时这些目录可能不存在。

## 文档

- [文档导航](docs/README.md)
- [Docker 快速部署](docs/README_DOCKER.md)
- [Web 管理说明](docs/WEB_MANUAL.md)
- [多设备配置](docs/MULTI_DEVICE_CONFIG.md)
- [更新日志](CHANGELOG.md)

历史资料中的目录示意和部署示例可能对应早期版本，当前启动方式与目录结构以本文及 Compose 文件为准。
