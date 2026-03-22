package main

import (
	"fmt"
	"home-wol/common"
	"log"
)

func main() {
	fmt.Println("========== Google Authenticator Secret 生成工具 ==========")
	fmt.Println()

	auth := common.NewGoogleAuth()

	// 生成 secret
	secret := auth.GetSecret()
	fmt.Println("生成的 Secret:", secret)
	fmt.Println()

	// 获取设备名称
	var deviceName string
	fmt.Print("请输入设备名称（如：我的电脑）: ")
	fmt.Scanln(&deviceName)

	if deviceName == "" {
		deviceName = "WOL设备"
	}

	// 生成当前验证码
	code, err := auth.GetCode(secret)
	if err != nil {
		log.Fatal("生成验证码失败:", err)
	}
	fmt.Println("当前验证码:", code)
	fmt.Println()

	// 生成二维码信息
	qrCode := auth.GetQrcode(deviceName, secret)
	fmt.Println("二维码内容:")
	fmt.Println(qrCode)
	fmt.Println()

	qrCodeUrl := auth.GetQrcodeUrl(deviceName, secret)
	fmt.Println("二维码图片 URL:")
	fmt.Println(qrCodeUrl)
	fmt.Println()

	// 生成配置示例
	fmt.Println("========== 配置文件示例 ==========")
	fmt.Println()
	fmt.Println(`{`)
	fmt.Printf(`  "devices": [` + "\n")
	fmt.Printf(`    {` + "\n")
	fmt.Printf(`      "mac": "AABBCCDDEEFF",` + "\n")
	fmt.Printf(`      "name": "%s",`+"\n", deviceName)
	fmt.Printf(`      "secret": "%s"`+"\n", secret)
	fmt.Printf(`    }` + "\n")
	fmt.Printf(`  ]` + "\n")
	fmt.Println(`}`)
	fmt.Println()

	fmt.Println("========== 使用说明 ==========")
	fmt.Println("1. 复制上面的 Secret 到 Google Authenticator 应用")
	fmt.Println("2. 或者扫描上面的二维码")
	fmt.Println("3. 将配置信息添加到 config.json 文件中")
	fmt.Println("4. 使用 /wol 接口测试唤醒功能")
}
