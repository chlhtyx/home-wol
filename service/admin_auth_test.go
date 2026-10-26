package service

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

const testAdminPassword = "test-only-password-42"

func adminTestServer(t *testing.T) (*echo.Echo, *AdminAuth) {
	t.Helper()
	settingsTestAPI(t, "192.168.1.255")
	auth, err := NewAdminAuth("admin", testAdminPassword)
	if err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	RegisterRoutes(e, auth, "../static")
	return e, auth
}

func adminRequest(e *echo.Echo, method, path, body string, cookie *http.Cookie, csrf bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if csrf {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func adminLogin(t *testing.T, e *echo.Echo, cookie *http.Cookie) *http.Cookie {
	t.Helper()
	rec := adminRequest(e, http.MethodPost, "/api/login", `{"username":"admin","password":"`+testAdminPassword+`"}`, cookie, true)
	if rec.Code != http.StatusNoContent || len(rec.Result().Cookies()) != 1 {
		t.Fatalf("login failed: %d %s", rec.Code, rec.Body.String())
	}
	return rec.Result().Cookies()[0]
}

func TestAdminRoutesRequireLogin(t *testing.T) {
	e, _ := adminTestServer(t)
	for _, path := range []string{"/admin", "/admin/", "/admin/index.html", "/admin/js/app.js", "/admin/css/style.css"} {
		rec := adminRequest(e, http.MethodGet, path, "", nil, false)
		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" {
			t.Errorf("%s was not redirected to login: %d", path, rec.Code)
		}
	}
	for _, route := range []struct{ method, path string }{
		{"GET", "/api/settings"}, {"PUT", "/api/settings"},
		{"GET", "/api/devices"}, {"POST", "/api/devices"},
		{"PUT", "/api/devices/AABBCCDDEEFF"}, {"DELETE", "/api/devices/AABBCCDDEEFF"},
		{"GET", "/api/devices/AABBCCDDEEFF/code"}, {"GET", "/api/devices/AABBCCDDEEFF/qrcode"},
		{"GET", "/generate_code?mac=AABBCCDDEEFF"}, {"POST", "/api/logout"},
	} {
		rec := adminRequest(e, route.method, route.path, "{}", nil, true)
		if rec.Code != http.StatusUnauthorized || rec.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("unprotected %s %s: %d %s", route.method, route.path, rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "secret") {
			t.Errorf("unauthenticated response exposed device secrets")
		}
	}
	for _, path := range []string{"/login", "/login.js", "/login.css", "/healthz"} {
		if rec := adminRequest(e, "GET", path, "", nil, false); rec.Code != http.StatusOK {
			t.Errorf("public route %s failed: %d", path, rec.Code)
		}
	}
	// Device-code clients must still reach their own authentication handler.
	rec := adminRequest(e, "GET", "/wol", "", nil, false)
	if !strings.Contains(rec.Body.String(), "验证码不能为空") {
		t.Fatal("WOL no longer uses device-code authentication")
	}
}

func TestAdminStaticPathsCannotBypassLoginOrEscapeRoot(t *testing.T) {
	root := t.TempDir()
	staticRoot := filepath.Join(root, "static")
	if err := os.Mkdir(staticRoot, 0700); err != nil {
		t.Fatal(err)
	}
	const adminContent = "protected admin page"
	const privateContent = "private file outside static root"
	if err := os.WriteFile(filepath.Join(staticRoot, "index.html"), []byte(adminContent), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "private.txt"), []byte(privateContent), 0600); err != nil {
		t.Fatal(err)
	}
	auth, err := NewAdminAuth("admin", testAdminPassword)
	if err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	RegisterRoutes(e, auth, staticRoot)
	cookie := adminLogin(t, e, nil)
	if rec := adminRequest(e, "GET", "/admin/", "", cookie, false); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), adminContent) {
		t.Fatal("authenticated static page is unavailable")
	}
	// Keep the raw URL escapes: these must exercise Echo's routing and file
	// resolution together, without a client normalizing the paths first.
	for _, path := range []string{
		"/admin/", "/admin/index.html", "/admin%2Findex.html", "/%61dmin/index.html",
		"/admin%2findex.html", "/admin//index.html", "/admin/./index.html",
	} {
		rec := adminRequest(e, "GET", path, "", nil, false)
		if rec.Code == http.StatusOK || strings.Contains(rec.Body.String(), adminContent) {
			t.Errorf("unauthenticated static path exposed admin content: %s (%d)", path, rec.Code)
		}
	}
	for _, path := range []string{
		"/admin/../private.txt", "/admin/%2e%2e/private.txt",
		"/admin/%2E%2E/private.txt", "/admin/%2e%2e%2fprivate.txt",
		"/admin/%252e%252e/private.txt", "/admin/..%5cprivate.txt",
		"/admin/%5c..%5cprivate.txt", "/admin//../private.txt",
	} {
		for _, session := range []*http.Cookie{nil, cookie} {
			rec := adminRequest(e, "GET", path, "", session, false)
			if rec.Code == http.StatusOK || strings.Contains(rec.Body.String(), privateContent) {
				t.Errorf("static path escaped its root (logged in=%t): %s (%d)", session != nil, path, rec.Code)
			}
		}
	}
}

