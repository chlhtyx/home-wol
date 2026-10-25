// API 基础路径
const API_BASE = '/api';

// 全局状态
let devices = [];
let currentEditingMac = null;

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    loadSettings();
    loadDevices();
});

// 加载网络设置；加载失败时禁用保存，避免误覆盖当前配置。
async function loadSettings() {
    const input = document.getElementById('broadcastAddress');
    const button = document.getElementById('saveSettingsButton');
    const status = document.getElementById('settingsStatus');
    const retry = document.getElementById('retrySettingsButton');
    input.disabled = true;
    button.disabled = true;
    retry.hidden = true;
    status.textContent = '正在加载网络设置…';
    status.className = 'settings-status';

    try {
        const response = await fetch(`${API_BASE}/settings`, { cache: 'no-store' });
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || '加载网络设置失败');
        input.value = data.broadcast_address;
        input.placeholder = '例如：192.168.1.255';
        input.disabled = false;
        button.disabled = false;
        status.textContent = `当前生效地址：${data.broadcast_address}`;
    } catch (error) {
        status.textContent = `加载失败：${error.message}`;
        status.className = 'settings-status error';
        retry.hidden = false;
    }
}

async function saveSettings(event) {
    event.preventDefault();
    const input = document.getElementById('broadcastAddress');
    const button = document.getElementById('saveSettingsButton');
    const status = document.getElementById('settingsStatus');
    if (button.disabled) return;

    const address = input.value.trim();
    if (!/^(?:\d{1,3}\.){3}\d{1,3}$/.test(address) || address.split('.').some(part => Number(part) > 255)) {
        showToast('请输入有效的 IPv4 广播地址，例如 192.168.1.255', 'warning');
        input.focus();
        return;
    }

    input.disabled = true;
    button.disabled = true;
    button.textContent = '保存中…';
    try {
        const response = await fetch(`${API_BASE}/settings`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ broadcast_address: address })
        });
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || '保存网络设置失败');
        input.value = data.broadcast_address;
        status.textContent = `已保存，当前生效地址：${data.broadcast_address}`;
        status.className = 'settings-status success';
        showToast('广播地址已保存，立即生效', 'success');
    } catch (error) {
        status.textContent = `保存失败：${error.message}`;
        status.className = 'settings-status error';
        showToast(error.message, 'error');
    } finally {
        input.disabled = false;
        button.disabled = false;
        button.textContent = '保存设置';
    }
}

