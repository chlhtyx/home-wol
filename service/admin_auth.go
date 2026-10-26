package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	adminCookieName = "home_wol_session"
	adminSessionTTL = 12 * time.Hour
)

// AdminAuth keeps opaque sessions in memory. Restarting invalidates all sessions.
type AdminAuth struct {
	username [32]byte
	password [32]byte
	mu       sync.Mutex
	sessions map[[32]byte]time.Time
	window   time.Time
	attempts int
	now      func() time.Time
}

func NewAdminAuth(username, password string) (*AdminAuth, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("必须设置 ADMIN_PASSWORD，ADMIN_USERNAME 默认为 admin")
	}
	return &AdminAuth{
		username: sha256.Sum256([]byte(username)),
		password: sha256.Sum256([]byte(password)),
		sessions: make(map[[32]byte]time.Time),
		now:      time.Now,
	}, nil
}

func AdminAuthFromEnv() (*AdminAuth, error) {
	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	return NewAdminAuth(username, os.Getenv("ADMIN_PASSWORD"))
}

func (a *AdminAuth) authenticated(c echo.Context) bool {
	cookie, err := c.Cookie(adminCookieName)
	if err != nil || len(cookie.Value) != 64 {
		return false
	}
	key := sha256.Sum256([]byte(cookie.Value))
	a.mu.Lock()
	defer a.mu.Unlock()
	expires, exists := a.sessions[key]
	if !exists || !a.now().Before(expires) {
		delete(a.sessions, key)
		return false
	}
	return true
}

func (a *AdminAuth) RequirePage(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		if !a.authenticated(c) {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		return next(c)
	}
}

func (a *AdminAuth) RequireAPI(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		if !a.authenticated(c) {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "请先登录管理端"})
		}
		return next(c)
	}
}

// A custom header prevents browser form submissions and cross-origin fetches
// from changing state. No cross-origin access is enabled for these routes.
func adminCSRF(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		switch c.Request().Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if c.Request().Header.Get("X-Requested-With") != "XMLHttpRequest" {
				return c.JSON(http.StatusForbidden, ErrorResponse{Error: "无效的请求来源，请刷新页面重试"})
			}
		}
		return next(c)
	}
}

func (a *AdminAuth) Login(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	// Bound login attempts and session storage for this single-admin service.
	now := a.now()
	a.mu.Lock()
	if a.window.IsZero() || !now.Before(a.window.Add(time.Minute)) {
		a.window, a.attempts = now, 0
	}
	a.attempts++
	limited := a.attempts > 20
	a.mu.Unlock()
	if limited {
		c.Response().Header().Set("Retry-After", "60")
		return c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "登录尝试过多，请一分钟后重试"})
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	reader := http.MaxBytesReader(c.Response(), c.Request().Body, 4096)
	if err := json.NewDecoder(reader).Decode(&credentials); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的登录信息"})
	}
	username := sha256.Sum256([]byte(credentials.Username))
	password := sha256.Sum256([]byte(credentials.Password))
	validUser := subtle.ConstantTimeCompare(username[:], a.username[:])
	validPassword := subtle.ConstantTimeCompare(password[:], a.password[:])
	if validUser&validPassword != 1 {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "账号或密码错误"})
	}

	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "创建登录会话失败"})
	}
	value := hex.EncodeToString(token[:])
	expires := now.Add(adminSessionTTL)
	a.mu.Lock()
	for key, expiry := range a.sessions {
		if !now.Before(expiry) {
			delete(a.sessions, key)
		}
	}
	if old, err := c.Cookie(adminCookieName); err == nil {
		delete(a.sessions, sha256.Sum256([]byte(old.Value)))
	}
	if len(a.sessions) >= 100 {
		a.mu.Unlock()
		return c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "登录会话过多，请稍后重试"})
	}
	a.sessions[sha256.Sum256([]byte(value))] = expires
	a.mu.Unlock()
	c.SetCookie(sessionCookie(c, value, expires, int(adminSessionTTL.Seconds())))
	return c.NoContent(http.StatusNoContent)
}

func sessionCookie(c echo.Context, value string, expires time.Time, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name: adminCookieName, Value: value, Path: "/",
		HttpOnly: true, Secure: c.IsTLS() || c.Scheme() == "https",
		SameSite: http.SameSiteStrictMode, Expires: expires, MaxAge: maxAge,
	}
}

func (a *AdminAuth) Logout(c echo.Context) error {
	if cookie, err := c.Cookie(adminCookieName); err == nil {
		a.mu.Lock()
		delete(a.sessions, sha256.Sum256([]byte(cookie.Value)))
		a.mu.Unlock()
	}
	c.SetCookie(sessionCookie(c, "", time.Unix(1, 0), -1))
	return c.NoContent(http.StatusNoContent)
}
