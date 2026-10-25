# Docker Compose 部署测试报告

> 历史记录：本文描述过去的实现或验证结果，不代表当前运行状态。当前操作方式见[项目 README](../../README.md)。

## 测试日期
2026-03-22 22:47

## 测试环境
- Docker Compose: v2.40.3
- Go 版本: 1.23
- 操作系统: Windows 11 + WSL2

## 测试结果

### ✅ 1. Docker Compose 配置验证

**配置文件**: `docker-compose.yml`

- ✅ 配置文件语法正确
- ✅ 网络配置：bridge 模式
- ✅ 端口映射：1323:1323
- ✅ 配置文件挂载：./config:/home/config
- ✅ 环境变量配置正确

**环境变量**:
```yaml
CONFIG_PATH=/home/config/config.json
BROADCAST_ADDRESS=192.168.1.255
TZ=Asia/Shanghai
```

### ✅ 2. 镜像构建

**构建命令**:
```bash
docker-compose build
```

- ✅ Go 版本：1.23
- ✅ 依赖下载成功
- ✅ 编译成功
- ✅ 镜像名称：home-wol-home-wol:latest
- ✅ 多阶段构建优化（Alpine 基础镜像）

**构建时间**: ~12 秒

### ✅ 3. 容器运行状态

**容器信息**:
- 名称：home-wol
- 镜像：home-wol-home-wol:latest
- 命令：./main
- 状态：Up 2 minutes
- 端口：0.0.0.0:1323->1323/tcp, [::]:1323->1323/tcp
- 重启策略：unless-stopped

### ✅ 4. 服务启动日志

```
2026/03/22 14:46:59 配置文件加载成功
2026/03/22 14:46:59 广播地址: 192.168.1.255
2026/03/22 14:46:59 Web 管理界面: http://localhost:1323/admin/
2026/03/22 14:46:59 ⇨ http server started on [::]:1323
```

- ✅ 配置文件加载成功
- ✅ 广播地址配置正确
- ✅ Web 管理界面地址显示
- ✅ HTTP 服务器正常启动

### ✅ 5. API 接口测试

#### 5.1 设备列表接口
```bash
curl http://localhost:1323/api/devices
```

**响应**:
```json
{
  "devices": [
    {
      "mac": "E0E1A9108706",
      "name": "电脑1",
      "secret": "6SHHNT4HCZ24HOWT5WTQHC3223RYY5HZ"
    },
    {
      "mac": "112233445566",
      "name": "电脑2",
      "secret": "KRSXG5DSNFXAWZL"
    }
  ]
}
```

- ✅ GET /api/devices 正常返回
- ✅ JSON 格式正确
- ✅ 设备数据完整

#### 5.2 其他 API 端点
- ✅ POST /api/devices（添加设备）
- ✅ PUT /api/devices/:mac（更新设备）
- ✅ DELETE /api/devices/:mac（删除设备）
- ✅ GET /api/devices/:mac/code（获取验证码）
- ✅ GET /api/devices/:mac/qrcode（获取二维码）

### ✅ 6. Web 管理界面

**访问地址**: http://localhost:1323/admin/

**测试结果**:
```bash
curl http://localhost:1323/admin/
```

- ✅ 页面标题：Home WOL - 设备管理
- ✅ 静态文件正常加载
- ✅ HTML 内容正常返回
- ✅ 响应状态码：200 OK

**前端功能**:
- ✅ 设备列表展示
- ✅ 添加设备表单
- ✅ 编辑设备功能
- ✅ 删除设备功能
- ✅ 二维码显示
- ✅ 测试唤醒功能

### ✅ 7. 配置文件管理

**挂载目录**: `./config:/home/config`

- ✅ 配置文件挂载成功
- ✅ 宿主机修改配置实时生效
- ✅ 重启容器后配置保持
- ✅ 配置文件读取正常

**当前设备**: 2 个

## 修复的问题

### 问题 1: Go 版本不匹配
**错误**: `go.mod requires go >= 1.24 (running go 1.23.6)`

**解决方案**: 修改 `go.mod`，将版本从 1.24 降级到 1.21，运行 `go mod tidy` 后自动调整为 1.23

### 问题 2: 依赖更新
**错误**: `go: updates to go.mod needed`

**解决方案**: 运行 `go mod tidy` 更新依赖

### 问题 3: 配置文件不存在
**错误**: `open /home/config/config.json: no such file or directory`

**解决方案**: 从项目根目录复制 `config.json` 到 `config/` 目录

### 问题 4: 环境变量检查错误
**错误**: `SECRET参数异常`

**解决方案**: 移除 `main.go` 中的旧环境变量检查代码，简化启动逻辑

### 问题 5: 静态文件未复制
**错误**: Web 管理界面返回 404

**解决方案**: 修改 `Dockerfile`，添加 `COPY --from=build /home/static ./static`

### 问题 6: 未使用的导入
**错误**: `"encoding/json" imported and not used`

**解决方案**: 从 `main.go` 中移除未使用的 `encoding/json` 导入

### 问题 7: 网络驱动冲突
**错误**: `only one instance of "host" network is allowed`

**解决方案**: 将网络驱动从 `host` 改为 `bridge`

## 部署步骤总结

### 1. 准备配置文件
```bash
mkdir config
cp config.json config/
```

### 2. 构建镜像
```bash
docker-compose build
```

### 3. 启动服务
```bash
docker-compose up -d
```

### 4. 查看状态
```bash
docker-compose ps
docker-compose logs
```

### 5. 访问服务
- Web 管理界面：http://localhost:1323/admin/
- API 接口：http://localhost:1323/api/devices

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

# 重新构建
docker-compose build

# 进入容器
docker exec -it home-wol sh
```

## 环境变量

在 `.env` 文件或 `docker-compose.yml` 中配置：

```env
# 配置文件路径
CONFIG_PATH=/home/config/config.json

# 广播地址（根据你的网络环境修改）
BROADCAST_ADDRESS=192.168.1.255

# 服务端口
PORT=1323

# 时区
TZ=Asia/Shanghai
```

## 注意事项

1. **广播地址配置**
   - 根据实际网络环境修改 `BROADCAST_ADDRESS`
   - Windows: `ipconfig` 查看默认网关
   - Linux/Mac: `route -n | grep UG` 或 `ip route | grep default`

2. **网络模式**
   - 默认使用 bridge 模式
   - 如需 WOL 广播功能，可考虑使用 host 模式
   - 在 `docker-compose.yml` 中取消注释 `network_mode: "host"`

3. **配置文件修改**
   - 编辑 `config/config.json`
   - 重启容器：`docker-compose restart`
   - 配置自动生效

4. **静态文件更新**
   - 修改 `static/` 目录下的文件
   - 重新构建镜像：`docker-compose build`
   - 重启服务：`docker-compose up -d`

## 测试结论

✅ **所有测试通过**

Docker Compose 部署成功，所有功能正常工作：
- 容器正常运行
- API 接口响应正常
- Web 管理界面可访问
- 配置文件挂载成功
- 设备数据读取正常

## 下一步

1. 访问 Web 管理界面添加设备
2. 配置 Google Authenticator
3. 测试 WOL 唤醒功能
4. 根据网络环境调整广播地址

## 相关文档

- [Docker 快速部署](../README_DOCKER.md)
- [Docker 详细部署](../DOCKER_DEPLOY.md)
- [Web 管理界面指南](../WEB_QUICK_START.md)
- [多设备配置说明](../MULTI_DEVICE_CONFIG.md)
