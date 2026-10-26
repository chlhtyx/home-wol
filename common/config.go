package common

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

// DeviceConfig 设备配置
type DeviceConfig struct {
	MAC    string `json:"mac"`
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

// Config 应用配置
type Config struct {
	BroadcastAddress string         `json:"broadcast_address,omitempty"`
	Devices          []DeviceConfig `json:"devices"`
}

var (
	config           *Config
	configOnce       sync.Once
	configErr        error
	configCacheMutex sync.RWMutex
)

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	// 标准化所有 MAC 地址为大写且无分隔符
	for i := range cfg.Devices {
		cfg.Devices[i].MAC = normalizeMAC(cfg.Devices[i].MAC)
	}

	return &cfg, nil
}

// GetConfig 获取配置（单例）
func GetConfig(path string) (*Config, error) {
	configOnce.Do(func() {
		cfg, err := LoadConfig(path)
		configCacheMutex.Lock()
		defer configCacheMutex.Unlock()
		config, configErr = cfg, err
	})
	configCacheMutex.RLock()
	defer configCacheMutex.RUnlock()
	return config, configErr
}

// ReloadConfig 刷新认证缓存，确保页面保存后的设备和密钥立即生效。
func ReloadConfig(path string) (*Config, error) {
	cfg, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}
	configCacheMutex.Lock()
	defer configCacheMutex.Unlock()
	config, configErr = cfg, nil
	return cfg, nil
}

// GetSecretByMAC 根据 MAC 地址获取对应的 secret
func GetSecretByMAC(mac string) (string, error) {
	mac = normalizeMAC(mac)
	configCacheMutex.RLock()
	defer configCacheMutex.RUnlock()

	if config == nil {
		return "", ErrConfigNotLoaded
	}

	for _, device := range config.Devices {
		if device.MAC == mac {
			return device.Secret, nil
		}
	}

	return "", ErrDeviceNotFound
}

// normalizeMAC 标准化 MAC 地址（大写且无分隔符）
func normalizeMAC(mac string) string {
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	mac = strings.ToUpper(mac)
	return mac
}
