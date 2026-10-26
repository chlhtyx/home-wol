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

编辑 `config.json` 中的设备信息，并在 `.env` 中设置 `ADMIN_PASSWORD`（必填，无默认密码）；`ADMIN_USERNAME` 默认是 `admin`。然后启动：

```bash
docker compose up -d --build
docker compose ps
docker compose logs -f
```

以上方式用于从源码构建，请在包含 `Dockerfile` 和源码的项目根目录启动；镜像名称以 `docker-compose.yml` 的 `image` 为准。构建需要能够获取 Dockerfile 中的基础镜像和 Go 依赖。

软路由使用已发布镜像时，请按[软路由部署说明](docs/ROUTER_DEPLOY.md)在软路由上登录阿里云仓库，再使用独立的 `docker-compose.router.yml` 拉取并启动。该配置适用于 Linux amd64，使用 host 网络，软路由端无需源码。

默认管理地址：<http://localhost:1323/admin/>；远程访问使用 `http://软路由的局域网IP:1323/admin/`。host 模式直接使用宿主机 `1323` 端口，`.env` 的 `PORT` 不生效，请确保端口未被其他服务占用。

管理端必须先登录；设备管理、网络设置、二维码、验证码接口（包括旧的 `/generate_code`）都要求登录。会话有效期为 12 小时，点击“退出登录”立即失效，服务重启后需重新登录。更换账号或密码后需重新创建容器，本地运行则重新启动进程。未配置密码时服务拒绝启动。`/wol` 继续使用设备的动态验证码验证，`/healthz` 为不含业务数据的公开健康检查接口。

本次登录认证是源码变更，使用源码部署时需重新构建；软路由使用远程镜像时，须先发布包含此变更的新镜像，并更新镜像标签后再拉取部署。

在管理页面顶部的“网络设置”中填写目标设备所在子网的 IPv4 广播地址并保存。配置会写入 `config.json` 的 `broadcast_address` 字段，下一次唤醒立即使用新地址，重启后保留，无需修改 `.env`。设备的新增、编辑和删除会保留网络设置。

旧版配置没有 `broadcast_address` 时，兼容读取 `BROADCAST_ADDRESS` 环境变量；未设置时使用 `192.168.100.255`。页面保存后的配置优先于环境变量。

默认使用 host 网络，适用于 Linux 软路由，避免容器 bridge 网络影响局域网 WOL 广播。Windows Docker Desktop 本地开发时可将 `network_mode: host` 删除并添加 `ports: ["1323:1323"]`，WOL 广播需要根据实际网络环境验证。

也可在 Windows 双击 `scripts/start.bat`，或在 Linux / macOS 执行 `bash scripts/start.sh`。脚本会自动切换到项目根目录并初始化缺失的配置。部署检查命令为 `bash scripts/check.sh`。

## 本地开发

需要 Go 1.27.3 或更新的受支持版本，`go.mod` 已设置最低版本；支持自动工具链的 Go 会按需下载。Docker 使用 `golang:1.27.3-alpine3.24` 编译和 `alpine:3.24.3` 运行，不再依赖未标版本的自定义构建镜像。在项目根目录运行，以便程序读取配置和 `static/`：

```bash
export ADMIN_USERNAME=admin
read -rsp '管理端密码: ' ADMIN_PASSWORD; echo
export ADMIN_PASSWORD
go run .
```

本地 `go run` 不会自动读取 `.env`，需像上面这样设置进程环境变量；`.env` 由 Docker Compose 加载。

生成设备密钥：

```bash
go run ./cmd/generate_secret
```

Windows 编译和启动：

```powershell
$env:ADMIN_USERNAME = 'admin'
$adminCredential = Get-Credential -UserName admin -Message '设置管理端登录密码'
$env:ADMIN_PASSWORD = $adminCredential.GetNetworkCredential().Password
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

## 安全验证

安全升级后可运行 `go test ./...`、`go vet ./...` 和 `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...` 验证。Docker 镜像应重新构建并扫描后再发布。

依赖模块中仍有 `golang.org/x/crypto/openpgp` 的未维护告警（GO-2026-5932），没有可升级的修复版本。本项目的依赖图只引入 crypto 的 ACME 相关包，未导入 OpenPGP；按模块版本扫描镜像时仍可能报告此项，应结合源码扫描判断可达性。

## 文档

- [文档导航](docs/README.md)
- [Docker 快速部署](docs/README_DOCKER.md)
- [Web 管理说明](docs/WEB_MANUAL.md)
- [多设备配置](docs/MULTI_DEVICE_CONFIG.md)
- [更新日志](CHANGELOG.md)

历史资料中的目录示意和部署示例可能对应早期版本，当前启动方式与目录结构以本文及 Compose 文件为准。
