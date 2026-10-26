# 更新日志

## 未发布 - 2026-10-26

- 升级 Echo 到 v4.16.1，修复静态资源路径、代理和中间件等安全问题；升级 `golang.org/x/net` 到 v0.60.0，并更新 crypto、text、sys 及相关间接依赖。
- 将最低 Go 版本提升到 1.27.3，Docker 构建镜像固定为 `golang:1.27.3-alpine3.24`，运行镜像升级为 `alpine:3.24.3`。
- 补充管理端静态资源的编码路径和目录穿越访问回归测试。

## 镜像 v1.2 - 2026-10-25

- 修复通过 Web 新增设备后，验证码接口仍读取启动时缓存，导致“设备不存在或未找到密钥”的问题。
- 设备新增、密钥修改和删除保存后立即刷新认证缓存，并为缓存读写添加同步保护。
- 测试唤醒失败时显示后端具体错误或 HTTP 状态，便于区分设备配置与代理问题。
- 增加空配置启动后新增设备、修改密钥、删除设备的认证回归测试。

## [2.0.0] - 2025-03-21

### 新增功能

#### 1. 多设备配置支持
- **配置文件系统**：新增 `config.json` 配置文件，支持为不同 MAC 地址配置独立的 secret
- **MAC 地址标准化**：自动处理多种 MAC 地址格式（`AABBCCDDEEFF`、`AA:BB:CC:DD:EE:FF`、`AA-BB-CC-DD-EE-FF`）
- **配置文件加载**：在服务启动时自动加载配置文件
- **环境变量支持**：通过 `CONFIG_PATH` 环境变量指定配置文件路径

#### 2. 新增接口
- **GET /generate_code**：生成动态校验码接口
  - 参数：`mac` - 设备 MAC 地址
  - 响应：JSON 格式，包含 mac、code、secret
  - 用途：测试和调试

#### 3. 工具和辅助功能
- **密钥生成工具**：`cmd/generate_secret/main.go`
  - 生成 Google Authenticator 密钥
  - 生成当前验证码
  - 生成二维码信息和配置示例
  - 交互式命令行工具

#### 4. 单元测试
- **配置测试** (`common/config_test.go`)
  - 测试配置文件加载
  - 测试 MAC 地址查找
  - 测试 MAC 地址标准化
  - 测试各种 MAC 地址格式

### 改进

#### 1. 验证逻辑优化
- **按设备验证**：根据 MAC 地址查找对应的 secret 进行验证
- **错误处理**：更详细的错误信息和错误类型
- **向后兼容**：保留全局 `SECRET` 环境变量支持

#### 2. 代码结构优化
- **模块化设计**：
  - `common/config.go`：配置文件管理
  - `common/errors.go`：错误定义
  - `service/service.go`：服务层增强

#### 3. 接口改进
- **POST /wolv2**：完善 POST 方式的 WOL 接口实现
- **错误提示**：更友好的错误信息

### 文档

#### 新增文档
- `MULTI_DEVICE_CONFIG.md`：多设备配置详细说明
- `QUICK_START.md`：快速开始指南
- `CHANGELOG.md`：更新日志
- `config.example.json`：配置文件示例

### 技术细节

#### 新增文件
```
D:\code\home-wol\
├── config.json                    # 配置文件
├── config.example.json            # 配置文件示例
├── common/
│   ├── config.go                  # 配置管理
│   ├── config_test.go             # 配置测试
│   └── errors.go                  # 错误定义
├── cmd/
│   └── generate_secret/
│       └── main.go                # 密钥生成工具
├── MULTI_DEVICE_CONFIG.md         # 配置文档
├── QUICK_START.md                 # 快速开始
└── CHANGELOG.md                   # 更新日志
```

#### 修改文件
```
D:\code\home-wol\
├── main.go                        # 添加配置文件加载
└── service/
    └── service.go                 # 改进验证逻辑，新增接口
```

### 依赖变更

无新增外部依赖，仅使用 Go 标准库和现有依赖。

### 测试

所有单元测试通过：
```
=== RUN   TestGetQrCode
--- PASS: TestGetQrCode (0.00s)
=== RUN   TestLoadConfig
--- PASS: TestLoadConfig (0.00s)
=== RUN   TestGetSecretByMAC
--- PASS: TestGetSecretByMAC (0.00s)
=== RUN   TestNormalizeMAC
--- PASS: TestNormalizeMAC (0.00s)
PASS
ok  	home-wol/common	0.045s
```

### 安全性改进

1. **独立密钥管理**：每个设备使用独立的 secret，降低风险
2. **配置文件保护**：配置文件包含敏感信息，需妥善保管
3. **错误信息安全**：不在错误信息中泄露敏感数据

### 向后兼容性

- ✅ 保留原有接口：`GET /wol` 和 `POST /wolv2`
- ✅ 保留全局 `SECRET` 环境变量支持（已废弃）
- ✅ 保持原有 WOL 功能不变

### 迁移指南

从 1.x 版本升级到 2.0：

1. **创建配置文件**
   ```bash
   # 使用示例配置
   cp config.example.json config.json

   # 或使用工具生成
   go run cmd/generate_secret/main.go
   ```

2. **配置设备**
   编辑 `config.json`，添加你的设备和密钥

3. **启动服务**
   ```bash
   # 使用默认配置
   ./wol

   # 或指定配置路径
   CONFIG_PATH=/path/to/config.json ./wol
   ```

4. **验证功能**
   ```bash
   # 测试生成验证码
   curl "http://localhost:1323/generate_code?mac=YOUR_MAC"

   # 测试唤醒设备
   curl "http://localhost:1323/wol?mac=YOUR_MAC&auth_code=YOUR_CODE"
   ```

### 已知问题

无

### 未来计划

- [ ] 支持 HTTPS/TLS
- [ ] 支持 Webhook 通知
- [ ] 支持设备分组管理
- [ ] 支持访问日志和审计
- [ ] 支持配置文件热重载
- [ ] 提供管理界面

---

## [1.0.0] - 2024-XX-XX

### 初始版本

- 基础 WOL 功能
- Google Authenticator 双因素认证
- GET 和 POST 接口支持
- 单一全局密钥管理
