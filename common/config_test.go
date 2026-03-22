package common

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	configContent := `{
		"devices": [
			{
				"mac": "AABBCCDDEEFF",
				"name": "测试设备1",
				"secret": "JBSWY3DPEHPK3PXP"
			},
			{
				"mac": "aa:bb:cc:dd:ee:ff",
				"name": "测试设备2",
				"secret": "KRSXG5DSNFXAWZL"
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	// 测试加载配置
	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if len(cfg.Devices) != 2 {
		t.Errorf("期望2个设备，实际得到 %d 个", len(cfg.Devices))
	}

	// 测试 MAC 地址标准化
	if cfg.Devices[0].MAC != "AABBCCDDEEFF" {
		t.Errorf("期望 MAC 为 AABBCCDDEEFF，实际得到 %s", cfg.Devices[0].MAC)
	}

	if cfg.Devices[1].MAC != "AABBCCDDEEFF" {
		t.Errorf("期望 MAC 为 AABBCCDDEEFF（标准化后），实际得到 %s", cfg.Devices[1].MAC)
	}
}

func TestGetSecretByMAC(t *testing.T) {
	// 创建临时配置文件
	configContent := `{
		"devices": [
			{
				"mac": "AABBCCDDEEFF",
				"name": "测试设备",
				"secret": "JBSWY3DPEHPK3PXP"
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	// 加载配置
	config, _ = LoadConfig(tmpFile.Name())

	tests := []struct {
		name    string
		mac     string
		want    string
		wantErr error
	}{
		{
			name:    "标准 MAC 地址",
			mac:     "AABBCCDDEEFF",
			want:    "JBSWY3DPEHPK3PXP",
			wantErr: nil,
		},
		{
			name:    "带分隔符的 MAC 地址",
			mac:     "AA:BB:CC:DD:EE:FF",
			want:    "JBSWY3DPEHPK3PXP",
			wantErr: nil,
		},
		{
			name:    "小写 MAC 地址",
			mac:     "aabbccddeeff",
			want:    "JBSWY3DPEHPK3PXP",
			wantErr: nil,
		},
		{
			name:    "不存在的 MAC 地址",
			mac:     "112233445566",
			want:    "",
			wantErr: ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSecretByMAC(tt.mac)
			if err != tt.wantErr {
				t.Errorf("GetSecretByMAC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSecretByMAC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		name string
		mac  string
		want string
	}{
		{
			name: "标准格式",
			mac:  "AABBCCDDEEFF",
			want: "AABBCCDDEEFF",
		},
		{
			name: "带冒号",
			mac:  "AA:BB:CC:DD:EE:FF",
			want: "AABBCCDDEEFF",
		},
		{
			name: "带连字符",
			mac:  "AA-BB-CC-DD-EE-FF",
			want: "AABBCCDDEEFF",
		},
		{
			name: "小写",
			mac:  "aabbccddeeff",
			want: "AABBCCDDEEFF",
		},
		{
			name: "混合大小写带分隔符",
			mac:  "Aa:Bb:Cc:Dd:Ee:Ff",
			want: "AABBCCDDEEFF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeMAC(tt.mac); got != tt.want {
				t.Errorf("normalizeMAC() = %v, want %v", got, tt.want)
			}
		})
	}
}
