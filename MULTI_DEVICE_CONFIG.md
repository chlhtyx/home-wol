# 多设备配置说明

## 功能概述

本项目现在支持为不同的 MAC 地址配置不同的 secret 密钥，实现多设备的独立验证。

## 配置文件

配置文件为 JSON 格式，默认路径为 `config.json`，可通过环境变量 `CONFIG_PATH` 指定。

### 配置文件格式

```json
{
  "devices": [
    {
      "mac": "AABBCCDDEEFF",
      "name": "设备名称",
      "secret": "BASE32编码的密钥"
    }
  ]
}
```

### 配置项说明

- `mac`: 设备的 MAC 地址（支持带分隔符和不带分隔符，如：`AA:BB:CC:DD:EE:FF` 或 `AABBCCDDEEFF`）
- `name`: 设备名称（可选，用于标识）
- `secret`: Google Authenticator 密钥（Base32 编码）

### 配置示例

```json
{
  "devices": [
    {
      "mac": "AABBCCDDEEFF",
      "name": "我的电脑",
      "secret": "JBSWY3DPEHPK3PXP"
    },
    {
      "mac": "11:22:33:44:55:66",
      "name": "NAS服务器",
      "secret": "KRSXG5DSNFXAWZL"
    }
  ]
}
```

## 接口说明

### 1. WOL 唤醒接口（GET）

**接口地址**: `/wol`

**请求参数**:
- `mac`: 设备 MAC 地址
- `auth_code`: Google Authenticator 动态验证码

**示例**:
```
GET /wol?mac=AABBCCDDEEFF&auth_code=123456
```

**响应**:
- 成功: `指令发送成功`
- 失败: `验证失败` 或具体错误信息

### 2. WOL 唤醒接口（POST）

**接口地址**: `/wolv2`

**请求参数**（form-data）:
- `mac`: 设备 MAC 地址
- `auth_code`: Google Authenticator 动态验证码

**示例**:
```
POST /wolv2
Content-Type: application/x-www-form-urlencoded

mac=AABBCCDDEEFF&auth_code=123456
```

### 3. 生成校验码接口（新增）

**接口地址**: `/generate_code`

**请求参数**:
- `mac`: 设备 MAC 地址

**示例**:
```
GET /generate_code?mac=AABBCCDDEEFF
```

**响应**:
```json
{
  "mac": "AABBCCDDEEFF",
  "code": "123456",
  "secret": "JBSWY3DPEHPK3PXP"
}
```

## 使用流程

### 1. 配置设备

1. 编辑 `config.json` 文件
2. 为每个设备配置 MAC 地址和对应的 secret
3. 生成 secret 可使用以下 Go 代码：

```go
package main

import (
    "fmt"
    "home-wol/common"
)

func main() {
    auth := common.NewGoogleAuth()
    secret := auth.GetSecret()
    fmt.Println("生成的 Secret:", secret)
}
```

### 2. 配置 Google Authenticator

对于每个设备的 secret：

1. 打开 Google Authenticator 应用
2. 添加新账户
3. 选择"输入提供的密钥"
4. 输入设备名称和 secret
5. 扫描二维码或手动输入密钥

生成二维码的 URL 格式：
```
otpauth://totp/设备名称?secret=YOUR_SECRET
```

### 3. 测试验证

1. 获取当前时间的动态验证码（从 Google Authenticator 应用）
2. 使用生成校验码接口测试：
   ```
   GET /generate_code?mac=AABBCCDDEEFF
   ```
3. 对比生成的验证码与应用中的验证码是否一致
4. 使用 WOL 接口唤醒设备

## 环境变量

- `CONFIG_PATH`: 配置文件路径（默认：`config.json`）
- `BROADCAST_ADDRESS`: 广播地址（如：`192.168.1.255`）
- `SECRET`: 全局密钥（已废弃，保留用于兼容旧版本）

## 错误处理

### 配置未加载
- 错误信息：`配置未加载`
- 解决方案：检查 `config.json` 文件是否存在且格式正确

### 设备未找到
- 错误信息：`设备未找到`
- 解决方案：检查 MAC 地址是否在配置文件中

### 验证失败
- 错误信息：`验证失败`
- 解决方案：
  - 检查验证码是否正确（30秒内有效）
  - 检查设备时间是否同步
  - 确认 secret 配置正确

## 安全建议

1. **保护配置文件**: `config.json` 包含敏感信息，请确保文件权限设置正确
2. **使用强密钥**: 使用足够长度的随机密钥（建议 16+ 字符）
3. **定期更换**: 建议定期更换密钥
4. **传输加密**: 生产环境建议使用 HTTPS
5. **访问控制**: 配置防火墙规则，限制接口访问来源

## 向后兼容

本版本保留了原有的全局 `SECRET` 环境变量支持，但建议使用新的配置文件方式管理多设备。