func TestAdminLoginLogoutAndSessionRotation(t *testing.T) {
	e, _ := adminTestServer(t)
	for _, body := range []string{
		`{"username":"admin","password":"incorrect"}`,
		`{"username":"other","password":"` + testAdminPassword + `"}`,
		`{}`, `null`,
	} {
		rec := adminRequest(e, "POST", "/api/login", body, nil, true)
		if rec.Code != http.StatusUnauthorized || len(rec.Result().Cookies()) != 0 {
			t.Fatalf("invalid credentials accepted: %d", rec.Code)
		}
	}
	cookie := adminLogin(t, e, nil)
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Secure || cookie.Path != "/" || cookie.MaxAge != 43200 {
		t.Fatalf("invalid HTTP session cookie attributes: %+v", cookie)
	}
	for _, path := range []string{"/admin/", "/admin/js/app.js", "/api/settings", "/api/devices"} {
		rec := adminRequest(e, "GET", path, "", cookie, false)
		if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("authenticated route %s failed: %d %s", path, rec.Code, rec.Body.String())
		}
	}
	if rec := adminRequest(e, "GET", "/login", "", cookie, false); rec.Header().Get("Location") != "/admin/" {
		t.Fatal("logged-in user not redirected to admin")
	}
	// Rotating a session invalidates its previous cookie at the server.
	newCookie := adminLogin(t, e, cookie)
	if cookie.Value == newCookie.Value || adminRequest(e, "GET", "/api/devices", "", cookie, false).Code != http.StatusUnauthorized {
		t.Fatal("old session survived login rotation")
	}
	if rec := adminRequest(e, "PUT", "/api/settings", `{"broadcast_address":"10.0.0.255"}`, newCookie, true); rec.Code != http.StatusOK {
		t.Fatalf("authenticated settings update failed: %d %s", rec.Code, rec.Body.String())
	}
	for _, route := range []struct{ method, path, body string }{
		{"POST", "/api/devices", `{"name":"Auth test PC","mac":"112233445566","secret":"` + strings.Repeat("A", 32) + `"}`},
		{"GET", "/api/devices/112233445566/code", ""},
		{"GET", "/api/devices/112233445566/qrcode", ""},
		{"GET", "/generate_code?mac=112233445566", ""},
		{"PUT", "/api/devices/112233445566", `{"name":"Updated PC"}`},
		{"DELETE", "/api/devices/112233445566", ""},
	} {
		if rec := adminRequest(e, route.method, route.path, route.body, newCookie, true); rec.Code != http.StatusOK {
			t.Fatalf("authenticated %s %s failed: %d %s", route.method, route.path, rec.Code, rec.Body.String())
		}
	}
	rec := adminRequest(e, "POST", "/api/logout", "", newCookie, true)
	if rec.Code != http.StatusNoContent || len(rec.Result().Cookies()) != 1 || rec.Result().Cookies()[0].MaxAge != -1 {
		t.Fatal("logout did not expire cookie")
	}
	if adminRequest(e, "GET", "/api/devices", "", newCookie, false).Code != http.StatusUnauthorized {
		t.Fatal("logged-out session is still valid")
	}
}

