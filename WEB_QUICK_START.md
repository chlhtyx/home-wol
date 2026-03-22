# Web 管理界面快速指南

## 功能特性

### 🎨 现代化的 Web 界面
- 响应式设计，支持移动端和桌面端
- 简洁美观的 UI 设计
- 流畅的动画和交互效果
- 实时反馈和通知提示

### 📱 设备管理
- ✅ **设备列表** - 查看所有已配置的设备
- ✅ **添加设备** - 快速添加新设备，自动生成密钥
- ✅ **编辑设备** - 修改设备名称
- ✅ **删除设备** - 移除不需要的设备
- ✅ **MAC 地址验证** - 自动验证和格式化 MAC 地址

### 🔐 密钥管理
- ✅ **自动生成密钥** - 添加设备时自动生成安全的 Base32 密钥
- ✅ **显示二维码** - 扫码即可添加到 Google Authenticator
- ✅ **查看验证码** - 实时获取当前 6 位验证码
- ✅ **密钥保护** - 密钥不可编辑，确保安全性

### ⚡ 测试功能
- ✅ **一键唤醒** - 点击按钮即可发送 WOL 唤醒指令
- ✅ **自动验证** - 自动生成验证码并完成验证
- ✅ **状态反馈** - 实时显示唤醒结果

## 快速开始

### 1. 启动服务

```bash
# 方式 1：直接运行
go run main.go

# 方式 2：编译后运行
go build -o wol.exe
./wol

# 方式 3：使用 Docker
docker-compose up -d
```

### 2. 访问管理界面

打开浏览器访问：
```
http://localhost:1323/admin/
```

### 3. 添加设备

1. 点击右上角 **"添加设备"** 按钮
2. 填写设备信息：
   - **设备名称**：如"我的电脑"、"NAS"等
   - **MAC 地址**：设备的 MAC 地址
     - 支持格式：`AA:BB:CC:DD:EE:FF`、`AA-BB-CC-DD-EE-FF`、`AABBCCDDEEFF`
   - **密钥**：系统自动生成（也可手动输入）
3. 点击 **"保存"**
4. 系统会显示成功提示

### 4. 配置 Google Authenticator

**方式一：扫描二维码**
1. 在设备列表中找到新添加的设备
2. 点击 **"二维码"** 按钮
3. 使用 Google Authenticator 应用扫描二维码

**方式二：手动输入**
1. 点击设备的 **"显示密钥"** 链接
2. 复制显示的密钥
3. 在 Google Authenticator 中手动输入

### 5. 测试唤醒

1. 在设备列表中点击 **"测试唤醒"** 按钮
2. 系统会自动：
   - 生成当前验证码
   - 发送唤醒指令
   - 显示结果提示

## API 接口

管理界面使用以下 API 接口：

### 获取设备列表
```http
GET /api/devices
```

响应：
```json
{
  "devices": [
    {
      "mac": "AABBCCDDEEFF",
      "name": "我的电脑",
      "secret": "JBSWY3DPEHPK3PXP"
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
  "mac": "001122334455",
  "secret": "JBSWY3DPEHPK3PXP"
}
```

### 更新设备
```http
PUT /api/devices/AABBCCDDEEFF
Content-Type: application/json

{
  "name": "更新的设备名"
}
```

### 删除设备
```http
DELETE /api/devices/AABBCCDDEEFF
```

### 获取验证码
```http
GET /api/devices/AABBCCDDEEFF/code
```

响应：
```json
{
  "code": "123456"
}
```

### 获取二维码
```http
GET /api/devices/AABBCCDDEEFF/qrcode
```

响应：
```json
{
  "qrcode": "https://api.qrserver.com/...",
  "code": "123456",
  "secret": "JBSWY3DPEHPK3PXP"
}
```

## 界面预览

### 主界面
- 顶部标题栏
- 设备列表卡片
- 添加设备按钮
- 响应式布局

