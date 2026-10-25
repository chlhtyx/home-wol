package service

import (
	"fmt"
	"home-wol/common"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

const defaultBroadcastAddress = "192.168.100.255"

type Settings struct {
	BroadcastAddress string `json:"broadcast_address"`
}

func parseBroadcastAddress(address string) (net.IP, error) {
	addr, err := netip.ParseAddr(strings.TrimSpace(address))
	if err != nil || !addr.Is4() || addr.IsUnspecified() || addr.IsLoopback() || addr.IsMulticast() {
		return nil, fmt.Errorf("请输入有效的 IPv4 广播地址，例如 192.168.1.255")
	}
	return net.IP(addr.AsSlice()), nil
}

func resolveBroadcastIP(cfg *common.Config) (net.IP, error) {
	address := cfg.BroadcastAddress
	if address == "" {
		// 旧配置仍可从环境变量读取；页面保存的配置始终优先。
		address = os.Getenv("BROADCAST_ADDRESS")
	}
	if address == "" {
		address = defaultBroadcastAddress
	}
	return parseBroadcastAddress(address)
}

// GetBroadcastIP 返回持久化配置中的当前广播地址，与配置写入共用锁。
func GetBroadcastIP() (net.IP, error) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	cfg, err := common.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return resolveBroadcastIP(cfg)
}

func GetSettings(c echo.Context) error {
	ip, err := GetBroadcastIP()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "加载网络设置失败: " + err.Error()})
	}
	return c.JSON(http.StatusOK, Settings{BroadcastAddress: ip.String()})
}

func UpdateSettings(c echo.Context) error {
	var settings Settings
	if err := c.Bind(&settings); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的请求参数"})
	}
	ip, err := parseBroadcastAddress(settings.BroadcastAddress)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	configMutex.Lock()
	defer configMutex.Unlock()

	cfg, err := common.LoadConfig(configPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "加载配置文件失败: " + err.Error()})
	}
	cfg.BroadcastAddress = ip.String()
	if err := saveConfig(cfg); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "保存网络设置失败: " + err.Error()})
	}
	return c.JSON(http.StatusOK, Settings{BroadcastAddress: cfg.BroadcastAddress})
}
