package service

import (
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes shares the production route protections with integration tests.
func RegisterRoutes(e *echo.Echo, auth *AdminAuth, staticRoot string) {
	e.GET("/healthz", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/admin/")
	})
	e.GET("/login", func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		if auth.authenticated(c) {
			return c.Redirect(http.StatusSeeOther, "/admin/")
		}
		return c.File(filepath.Join(staticRoot, "login.html"))
	})
	e.File("/login.css", filepath.Join(staticRoot, "css", "login.css"))
	e.File("/login.js", filepath.Join(staticRoot, "js", "login.js"))
	e.POST("/api/login", auth.Login, adminCSRF)
	e.POST("/api/logout", auth.Logout, auth.RequireAPI, adminCSRF)

	admin := e.Group("/admin", auth.RequirePage)
	admin.GET("", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/admin/")
	})
	admin.Static("/", staticRoot)

	// WOL clients still authenticate with their device's dynamic code.
	e.GET("/wol", Wol)
	e.GET("/generate_code", GenerateCode, auth.RequireAPI)

	api := e.Group("/api", auth.RequireAPI, adminCSRF)
	api.GET("/settings", GetSettings)
	api.PUT("/settings", UpdateSettings)
	api.GET("/devices", GetDevices)
	api.POST("/devices", AddDevice)
	api.PUT("/devices/:mac", UpdateDevice)
	api.DELETE("/devices/:mac", DeleteDevice)
	api.GET("/devices/:mac/code", GetDeviceCode)
	api.GET("/devices/:mac/qrcode", GetDeviceQRCode)
}
