package service

import (
	"encoding/json"
	"home-wol/common"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

var (
	configPath  string
	configMutex sync.RWMutex
)

// InitAPI 初始化 API 服务
func InitAPI(path string) {
	configPath = path
}

// DeviceResponse 设备响应
type DeviceResponse struct {
	Devices []common.DeviceConfig `json:"devices"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// QRCodeResponse 二维码响应
type QRCodeResponse struct {
	Qrcode string `json:"qrcode"`
	Code   string `json:"code"`
	Secret string `json:"secret"`
}

// GetDevices 获取设备列表
func GetDevices(c echo.Context) error {
	configMutex.RLock()
	defer configMutex.RUnlock()

	config, err := common.LoadConfig(configPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "加载配置文件失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, DeviceResponse{
		Devices: config.Devices,
	})
}

// AddDevice 添加设备
func AddDevice(c echo.Context) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	var device common.DeviceConfig
	if err := c.Bind(&device); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "无效的请求参数: " + err.Error(),
		})
	}

	// 验证设备信息
	if device.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "设备名称不能为空",
		})
	}

	if device.MAC == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "MAC 地址不能为空",
		})
	}

	// 标准化 MAC 地址
	device.MAC = normalizeMAC(device.MAC)

	// 验证 MAC 地址格式
	if len(device.MAC) != 12 || !isValidMAC(device.MAC) {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "MAC 地址格式不正确",
		})
	}

	// 验证密钥
	if device.Secret == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "密钥不能为空",
		})
	}

	// 加载现有配置
	config, err := common.LoadConfig(configPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "加载配置文件失败: " + err.Error(),
		})
	}

	// 检查设备是否已存在
	for _, d := range config.Devices {
		if d.MAC == device.MAC {
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error: "设备已存在",
			})
		}
	}

	// 添加设备
	config.Devices = append(config.Devices, device)

	// 保存配置
	if err := saveConfig(config); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "保存配置失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, device)
}

// UpdateDevice 更新设备
func UpdateDevice(c echo.Context) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	mac := c.Param("mac")
	if mac == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "MAC 地址不能为空",
		})
	}

	// 标准化 MAC 地址
	mac = normalizeMAC(mac)

	var device common.DeviceConfig
	if err := c.Bind(&device); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "无效的请求参数: " + err.Error(),
		})
	}

	// 验证设备信息
	if device.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "设备名称不能为空",
		})
	}

	// 加载现有配置
	config, err := common.LoadConfig(configPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "加载配置文件失败: " + err.Error(),
		})
	}

	// 查找并更新设备
	found := false
	for i, d := range config.Devices {
		if d.MAC == mac {
			config.Devices[i].Name = device.Name
			// 如果提供了新密钥则更新，否则保持原密钥
			if device.Secret != "" {
				config.Devices[i].Secret = device.Secret
			} else {
				config.Devices[i].Secret = d.Secret
			}
			found = true
			break
		}
	}

	if !found {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "设备不存在",
		})
	}

	// 保存配置
	if err := saveConfig(config); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "保存配置失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, config.Devices)
}

// DeleteDevice 删除设备
func DeleteDevice(c echo.Context) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	mac := c.Param("mac")
	if mac == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "MAC 地址不能为空",
		})
	}

	// 标准化 MAC 地址
	mac = normalizeMAC(mac)

	// 加载现有配置
	config, err := common.LoadConfig(configPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "加载配置文件失败: " + err.Error(),
		})
	}

	// 查找并删除设备
	found := false
	newDevices := make([]common.DeviceConfig, 0, len(config.Devices)-1)
	for _, d := range config.Devices {
		if d.MAC == mac {
			found = true
		} else {
			newDevices = append(newDevices, d)
		}
	}

	if !found {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "设备不存在",
		})
	}

	config.Devices = newDevices

	// 保存配置
	if err := saveConfig(config); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "保存配置失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "删除成功",
	})
}

// GetDeviceCode 获取设备验证码
func GetDeviceCode(c echo.Context) error {
	mac := c.Param("mac")
	if mac == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "MAC 地址不能为空",
		})
	}

	// 标准化 MAC 地址
	mac = normalizeMAC(mac)

	// 获取密钥
	secret, err := common.GetSecretByMAC(mac)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "设备不存在或未找到密钥",
		})
	}

	// 生成验证码
	auth := common.NewGoogleAuth()
	code, err := auth.GetCode(secret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "生成验证码失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"mac":    mac,
		"code":   code,
		"secret": secret,
	})
}

// GetDeviceQRCode 获取设备二维码
func GetDeviceQRCode(c echo.Context) error {
	mac := c.Param("mac")
	if mac == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "MAC 地址不能为空",
		})
	}

	// 标准化 MAC 地址
	mac = normalizeMAC(mac)

	// 加载配置获取设备名称
	config, err := common.LoadConfig(configPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "加载配置文件失败: " + err.Error(),
		})
	}

	var deviceName string
	var secret string
	for _, d := range config.Devices {
		if d.MAC == mac {
			deviceName = d.Name
			secret = d.Secret
			break
		}
	}

	if deviceName == "" {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "设备不存在",
		})
	}

	// 生成验证码
	auth := common.NewGoogleAuth()
	code, err := auth.GetCode(secret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "生成验证码失败: " + err.Error(),
		})
	}

	// 生成二维码（使用第三方 API）
	qrcodeURL := auth.GetQrcodeUrl(deviceName, secret)

	return c.JSON(http.StatusOK, QRCodeResponse{
		Qrcode: qrcodeURL,
		Code:   code,
		Secret: secret,
	})
}

// saveConfig 保存配置到文件
func saveConfig(config *common.Config) error {
	// 创建备份
	if err := backupConfig(); err != nil {
		return err
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	// 更新全局配置（需要重新加载以刷新缓存）
	common.LoadConfig(configPath)

	return nil
}

// backupConfig 备份配置文件
func backupConfig() error {
	// 读取现有配置
	data, err := os.ReadFile(configPath)
	if err != nil {
		// 如果文件不存在，不需要备份
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// 创建备份文件名
	ext := filepath.Ext(configPath)
	base := configPath[:len(configPath)-len(ext)]
	backupPath := base + ".backup" + ext

	// 写入备份
	return os.WriteFile(backupPath, data, 0644)
}

// normalizeMAC 标准化 MAC 地址
func normalizeMAC(mac string) string {
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	mac = strings.ToUpper(mac)
	return mac
}

// isValidMAC 验证 MAC 地址格式
func isValidMAC(mac string) bool {
	if len(mac) != 12 {
		return false
	}

	for _, c := range mac {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}
