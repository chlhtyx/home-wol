package main

import (
	_ "database/sql"
	_ "fmt"
	"home-wol/common"
	"home-wol/service"
	"log"

	"os"

	"github.com/labstack/echo/v4"
)

func main() {
	adminAuth, err := service.AdminAuthFromEnv()
	if err != nil {
		log.Fatalf("管理端认证配置错误: %v", err)
	}

	// 加载配置文件
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json" // 默认配置文件路径
	}

	_, err = common.GetConfig(configPath)
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

	service.RegisterRoutes(e, adminAuth, "static")

	log.Println("Web 管理界面: http://localhost:1323/admin/")
	e.Logger.Fatal(e.Start(":1323"))
}
