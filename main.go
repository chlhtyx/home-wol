package main

import (
	_ "database/sql"
	_ "fmt"
	"home-wol/common"
	"home-wol/service"
	"log"
	"net/http"

	"os"

	"github.com/labstack/echo/v4"
)

func main() {

	// 加载配置文件
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json" // 默认配置文件路径
	}

	_, err := common.GetConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	log.Println("配置文件加载成功")

	// 初始化 API 服务
	service.InitAPI(configPath)

	// 兼容旧版本的全局 SECRET（可选）
	common.Secret = os.Getenv("SECRET")
	ip, err := service.GetBroadcastIP()
	if err != nil {
		log.Fatalf("加载广播地址失败: %v", err)
	}
	log.Println("广播地址:", ip)

	e := echo.New()

	// 静态文件服务 - Web 管理界面
	e.Static("/admin", "static")
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/admin/")
	})

	// 兼容旧版 API
	e.GET("/wol", service.Wol)
	e.GET("/generate_code", service.GenerateCode)

	// 新版 RESTful API
	api := e.Group("/api")
	api.GET("/settings", service.GetSettings)
	api.PUT("/settings", service.UpdateSettings)
	api.GET("/devices", service.GetDevices)
	api.POST("/devices", service.AddDevice)
	api.PUT("/devices/:mac", service.UpdateDevice)
	api.DELETE("/devices/:mac", service.DeleteDevice)
	api.GET("/devices/:mac/code", service.GetDeviceCode)
	api.GET("/devices/:mac/qrcode", service.GetDeviceQRCode)

	log.Println("Web 管理界面: http://localhost:1323/admin/")
	e.Logger.Fatal(e.Start(":1323"))
}
