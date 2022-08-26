package main

import (
	"home-wol/common"
	"home-wol/service"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

func main() {
	common.Secret = os.Getenv("SECRET")
	broadcastAddress := os.Getenv("BROADCAST_ADDRESS")
	//检查广播地址

	ip := net.ParseIP(broadcastAddress)
	if ip == nil {
		log.Fatalln("广播地址格式异常")
		os.Exit(-1)
	}
	service.BroadcastIP = ip

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/wol", service.Wol)

	e.Logger.Fatal(e.Start(":1323"))
}