// 加载设备列表
async function loadDevices() {
    const deviceList = document.getElementById('deviceList');
    deviceList.innerHTML = '<div class="loading">加载中...</div>';

    try {
        const response = await fetch(`${API_BASE}/devices`);
        if (!response.ok) {
            throw new Error('加载设备列表失败');
        }

        const data = await response.json();
        devices = data.devices || [];

        if (devices.length === 0) {
            deviceList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">🖥️</div>
                    <p>暂无设备</p>
                    <p>点击上方"添加设备"按钮添加第一个设备</p>
                </div>
            `;
            return;
        }

        deviceList.innerHTML = devices.map(device => createDeviceCard(device)).join('');
    } catch (error) {
        console.error('加载设备列表失败:', error);
        deviceList.innerHTML = `
            <div class="empty-state">
                <p style="color: var(--danger-color);">加载失败: ${error.message}</p>
                <button class="btn btn-primary" onclick="loadDevices()" style="margin-top: 12px;">重试</button>
            </div>
        `;
    }
}

// 创建设备卡片
function createDeviceCard(device) {
    const mac = device.mac || '';
    const name = device.name || '未命名设备';
    const secret = device.secret || '';

    // 格式化 MAC 地址显示
    const formattedMac = formatMac(mac);

    return `
        <div class="device-item" data-mac="${mac}">
            <div class="device-header">
                <div>
                    <div class="device-name">${escapeHtml(name)}</div>
                    <div class="device-mac">${formattedMac}</div>
                </div>
            </div>
            <div class="device-actions">
                <button class="btn btn-success btn-small" onclick="testWol('${mac}')">
                    <span>⚡</span> 测试唤醒
                </button>
                <button class="btn btn-secondary btn-small" onclick="showQrcode('${mac}', '${escapeHtml(name)}', '${secret}')">
                    <span>📱</span> 二维码
                </button>
                <button class="btn btn-warning btn-small" onclick="editDevice('${mac}')">
                    <span>✏️</span> 编辑
                </button>
                <button class="btn btn-danger btn-small" onclick="deleteDevice('${mac}', '${escapeHtml(name)}')">
                    <span>🗑️</span> 删除
                </button>
            </div>
        </div>
    `;
}

// 格式化 MAC 地址
function formatMac(mac) {
    if (!mac) return '';
    const cleaned = mac.replace(/[:-]/g, '').toUpperCase();
    if (cleaned.length !== 12) return mac;
    return cleaned.match(/.{2}/g).join(':');
}

// 测试唤醒功能
async function testWol(mac) {
    const device = devices.find(d => d.mac === mac);
    if (!device) {
        showToast('设备不存在', 'error');
        return;
    }

    // 生成动态验证码
    try {
        const codeResponse = await fetch(`${API_BASE}/devices/${mac}/code`);
        if (!codeResponse.ok) {
            throw new Error('获取验证码失败');
        }

        const codeData = await codeResponse.json();
        const authCode = codeData.code;

        // 发送唤醒指令
        const wolResponse = await fetch(`/wol?mac=${mac}&auth_code=${authCode}`);
        if (!wolResponse.ok) {
            const errorText = await wolResponse.text();
            throw new Error(errorText || '唤醒失败');
        }

        showToast('唤醒指令已发送', 'success');
    } catch (error) {
        console.error('测试唤醒失败:', error);
        showToast(`唤醒失败: ${error.message}`, 'error');
    }
}

// 显示二维码
async function showQrcode(mac, name, secret) {
    const modal = document.getElementById('qrcodeModal');
    const content = document.getElementById('qrcodeContent');
    const info = document.getElementById('qrcodeDeviceInfo');

    modal.classList.add('show');
    content.innerHTML = '<div class="loading">生成中...</div>';
    info.innerHTML = '';

    try {
        const response = await fetch(`${API_BASE}/devices/${mac}/qrcode`);
        if (!response.ok) {
            throw new Error('生成二维码失败');
        }

        const data = await response.json();

        content.innerHTML = `<img src="${data.qrcode}" alt="设备二维码">`;

        info.innerHTML = `
            <div class="device-info-item">
                <span class="device-info-label">设备名称</span>
                <span class="device-info-value">${escapeHtml(name)}</span>
            </div>
            <div class="device-info-item">
                <span class="device-info-label">MAC 地址</span>
                <span class="device-info-value">${formatMac(mac)}</span>
            </div>
            <div class="device-info-item">
                <span class="device-info-label">密钥</span>
                <span class="device-info-value">${secret}</span>
            </div>
            <div class="device-info-item">
                <span class="device-info-label">当前验证码</span>
                <span class="device-info-value" style="font-size: 1.2rem; font-weight: bold; color: var(--primary-color);">${data.code}</span>
            </div>
        `;
    } catch (error) {
        console.error('生成二维码失败:', error);
        content.innerHTML = `<p style="color: var(--danger-color);">生成失败: ${error.message}</p>`;
    }
}

// 关闭二维码模态框
function closeQrcodeModal() {
    const modal = document.getElementById('qrcodeModal');
    modal.classList.remove('show');
}

// 显示添加设备模态框
function showAddModal() {
    currentEditingMac = null;
    document.getElementById('modalTitle').textContent = '添加设备';
    document.getElementById('originalMac').value = '';
    document.getElementById('deviceName').value = '';
    document.getElementById('deviceMac').value = '';
    document.getElementById('deviceMac').disabled = false;
    document.getElementById('secretGroup').style.display = 'block';
    document.getElementById('deviceSecret').value = '';
    document.getElementById('secretHelpText').textContent = '用于生成动态验证码的密钥';
    document.getElementById('secretWarning').style.display = 'none';

    const modal = document.getElementById('deviceModal');
    modal.classList.add('show');
}

// 编辑设备
function editDevice(mac) {
    const device = devices.find(d => d.mac === mac);
    if (!device) {
        showToast('设备不存在', 'error');
        return;
    }

    currentEditingMac = mac;
    document.getElementById('modalTitle').textContent = '编辑设备';
    document.getElementById('originalMac').value = mac;
    document.getElementById('deviceName').value = device.name || '';
    document.getElementById('deviceMac').value = formatMac(device.mac);
    document.getElementById('deviceMac').disabled = true; // MAC 地址不可编辑

    // 显示密钥字段，但默认不显示当前密钥（安全考虑）
    document.getElementById('secretGroup').style.display = 'block';
    document.getElementById('deviceSecret').value = device.secret || '';
    document.getElementById('secretHelpText').textContent = '留空则不修改密钥，输入新值则更新';
    document.getElementById('secretWarning').style.display = 'block';

    const modal = document.getElementById('deviceModal');
    modal.classList.add('show');
}

// 关闭模态框
function closeModal() {
    const modal = document.getElementById('deviceModal');
    modal.classList.remove('show');
    currentEditingMac = null;
}

// 生成密钥
function generateSecret() {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
    let secret = '';
    for (let i = 0; i < 32; i++) {
        secret += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    document.getElementById('deviceSecret').value = secret;
}

// 处理表单提交
async function handleSubmit(event) {
    event.preventDefault();

    const name = document.getElementById('deviceName').value.trim();
    const mac = document.getElementById('deviceMac').value.trim();
    let secret = document.getElementById('deviceSecret').value.trim();

    // 验证
    if (!name) {
        showToast('请输入设备名称', 'warning');
        return;
    }

    if (!mac) {
        showToast('请输入 MAC 地址', 'warning');
        return;
    }

    // 标准化 MAC 地址
    const normalizedMac = mac.replace(/[:-]/g, '').toUpperCase();

    if (normalizedMac.length !== 12 || !/^[0-9A-F]{12}$/.test(normalizedMac)) {
        showToast('MAC 地址格式不正确', 'warning');
        return;
    }

    // 添加设备时生成密钥
    if (!currentEditingMac && !secret) {
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
        secret = '';
        for (let i = 0; i < 32; i++) {
            secret += chars.charAt(Math.floor(Math.random() * chars.length));
        }
    }

    try {
        if (currentEditingMac) {
            // 更新设备
            const updateData = {
                name: name,
                mac: normalizedMac
            };

            // 如果用户输入了新密钥，则包含在更新请求中
            if (secret && secret !== devices.find(d => d.mac === currentEditingMac)?.secret) {
                updateData.secret = secret;

                // 确认警告
                if (!confirm('⚠️ 警告：修改密钥后，需要重新在 Google Authenticator 中配置此设备，否则验证将失败！\n\n确定要修改密钥吗？')) {
                    return;
                }
            }

            const response = await fetch(`${API_BASE}/devices/${currentEditingMac}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(updateData)
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || '更新设备失败');
            }

            showToast('设备更新成功', 'success');
        } else {
            // 添加设备
            const response = await fetch(`${API_BASE}/devices`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name: name,
                    mac: normalizedMac,
                    secret: secret
                })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || '添加设备失败');
            }

            showToast('设备添加成功', 'success');
        }

        closeModal();
        loadDevices();
    } catch (error) {
        console.error('保存设备失败:', error);
        showToast(error.message, 'error');
    }
}

// 删除设备
async function deleteDevice(mac, name) {
    if (!confirm(`确定要删除设备 "${name}" 吗？\n\nMAC 地址: ${formatMac(mac)}`)) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/devices/${mac}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '删除设备失败');
        }

        showToast('设备删除成功', 'success');
        loadDevices();
    } catch (error) {
        console.error('删除设备失败:', error);
        showToast(error.message, 'error');
    }
}

// 显示 Toast 提示
function showToast(message, type = 'info') {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = `toast ${type} show`;

    setTimeout(() => {
        toast.classList.remove('show');
    }, 3000);
}

// HTML 转义
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 点击模态框外部关闭
window.onclick = function(event) {
    const deviceModal = document.getElementById('deviceModal');
    const qrcodeModal = document.getElementById('qrcodeModal');

    if (event.target === deviceModal) {
        closeModal();
    }

    if (event.target === qrcodeModal) {
        closeQrcodeModal();
    }
}
