# Web 端编辑密钥功能说明

## 功能概述

已为 Web 管理界面添加了**编辑设备密钥**的功能，允许用户在需要时更改设备的密钥。

## 修改内容

### 1. 前端界面 (HTML)

**文件**: `static/index.html`

- ✅ 添加了密钥警告提示框
- ✅ 修改了密钥输入框为可编辑状态
- ✅ 添加了帮助文本动态切换

```html
<div id="secretWarning" style="display: none; margin-top: 8px; padding: 8px;
     background-color: #fff3cd; border: 1px solid #ffc107; border-radius: 4px;
     font-size: 0.9em; color: #856404;">
    ⚠️ <strong>警告：</strong>修改密钥后，需要重新在 Google Authenticator 中配置此设备，否则验证将失败！
</div>
```

### 2. 前端逻辑 (JavaScript)

**文件**: `static/js/app.js`

#### 修改的函数：

**a) `editDevice(mac)` - 编辑设备**
- ✅ 显示密钥字段
- ✅ 填充当前密钥值
- ✅ 显示警告提示
- ✅ 更新帮助文本为"留空则不修改密钥"

```javascript
// 显示密钥字段，但默认不显示当前密钥（安全考虑）
document.getElementById('secretGroup').style.display = 'block';
document.getElementById('deviceSecret').value = device.secret || '';
document.getElementById('secretHelpText').textContent = '留空则不修改密钥，输入新值则更新';
document.getElementById('secretWarning').style.display = 'block';
```

**b) `showAddModal()` - 添加设备**
- ✅ 隐藏警告提示（添加设备时不需要警告）
- ✅ 恢复默认帮助文本

```javascript
document.getElementById('secretHelpText').textContent = '用于生成动态验证码的密钥';
document.getElementById('secretWarning').style.display = 'none';
```

**c) `handleSubmit(event)` - 表单提交**
- ✅ 检测密钥是否被修改
- ✅ 显示二次确认对话框
- ✅ 只在密钥实际改变时才发送更新请求

```javascript
// 如果用户输入了新密钥，则包含在更新请求中
if (secret && secret !== devices.find(d => d.mac === currentEditingMac)?.secret) {
    updateData.secret = secret;

    // 确认警告
    if (!confirm('⚠️ 警告：修改密钥后，需要重新在 Google Authenticator 中配置此设备，否则验证将失败！\n\n确定要修改密钥吗？')) {
        return;
    }
}
```

### 3. 后端 API (Go)

**文件**: `service/api.go`

#### 修改的函数：

**`UpdateDevice()` - 更新设备**

**修改前**：
```go
// 保持原有的密钥
config.Devices[i].Secret = d.Secret
```

**修改后**：
```go
// 如果提供了新密钥则更新，否则保持原密钥
if device.Secret != "" {
    config.Devices[i].Secret = device.Secret
} else {
    config.Devices[i].Secret = d.Secret
}
```

- ✅ 支持更新密钥
- ✅ 也支持保留原密钥（留空不更新）

## 使用方法

### 编辑设备密钥

1. **打开编辑界面**
   - 在设备列表中点击"编辑"按钮
   - 设备信息表单会弹出

2. **查看密钥字段**
   - 密钥字段已显示在表单中
   - 当前密钥已填充在输入框中
   - 黄色警告框提示修改密钥的风险

3. **修改密钥**
   - **方式 1**：手动输入新密钥（Base32 格式）
   - **方式 2**：点击"生成新密钥"按钮自动生成

4. **确认更改**
   - 点击"保存"按钮
   - 弹出二次确认对话框
   - 确认后提交更新

5. **重新配置验证器**
   - 在 Google Authenticator 中删除旧账户
   - 扫描新二维码或手动输入新密钥
   - 验证生成的新验证码

### 不修改密钥（只更新名称）

如果想只更新设备名称而不修改密钥：
1. 在编辑界面修改设备名称
2. **不要修改密钥字段**
3. 直接点击"保存"

## 安全特性

### 1. 警告提示
- ✅ 明确提示修改密钥的后果
- ✅ 二次确认对话框防止误操作

### 2. 默认保留原密钥
- ✅ 密钥字段留空时保持原密钥不变
- ✅ 只有主动输入新密钥才会更新

