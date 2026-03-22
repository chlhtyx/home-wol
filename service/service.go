package service

import (
	"encoding/hex"
	"fmt"
	"home-wol/common"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

var BroadcastIP net.IP

func Wol(c echo.Context) error {
	mac := c.QueryParam("mac")
	authCode := c.QueryParam("auth_code")
	//userName := c.QueryParam("user_name")

	if authCode == "" {
		return c.String(http.StatusInternalServerError, "验证码不能为空")
	}

	log.Println("mac:", mac)
	log.Println("auth_code:", authCode)

	// 根据 MAC 地址获取对应的 secret
	secret, err := common.GetSecretByMAC(mac)
	if err != nil {
		log.Println("获取secret失败:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	log.Println("secret:", secret)

	auth := common.NewGoogleAuth()

	code, err := auth.GetCode(secret)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if authCode != code {
		log.Println("验证失败:")
		return c.String(http.StatusInternalServerError, "验证失败")
	}

	if mac == "" {
		return c.String(http.StatusInternalServerError, "mac为空")
	}

	//处理mac地址
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	mac = strings.ToUpper(mac)

	hexStr := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F"}

	for _, str := range mac {
		flag := ArrayIndex(string(str), hexStr)
		if flag == -1 {
			return c.String(http.StatusInternalServerError, "mac格式无效")
		}
	}

	if len(mac) != 12 {
		return c.String(http.StatusInternalServerError, "mac格式无效")
	}

	data := "FFFFFFFFFFFF"

	for i := 0; i < 16; i++ {
		data = fmt.Sprintf("%s%s", data, mac)
	}

	// 将 16进制的字符串 转换 byte
	hexData, _ := hex.DecodeString(data)

	go common.SendWol(BroadcastIP, 9, hexData)
	log.Println("指令发送成功:")
	return c.String(http.StatusOK, "指令发送成功")
}

// GenerateCode 生成动态校验码接口
func GenerateCode(c echo.Context) error {
	mac := c.QueryParam("mac")

	if mac == "" {
		return c.String(http.StatusInternalServerError, "mac不能为空")
	}

	log.Println("生成校验码请求 - mac:", mac)

	// 根据 MAC 地址获取对应的 secret
	secret, err := common.GetSecretByMAC(mac)
	if err != nil {
		log.Println("获取secret失败:", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	auth := common.NewGoogleAuth()
	code, err := auth.GetCode(secret)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	log.Println("生成校验码成功 - mac:", mac, "code:", code)
	return c.JSON(http.StatusOK, map[string]string{
		"mac":    mac,
		"code":   code,
		"secret": secret,
	})
}

func auth(authCode string) error {
	log.Println(authCode)

	return nil
}

func ArrayIndex(v string, a []string) int {
	for ai, av := range a {
		if v == av {
			return ai
		}
	}

	return -1
}
