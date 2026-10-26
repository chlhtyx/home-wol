package service

import (
	"encoding/json"
	"home-wol/common"
	"net/http"
	"os"
	"strings"
	"testing"
)

// Reproduce a new router: start with no devices, then add one through the API.
func TestDeviceAuthRefreshesAfterSaving(t *testing.T) {
	e, path := settingsTestAPI(t, "192.168.1.255")
	if err := os.WriteFile(path, []byte(`{"devices":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := common.GetConfig(path); err != nil {
		t.Fatal(err)
	}
	e.GET("/api/devices/:mac/code", GetDeviceCode)
	e.GET("/generate_code", GenerateCode)

	firstSecret := strings.Repeat("A", 32)
	secondSecret := strings.Repeat("B", 32)
	device, err := json.Marshal(common.DeviceConfig{MAC: "112233445566", Name: "Router PC", Secret: firstSecret})
	if err != nil {
		t.Fatal(err)
	}
	settingsRequest(t, e, http.MethodPost, "/api/devices", string(device), http.StatusOK)
	settingsRequest(t, e, http.MethodGet, "/api/devices/112233445566/code", "", http.StatusOK)
	settingsRequest(t, e, http.MethodGet, "/generate_code?mac=112233445566", "", http.StatusOK)

	device, err = json.Marshal(common.DeviceConfig{Name: "Router PC", Secret: secondSecret})
	if err != nil {
		t.Fatal(err)
	}
	settingsRequest(t, e, http.MethodPut, "/api/devices/112233445566", string(device), http.StatusOK)
	rec := settingsRequest(t, e, http.MethodGet, "/api/devices/112233445566/code", "", http.StatusOK)
	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["secret"] != secondSecret {
		t.Fatal("verification still uses the old device secret")
	}
	if secret, err := common.GetSecretByMAC("112233445566"); err != nil || secret != result["secret"] {
		t.Fatalf("WOL authentication cache is stale: %v", err)
	}
	settingsRequest(t, e, http.MethodDelete, "/api/devices/112233445566", "", http.StatusOK)
	settingsRequest(t, e, http.MethodGet, "/api/devices/112233445566/code", "", http.StatusNotFound)
	if _, err := common.GetSecretByMAC("112233445566"); err != common.ErrDeviceNotFound {
		t.Fatal("deleted device remains available for WOL authentication")
	}
}
