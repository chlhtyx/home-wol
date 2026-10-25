package service

import (
	"encoding/json"
	"home-wol/common"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func settingsTestAPI(t *testing.T, address string) (*echo.Echo, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := common.Config{
		BroadcastAddress: address,
		Devices: []common.DeviceConfig{{
			MAC: "AABBCCDDEEFF", Name: "Existing device", Secret: "JBSWY3DPEHPK3PXP",
		}},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	previousPath := configPath
	InitAPI(path)
	t.Cleanup(func() { InitAPI(previousPath) })
	e := echo.New()
	e.GET("/api/settings", GetSettings)
	e.PUT("/api/settings", UpdateSettings)
	e.POST("/api/devices", AddDevice)
	e.PUT("/api/devices/:mac", UpdateDevice)
	e.DELETE("/api/devices/:mac", DeleteDevice)
	return e, path
}

func settingsRequest(t *testing.T, e *echo.Echo, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != status {
		t.Fatalf("%s %s: got %d, want %d: %s", method, path, rec.Code, status, rec.Body.String())
	}
	return rec
}

func TestSettingsAddressPrecedence(t *testing.T) {
	for _, test := range []struct {
		name, saved, legacy, want string
	}{
		{"default for old config", "", "", defaultBroadcastAddress},
		{"legacy environment", "", "192.168.2.255", "192.168.2.255"},
		{"saved setting overrides environment", "10.0.0.127", "192.168.2.255", "10.0.0.127"},
		{"saved setting overrides invalid environment", "10.0.0.255", "invalid", "10.0.0.255"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("BROADCAST_ADDRESS", test.legacy)
			e, _ := settingsTestAPI(t, test.saved)
			rec := settingsRequest(t, e, http.MethodGet, "/api/settings", "", http.StatusOK)
			var result Settings
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.BroadcastAddress != test.want {
				t.Fatalf("got %q, want %q", result.BroadcastAddress, test.want)
			}
			// Settings responses must not expose device secrets.
			if strings.Contains(rec.Body.String(), "secret") {
				t.Fatal("settings response contains device secrets")
			}
		})
	}
}

func TestSettingsPersistAndSurviveDeviceChanges(t *testing.T) {
	t.Setenv("BROADCAST_ADDRESS", "192.168.2.255")
	e, path := settingsTestAPI(t, "")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	settingsRequest(t, e, http.MethodPut, "/api/settings", `{"broadcast_address":" 10.20.30.127 "}`, http.StatusOK)

	// The actual WOL call site reads this function on every request.
	ip, err := GetBroadcastIP()
	if err != nil || ip.String() != "10.20.30.127" {
		t.Fatalf("WOL target did not change immediately: %v, %v", ip, err)
	}
	backup, err := os.ReadFile(filepath.Join(filepath.Dir(path), "config.backup.json"))
	if err != nil || string(backup) != string(before) {
		t.Fatalf("original configuration was not backed up: %v", err)
	}
	InitAPI(path) // Simulate restarting the service using the persisted file.
	settingsRequest(t, e, http.MethodGet, "/api/settings", "", http.StatusOK)
	cfg, err := common.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Devices) != 1 || cfg.Devices[0].Name != "Existing device" || cfg.Devices[0].Secret != "JBSWY3DPEHPK3PXP" {
		t.Fatal("saving settings changed existing devices")
	}
	settingsRequest(t, e, http.MethodPost, "/api/devices", `{"mac":"112233445566","name":"New device","secret":"JBSWY3DPEHPK3PXP"}`, http.StatusOK)
	settingsRequest(t, e, http.MethodPut, "/api/devices/112233445566", `{"name":"Renamed device"}`, http.StatusOK)
	settingsRequest(t, e, http.MethodDelete, "/api/devices/112233445566", "", http.StatusOK)
	cfg, err = common.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BroadcastAddress != "10.20.30.127" || len(cfg.Devices) != 1 {
		t.Fatalf("device changes lost network settings or devices: %+v", cfg)
	}
}

func TestInvalidSettingsDoNotChangeConfig(t *testing.T) {
	e, path := settingsTestAPI(t, "192.168.1.255")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{}`, `null`, `{"broadcast_address":42}`, `{"broadcast_address":""}`,
		`{"broadcast_address":"192.168.1.999"}`, `{"broadcast_address":"192.168.01.255"}`,
		`{"broadcast_address":"::1"}`, `{"broadcast_address":"::ffff:192.168.1.255"}`,
		`{"broadcast_address":"0.0.0.0"}`, `{"broadcast_address":"127.0.0.1"}`,
		`{"broadcast_address":"224.0.0.1"}`, `{"broadcast_address":"example.com"}`, `{`,
	} {
		settingsRequest(t, e, http.MethodPut, "/api/settings", body, http.StatusBadRequest)
		after, err := os.ReadFile(path)
		if err != nil || string(after) != string(before) {
			t.Fatalf("invalid input changed config: %s, %v", body, err)
		}
	}
}

func TestSettingsSaveFailureKeepsCurrentAddress(t *testing.T) {
	e, path := settingsTestAPI(t, "192.168.1.255")
	// A directory at the backup path causes the real persistence path to fail.
	if err := os.Mkdir(filepath.Join(filepath.Dir(path), "config.backup.json"), 0700); err != nil {
		t.Fatal(err)
	}
	settingsRequest(t, e, http.MethodPut, "/api/settings", `{"broadcast_address":"10.0.0.255"}`, http.StatusInternalServerError)
	ip, err := GetBroadcastIP()
	if err != nil || ip.String() != "192.168.1.255" {
		t.Fatalf("failed save changed WOL target: %v, %v", ip, err)
	}
}

func TestSettingsConfigReadFailure(t *testing.T) {
	e, path := settingsTestAPI(t, "192.168.1.255")
	if err := os.WriteFile(path, []byte("invalid JSON"), 0600); err != nil {
		t.Fatal(err)
	}
	settingsRequest(t, e, http.MethodGet, "/api/settings", "", http.StatusInternalServerError)
	settingsRequest(t, e, http.MethodPut, "/api/settings", `{"broadcast_address":"10.0.0.255"}`, http.StatusInternalServerError)
}
