package service

import (
	"encoding/hex"
	"fmt"
	"github.com/labstack/echo"
	"home-wol/common"
	"net"
	"net/http"
	"strings"
)

func Wol(c echo.Context) error {
	mac := c.QueryParam("mac")
	authCode := c.QueryParam("auth_code")

	if authCode == "" {
		return c.String(http.StatusInternalServerError, "验证失败")
	}

	fmt.Printf("mac:%s\n", mac)
	fmt.Printf("auth_code:%s\n", authCode)
	fmt.Printf("secret:%s\n", common.Secret)

	auth := common.NewGoogleAuth()
	code, err := auth.GetCode(common.Secret)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if authCode != code {
		return c.String(http.StatusInternalServerError, "验证失败")
	}

	if mac == "" {
		return c.String(http.StatusInternalServerError, "mac为空")
	}

	//处理mac地址
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")

	if len(mac) != 12 {
		return c.String(http.StatusInternalServerError, "mac格式无效")
	}

	data := "FFFFFFFFFFFF"

	for i := 0; i < 16; i++ {
		data = fmt.Sprintf("%s%s", data, strings.ToUpper(mac))
	}

	byte_data := []byte(data)

	// 将 byte 装换为 16进制的字符串
	hex_string_data := hex.EncodeToString(byte_data)
	// byte 转 16进制 的结果
	println(hex_string_data)

	/* ====== 分割线 ====== */

	// 将 16进制的字符串 转换 byte
	hex_data, _ := hex.DecodeString(hex_string_data)

	go common.SendWol(net.IPv4(192, 168, 2, 255), 7, hex_data)

	return c.String(http.StatusOK, "指令发送成功")
}
