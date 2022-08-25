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

	// 将 16进制的字符串 转换 byte
	hexData, _ := hex.DecodeString(data)

	go common.SendWol(net.IPv4(192, 168, 2, 255), 9, hexData)

	return c.String(http.StatusOK, "指令发送成功")
}
