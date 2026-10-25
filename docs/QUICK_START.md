# 快速开始指南

> 命令均从项目根目录执行；当前目录布局和启动方式见[项目 README](../README.md)。旧目录示意保留作参考。

## 功能概述

本项目已升级为支持多设备独立密钥管理的 WOL (Wake-On-LAN) 服务。每个设备可以配置独立的 Google Authenticator 密钥，提供更安全的管理方式。

## 主要变更

1. ✅ 新增配置文件支持 (`config.json`)，为不同 MAC 地址绑定不同的 secret
2. ✅ 改进验证逻辑，根据 MAC 地址自动查找对应的 secret
3. ✅ 新增 `/generate_code` 接口，用于生成校验码
4. ✅ 完善错误处理，提供更友好的错误提示

## 快速开始

### 1. 编译程序

```bash
# 编译主程序
go build -o bin/wol.exe .

# 编译密钥生成工具
go build -o bin/generate_secret.exe ./cmd/generate_secret
```

### 2. 生成设备密钥

运行密钥生成工具：

```bash
./bin/generate_secret.exe
```

按提示输入设备名称（如：我的电脑），程序会生成：
- Secret 密钥
- 当前验证码
- 二维码信息
- 配置文件示例

### 3. 配置设备

编辑 `config.json` 文件，添加设备信息：

```json
{
  "devices": [
    {
      "mac": "AABBCCDDEEFF",
      "name": "我的电脑",
      "secret": "生成的密钥"
    }
  ]
}
```

### 4. 配置 Google Authenticator

**方式一：扫描二维码**
1. 使用生成的二维码图片 URL
2. 用 Google Authenticator 应用扫描

**方式二：手动输入**
1. 打开 Google Authenticator 应用
2. 点击 "+" 添加账户
3. 选择"输入提供的密钥"
4. 输入设备名称和 Secret

### 5. 启动服务

```bash
# 使用默认配置文件路径
./bin/wol.exe

# 或指定配置文件路径
set CONFIG_PATH=/path/to/config.json
./bin/wol.exe
```

服务默认监听端口：1323

### 6. 测试验证

#### 6.1 生成校验码（测试用）

```bash
curl "http://localhost:1323/generate_code?mac=AABBCCDDEEFF"
```

响应示例：
```json
{
  "mac": "AABBCCDDEEFF",
  "code": "123456",
  "secret": "JBSWY3DPEHPK3PXP"
}
```

#### 6.2 唤醒设备（GET 方式）

```bash
curl "http://localhost:1323/wol?mac=AABBCCDDEEFF&auth_code=123456"
```

#### 6.3 唤醒设备（POST 方式）

```bash
curl -X POST "http://localhost:1323/wolv2" \
  -d "mac=AABBCCDDEEFF" \
  -d "auth_code=123456"
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| CONFIG_PATH | 配置文件路径 | config.json |
| BROADCAST_ADDRESS | 旧配置回退；广播地址优先在 Web 页面“网络设置”中保存 | 192.168.100.255 |
| SECRET | 全局密钥（已废弃） | - |

## 配置示例

### 单设备配置

```json
{
  "devices": [
    {
      "mac": "AABBCCDDEEFF",
      "name": "家庭电脑",
      "secret": "JBSWY3DPEHPK3PXP"
    }
  ]
}
```

### 多设备配置

```json
{
  "devices": [
    {
      "mac": "AABBCCDDEEFF",
      "name": "家庭电脑",
      "secret": "JBSWY3DPEHPK3PXP"
    },
    {
      "mac": "11:22:33:44:55:66",
      "name": "NAS服务器",
      "secret": "KRSXG5DSNFXAWZL"
    },
    {
      "mac": "00-11-22-33-44-55",
      "name": "办公电脑",
      "secret": "MFRGG3DFMZTWQ2LK"
    }
  ]
}
```

注意：MAC 地址支持多种格式（`AABBCCDDEEFF`、`AA:BB:CC:DD:EE:FF`、`AA-BB-CC-DD-EE-FF`），系统会自动标准化。

## API 接口

### 1. GET /wol

唤醒设备接口（GET 方式）

**参数：**
- `mac`: 设备 MAC 地址
- `auth_code`: Google Authenticator 验证码

**响应：**
- 成功：`指令发送成功`
- 失败：错误信息

### 2. POST /wolv2

唤醒设备接口（POST 方式）

**参数：**
- `mac`: 设备 MAC 地址
- `auth_code`: Google Authenticator 验证码

**响应：**
- 成功：`指令发送成功`
- 失败：错误信息

### 3. GET /generate_code

生成校验码接口（新增）

**参数：**
- `mac`: 设备 MAC 地址

**响应：**
```json
{
  "mac": "AABBCCDDEEFF",
  "code": "123456",
  "secret": "JBSWY3DPEHPK3PXP"
}
```

## 错误处理

### 常见错误及解决方案

#### 1. 配置未加载
```
错误信息: 配置未加载
```
**解决方案：**
- 检查 `config.json` 文件是否存在
- 检查文件格式是否正确（JSON 格式）
- 检查 `CONFIG_PATH` 环境变量是否正确

#### 2. 设备未找到
```
错误信息: 设备未找到
```
**解决方案：**
- 检查 MAC 地址是否在配置文件中
- MAC 地址格式会自动标准化，无需担心格式问题

#### 3. 验证失败
```
错误信息: 验证失败
```
**解决方案：**
- 检查验证码是否正确（6位数字）
- 检查验证码是否过期（30秒有效期）
- 确保设备时间与服务器时间同步
- 确认 Secret 配置正确

## 安全建议

1. **保护配置文件**：`config.json` 包含敏感信息，请确保文件权限设置正确
2. **使用强密钥**：使用 `generate_secret.exe` 生成的随机密钥
3. **定期更换密钥**：建议定期更换设备的 Secret
4. **传输加密**：生产环境建议使用 HTTPS
5. **访问控制**：配置防火墙规则，限制接口访问来源

## 测试

运行单元测试：

```bash
go test -v ./common
```

## Docker 部署

```dockerfile
FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o wol .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/wol .
COPY config.json .
EXPOSE 1323
CMD ["./wol"]
```

构建和运行：

```bash
# 构建镜像
docker build -t home-wol .

# 运行容器
docker run -d \
  -p 1323:1323 \
  -v $(pwd)/config.json:/root/config.json \
  -e BROADCAST_ADDRESS=192.168.1.255 \
  home-wol
```

## 故障排查

### 1. 无法唤醒设备

检查项：
- [ ] 设备是否支持 WOL
- [ ] 设备是否已开启 WOL 功能（BIOS 设置）
- [ ] 网络是否可达（ping 测试）
- [ ] 广播地址是否正确
- [ ] MAC 地址是否正确

### 2. 验证总是失败

检查项：
- [ ] Secret 是否配置正确
- [ ] 设备时间是否与服务器同步
- [ ] 验证码是否在有效期内（30秒）
- [ ] Google Authenticator 应用是否正常工作

## 版本兼容性

本版本保留了原有的全局 `SECRET` 环境变量支持，但建议使用新的配置文件方式管理多设备。

## 相关文档

- [多设备配置详解](./MULTI_DEVICE_CONFIG.md)
- [GoDoc](https://pkg.go.dev/github.com/yourusername/home-wol)

## 许可证

请参阅 [LICENSE](../LICENSE) 文件了解详情。