### 设备卡片
每个设备显示：
- 设备名称
- MAC 地址
- 操作按钮：
  - 🔄 测试唤醒
  - 📱 二维码
  - ✏️ 编辑
  - 🗑️ 删除

### 模态框
- 添加/编辑设备表单
- 二维码显示
- 确认删除对话框

## 技术细节

### 前端技术
- 纯 HTML5 + CSS3 + JavaScript
- 无框架依赖
- Fetch API 进行 HTTP 请求
- CSS Variables 主题系统
- 响应式设计（Mobile First）

### 后端技术
- Go + Echo 框架
- RESTful API 设计
- 线程安全的配置文件操作
- 自动配置备份
- MAC 地址标准化

### 安全特性
- 配置文件操作使用互斥锁保护
- 修改前自动备份配置
- MAC 地址格式验证
- HTML 转义防止 XSS

## 文件结构

```
D:\code\home-wol\
├── static/                    # 静态文件目录
│   ├── index.html            # 管理页面
│   ├── css/
│   │   └── style.css         # 样式文件
│   └── js/
│       └── app.js            # JavaScript 逻辑
├── service/
│   └── api.go                # API 接口
├── main.go                   # 主程序（已更新路由）
├── WEB_MANUAL.md             # 详细文档
└── config.json               # 配置文件
```

## 常见问题

### 1. 无法访问管理界面

**检查项：**
- 服务是否正常启动
- 端口 1323 是否被占用
- 浏览器地址是否正确：`http://localhost:1323/admin/`

### 2. 设备列表为空

**原因：**
- 首次使用，配置文件为空
- 配置文件路径不正确

**解决：**
- 点击"添加设备"添加设备
- 检查配置文件路径

### 3. 添加设备失败

**常见原因：**
- MAC 地址格式不正确
- 设备名称为空
- MAC 地址重复

**解决：**
- 检查 MAC 地址格式
- 填写设备名称
- 确认 MAC 地址唯一

### 4. 测试唤醒失败

**检查项：**
- 广播地址配置是否正确
- 设备是否支持 WOL
- 设备是否在同一网络

### 5. 二维码无法扫描

**解决：**
- 检查 Google Authenticator 应用是否安装
- 尝试手动输入密钥
- 点击"显示密钥"查看密钥内容

## 高级功能

### 配置文件备份

每次修改配置时，系统会自动创建备份文件：
- `config.json.backup`

修改失败时可以恢复：
```bash
cp config.json.backup config.json
```

### 批量操作

可以通过 API 进行批量操作：
```bash
# 批量添加设备
curl -X POST http://localhost:1323/api/devices \
  -H "Content-Type: application/json" \
  -d '{"name":"设备1","mac":"001122334455","secret":"SECRET1"}'

# 批量删除设备
curl -X DELETE http://localhost:1323/api/devices/001122334455
```

### 自定义样式

修改 `static/css/style.css` 可以自定义界面样式：
```css
/* 修改主题色 */
:root {
  --primary-color: #007bff;  /* 改为你的颜色 */
}
```

## 部署建议

### 生产环境

1. **启用 HTTPS**
   - 使用 Nginx 反向代理
   - 配置 SSL 证书

2. **访问控制**
   - 添加身份验证
   - 限制访问 IP

3. **安全加固**
   - 定期更新依赖
   - 使用强密钥

### Docker 部署

```yaml
# docker-compose.yml
services:
  home-wol:
    ports:
      - "1323:1323"
    volumes:
      - ./config:/home/config
    environment:
      - CONFIG_PATH=/home/config/config.json
```

访问：`http://localhost:1323/admin/`

## 更多信息

- 📖 [详细文档](./WEB_MANUAL.md)
- 🚀 [快速开始](./QUICK_START.md)
- 🐳 [Docker 部署](./README_DOCKER.md)
- 📝 [配置说明](./MULTI_DEVICE_CONFIG.md)

## 支持

如有问题，请查看：
1. 浏览器控制台（F12）查看错误信息
2. 服务日志：`docker-compose logs -f`
3. 配置文件格式：验证 JSON 是否正确
