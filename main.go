package main

import (
	"home-wol/common"
	"home-wol/service"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

func main() {
	common.Secret = os.Getenv("SECRET")

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/wol", service.Wol)

	e.Logger.Fatal(e.Start(":1323"))
}