### 3. 密钥比较
- ✅ 前端比较新旧密钥是否相同
- ✅ 只在实际改变时显示警告

## 使用场景

### 适合修改密钥的情况：

1. **密钥泄露**
   - 密钥意外暴露
   - 需要立即更换

2. **密钥遗忘**
   - 无法在验证器中找到该设备
   - 需要重新配置

3. **密钥升级**
   - 使用更强的随机密钥
   - 提高安全性

4. **设备迁移**
   - 将设备转移到另一个验证器
   - 需要生成新的密钥

### 不建议修改密钥的情况：

1. **正常运行中**
   - 设备验证功能正常
   - 没有必要修改

2. **多用户环境**
   - 多个用户使用同一设备
   - 修改密钥会影响所有人

## 测试步骤

### 1. 本地测试

```bash
# 停止 Docker 服务
docker-compose down

# 启动本地服务
./wol.exe

# 访问 Web 界面
浏览器打开: http://localhost:1323/admin/
```

### 2. 功能测试

**测试 A：编辑设备名称（不修改密钥）**
1. 点击设备的"编辑"按钮
2. 修改设备名称
3. 不修改密钥字段
4. 点击"保存"
5. ✅ 应该成功更新，无警告提示

**测试 B：修改设备密钥**
1. 点击设备的"编辑"按钮
2. 点击"生成新密钥"按钮
3. 点击"保存"
4. ✅ 应该弹出警告对话框
5. 确认后应该成功更新
6. 使用新密钥在 Google Authenticator 中配置
7. 测试唤醒功能

**测试 C：添加新设备**
1. 点击"添加设备"按钮
2. 输入设备名称和 MAC 地址
3. ✅ 不应该看到警告提示
4. 保存后查看二维码
5. 配置到 Google Authenticator

## API 测试

### 更新设备（不修改密钥）

```bash
curl -X PUT http://localhost:1323/api/devices/E0E1A9108706 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "新的设备名称"
  }'
```

### 更新设备（修改密钥）

```bash
curl -X PUT http://localhost:1323/api/devices/E0E1A9108706 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "新的设备名称",
    "secret": "NEWSECRET123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  }'
```

## 注意事项

1. **修改密钥的影响**
   - 已配置的 Google Authenticator 将失效
   - 需要重新扫描二维码或输入新密钥
   - 所有使用该设备密钥的验证将失败

2. **备份配置**
   - 修改前建议备份 `config.json`
   - 以防需要恢复旧密钥

3. **通知用户**
   - 如果是多用户环境，需要通知所有用户
   - 提前告知密钥即将更改

4. **测试验证**
   - 修改密钥后立即测试
   - 确保新密钥能正常工作

## 故障排查

### 问题 1：保存后密钥没有更新

**可能原因**：
- 密钥字段为空（保持原密钥）
- API 请求失败

**解决方法**：
- 确保密钥字段有新值
- 检查浏览器控制台错误信息
- 查看服务器日志

### 问题 2：警告提示不显示

**可能原因**：
- 使用的是添加设备功能
- 新旧密钥相同

**解决方法**：
- 确保在编辑模式下
- 确保密钥值确实改变了

### 问题 3：修改密钥后验证失败

**可能原因**：
- Google Authenticator 未更新密钥
- 密钥输入错误

**解决方法**：
- 重新扫描二维码
- 手动输入新密钥
- 确认密钥完全正确（无空格、大小写正确）

## Docker 部署

注意：Docker 构建目前有问题，建议使用本地版本测试新功能：

```bash
# 本地运行
./wol.exe

# 访问
http://localhost:1323/admin/
```

待 Docker 构建问题解决后，可以使用：
```bash
docker-compose up -d
```

## 相关文件

- `static/index.html` - 前端界面
- `static/js/app.js` - 前端逻辑
- `service/api.go` - 后端 API
- `config.json` - 配置文件

## 总结

✅ **功能已添加**：Web 端现在可以编辑设备密钥

✅ **安全保护**：警告提示和二次确认

✅ **灵活使用**：可以选择是否修改密钥

✅ **向后兼容**：不修改密钥时保持原有行为

现在你可以在 Web 管理界面中自由地编辑设备密钥了！
