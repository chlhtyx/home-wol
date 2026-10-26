# Home WOL Web 管理界面

> 命令均从项目根目录执行；当前目录布局和启动方式见[项目 README](../README.md)。旧目录示意保留作参考。

## 功能概述

Home WOL 项目现在包含了一个完整的 Web 管理界面，用于方便地配置和管理 Wake-on-LAN 设备。

## 新增功能

### 1. 设备管理
- **设备列表展示**: 查看所有已配置的设备
- **添加设备**: 添加新的 Wake-on-LAN 设备
- **编辑设备**: 修改设备名称
- **删除设备**: 移除不需要的设备

### 2. 设备操作
- **测试唤醒**: 直接从 Web 界面测试设备的 Wake-on-LAN 功能
- **生成二维码**: 为每个设备生成二维码，方便移动设备扫描
- **查看密钥**: 显示设备的密钥和当前动态验证码

## 访问方式

启动前设置环境变量 `ADMIN_PASSWORD`，账号由 `ADMIN_USERNAME` 指定，默认为 `admin`；Docker Compose 从 `.env` 读取，本地运行需设置进程环境变量。没有默认密码，缺少密码时服务拒绝启动。

访问 `/admin/` 会先跳转 `/login`。登录会话有效期 12 小时，页面提供“退出登录”；退出或服务重启后原会话失效，过期访问接口会自动返回登录页。

页面顶部“网络设置”可查看和修改广播地址。填写目标子网的 IPv4 广播地址（例如 `192.168.1.255`），点击“保存设置”后立即用于后续唤醒请求，无需重启。设置持久化在根目录 `config.json` 的 `broadcast_address` 字段。

`GET /api/settings` 读取当前生效的广播地址；`PUT /api/settings` 使用 JSON `{"broadcast_address":"192.168.1.255"}` 保存。空地址、IPv6、无效 IPv4、回环、多播和未指定地址会返回 400，保存失败会显示错误提示。旧环境变量只在配置文件未设置广播地址时作为回退。

启动服务后，通过浏览器访问：

```
http://localhost:1323/admin/
```

## API 接口

以下管理接口以及旧的 `/generate_code` 均需携带登录 Cookie，未登录返回 401。先向 `POST /api/login` 发送 JSON `{"username":"admin","password":"你的密码"}`，并保存响应 Cookie；登录、退出和所有写接口必须携带 `X-Requested-With: XMLHttpRequest` 请求头（防止跨站请求）。`POST /api/logout` 使当前会话失效。

`GET /wol?mac=...&auth_code=...` 继续使用设备动态验证码，无需管理端会话；`GET /healthz` 为公开健康检查。

### 获取设备列表
```http
GET /api/devices
```

响应示例：
```json
{
  "devices": [
    {
      "mac": "001122334455",
      "name": "我的电脑",
      "secret": "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234"
    }
  ]
}
```

### 添加设备
```http
POST /api/devices
Content-Type: application/json

{
  "name": "新设备",
  "mac": "00:11:22:33:44:55",
  "secret": "SECRETKEY1234567890123456789012"
}
```

### 更新设备
```http
PUT /api/devices/{mac}
Content-Type: application/json

{
  "name": "更新后的设备名"
}
```

### 删除设备
```http
DELETE /api/devices/{mac}
```

### 获取设备验证码
```http
GET /api/devices/{mac}/code
```

响应示例：
```json
{
  "mac": "001122334455",
  "code": "123456",
  "secret": "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234"
}
```

### 获取设备二维码
```http
GET /api/devices/{mac}/qrcode
```

响应示例：
```json
{
  "qrcode": "https://api.qrserver.com/v1/create-qr-code/?...",
  "code": "123456",
  "secret": "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234"
}
```

## 文件结构

```
home-wol/
├── static/
│   ├── index.html          # Web 管理界面主页面
│   ├── css/
│   │   └── style.css       # 样式文件
│   └── js/
│       └── app.js          # 前端逻辑
├── service/
│   ├── service.go          # 原有服务
│   └── api.go              # 新增 API 接口
├── main.go                 # 主程序（已更新）
└── config.json             # 设备配置文件
```

## 配置文件格式

`config.json` 示例：

```json
{
  "devices": [
    {
      "mac": "001122334455",
      "name": "我的电脑",
      "secret": "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234"
    },
    {
      "mac": "AABBCCDDEEFF",
      "name": "NAS 服务器",
      "secret": "2345678901234567890123456789012"
    }
  ]
}
```

## 使用说明

### 添加新设备

1. 点击"添加设备"按钮
2. 填写设备信息：
   - 设备名称：例如"主机"、"NAS"、"游戏机"
   - MAC 地址：格式为 `XX:XX:XX:XX:XX:XX` 或 `XX-XX-XX-XX-XX-XX`
   - 密钥：点击"生成新密钥"自动生成，或手动输入
3. 点击"保存"

### 测试唤醒功能

1. 在设备列表中找到要测试的设备
2. 点击"测试唤醒"按钮
3. 系统会自动获取当前验证码并发送唤醒指令

### 生成二维码

1. 点击设备的"二维码"按钮
2. 系统会显示设备的二维码和当前验证码
3. 使用移动设备的验证器应用（如 Google Authenticator、FreeOTP 等）扫描二维码
4. 扫描后可以在移动设备上获取动态验证码

### 编辑设备

1. 点击设备的"编辑"按钮
2. 修改设备名称
3. 点击"保存"
4. 注意：MAC 地址不可编辑（如需修改，请删除后重新添加）

### 删除设备

1. 点击设备的"删除"按钮
2. 确认删除操作

## 技术特性

### 前端
- 纯 HTML/CSS/JavaScript 实现
- 响应式设计，支持移动端和桌面端
- 现代化的 UI 设计
- 无需额外框架

### 后端
- RESTful API 设计
- 支持并发安全的配置文件操作
- 自动备份配置文件（生成 `.backup.json`）
- 完整的错误处理

### 安全性
- MAC 地址格式验证
- 配置文件修改时的自动备份
- 密钥管理（使用 Google Authenticator 兼容的 TOTP）

## 兼容性

### 浏览器支持
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

### 设备支持
- 所有支持 Wake-on-LAN 的设备
- 移动设备验证器应用（Google Authenticator、Authy、FreeOTP 等）

## 注意事项

1. **配置文件备份**: 每次修改设备配置时，系统会自动创建备份文件（`config.backup.json`）
2. **MAC 地址**: MAC 地址会被标准化为大写且无分隔符的格式
3. **密钥安全**: 请妥善保管设备密钥，不要泄露给他人
4. **网络要求**: 确保设备在同一网络或可访问的广播域内

## 故障排除

### 无法访问 Web 界面
- 检查服务是否正常启动
- 检查端口 1323 是否被占用
- 检查防火墙设置

### 唤醒失败
- 确认设备已启用 Wake-on-LAN 功能
- 确认设备与服务器在同一网络
- 检查设备的 MAC 地址是否正确
- 检查网络配置和广播地址设置

### 配置文件错误
- 检查配置文件格式是否正确
- 尝试使用备份文件恢复：将 `config.backup.json` 重命名为 `config.json`

## 开发与贡献

如需修改或扩展功能：

1. 前端代码位于 `static/` 目录
2. 后端 API 代码位于 `service/api.go`
3. 路由配置在 `main.go` 中

## 许可证

本项目遵循原有的开源许可证。