func TestAdminSessionExpiryTamperingAndRestart(t *testing.T) {
	e, auth := adminTestServer(t)
	now := time.Now()
	auth.now = func() time.Time { return now }
	cookie := adminLogin(t, e, nil)
	for _, value := range []string{"", strings.Repeat("0", 64), "forged-token"} {
		forged := &http.Cookie{Name: adminCookieName, Value: value}
		if adminRequest(e, "GET", "/api/devices", "", forged, false).Code != http.StatusUnauthorized {
			t.Fatal("forged cookie accepted")
		}
	}
	_, restartedAuth := adminTestServer(t)
	restarted := echo.New()
	RegisterRoutes(restarted, restartedAuth, "../static")
	if adminRequest(restarted, "GET", "/api/devices", "", cookie, false).Code != http.StatusUnauthorized {
		t.Fatal("session survived restart")
	}
	now = now.Add(adminSessionTTL)
	if adminRequest(e, "GET", "/api/devices", "", cookie, false).Code != http.StatusUnauthorized {
		t.Fatal("expired cookie accepted")
	}
}

func TestAdminCSRFAndLoginLimits(t *testing.T) {
	e, auth := adminTestServer(t)
	if adminRequest(e, "POST", "/api/login", `{}`, nil, false).Code != http.StatusForbidden {
		t.Fatal("login without CSRF header accepted")
	}
	cookie := adminLogin(t, e, nil)
	for _, route := range []struct{ method, path string }{
		{"PUT", "/api/settings"}, {"POST", "/api/devices"},
		{"PUT", "/api/devices/AABBCCDDEEFF"}, {"DELETE", "/api/devices/AABBCCDDEEFF"},
		{"POST", "/api/logout"},
	} {
		if adminRequest(e, route.method, route.path, `{}`, cookie, false).Code != http.StatusForbidden {
			t.Fatalf("CSRF protection missing: %s %s", route.method, route.path)
		}
	}
	// Oversized and malformed bodies cannot create sessions.
	for _, body := range []string{`{`, `{"password":"` + strings.Repeat("a", 5000) + `"}`} {
		if adminRequest(e, "POST", "/api/login", body, nil, true).Code != http.StatusBadRequest {
			t.Fatal("invalid login body accepted")
		}
	}
	for i := 0; i < 17; i++ {
		if adminRequest(e, "POST", "/api/login", `{}`, nil, true).Code != http.StatusUnauthorized {
			t.Fatal("login limited too early")
		}
	}
	rec := adminRequest(e, "POST", "/api/login", `{}`, nil, true)
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" {
		t.Fatal("login rate limit missing")
	}
	auth.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	adminLogin(t, e, nil)
}

func TestAdminHTTPSCookie(t *testing.T) {
	e, _ := adminTestServer(t)
	rec := adminRequest(e, "POST", "https://wol.example/api/login", `{"username":"admin","password":"`+testAdminPassword+`"}`, nil, true)
	if rec.Code != http.StatusNoContent || len(rec.Result().Cookies()) != 1 || !rec.Result().Cookies()[0].Secure {
		t.Fatal("HTTPS session cookie must be Secure")
	}
}

func TestAdminAuthRequiresConfiguredPassword(t *testing.T) {
	t.Setenv("ADMIN_USERNAME", "")
	t.Setenv("ADMIN_PASSWORD", "")
	if _, err := AdminAuthFromEnv(); err == nil {
		t.Fatal("missing admin password must fail closed")
	}
	t.Setenv("ADMIN_PASSWORD", testAdminPassword)
	auth, err := AdminAuthFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	RegisterRoutes(e, auth, "../static")
	adminLogin(t, e, nil) // Empty username configuration defaults to admin.
}
