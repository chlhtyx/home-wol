package common

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var Secret string

type GoogleAuth struct {
}

func NewGoogleAuth() *GoogleAuth {
	return &GoogleAuth{}
}

func (ga *GoogleAuth) un() int64 {
	return time.Now().UnixNano() / 1000 / 30
}

func (ga *GoogleAuth) hmacSha1(key, data []byte) []byte {
	h := hmac.New(sha1.New, key)
	total := len(data)
	if total > 0 {
		h.Write(data)
	}
	return h.Sum(nil)
}

func (ga *GoogleAuth) base32encode(src []byte) string {
	return base32.StdEncoding.EncodeToString(src)
}

func (ga *GoogleAuth) base32decode(s string) ([]byte, error) {
	return base32.StdEncoding.DecodeString(s)
}

func (ga *GoogleAuth) toBytes(value int64) []byte {
	var result []byte
	mask := int64(0xFF)
	shifts := [8]uint16{56, 48, 40, 32, 24, 16, 8, 0}
	for _, shift := range shifts {
		result = append(result, byte((value>>shift)&mask))
	}
	return result
}

func (ga *GoogleAuth) toUint32(bts []byte) uint32 {
	return (uint32(bts[0]) << 24) + (uint32(bts[1]) << 16) +
		(uint32(bts[2]) << 8) + uint32(bts[3])
}

func (ga *GoogleAuth) oneTimePassword(key []byte, data []byte) uint32 {
	hash := ga.hmacSha1(key, data)
	offset := hash[len(hash)-1] & 0x0F
	hashParts := hash[offset : offset+4]
	hashParts[0] = hashParts[0] & 0x7F
	number := ga.toUint32(hashParts)
	return number % 1000000
}

// GetSecret 获取秘钥
func (ga *GoogleAuth) GetSecret() string {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, ga.un())
	return strings.ToUpper(ga.base32encode(ga.hmacSha1(buf.Bytes(), nil)))
}

// GetCode 获取动态码
func (ga *GoogleAuth) GetCode(secret string) (string, error) {
	secretUpper := strings.ToUpper(secret)
	secretKey, err := ga.base32decode(secretUpper)
	if err != nil {
		return "", err
	}
	time := time.Now().Unix() / 30
	b := ga.toBytes(time)
	number := ga.oneTimePassword(secretKey, b)
	return fmt.Sprintf("%06d", number), nil
}

// GetQrcode 获取动态码二维码内容
func (ga *GoogleAuth) GetQrcode(user, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s?secret=%s", user, secret)
}

func (ga *GoogleAuth) GerQrcodePictureBase64(qrcode string) string {
	code, _ := QrcodePictureEncode(qrcode, 200, 200)
	return code
}

// GetQrcodeUrl 获取动态码二维码图片地址,这里是第三方二维码api
func (ga *GoogleAuth) GetQrcodeUrl(user, secret string) string {
	qrcode := ga.GetQrcode(user, secret)
	width := "200"
	height := "200"
	data := url.Values{}
	data.Set("data", qrcode)
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?%s&size=%sx%s&ecc=M", data.Encode(), width, height)
}

// VerifyCode 验证动态码
func (ga *GoogleAuth) VerifyCode(secret, code string) (bool, error) {
	_code, err := ga.GetCode(secret)
	fmt.Println(_code, code, err)
	if err != nil {
		return false, err
	}
	return _code == code, nil
}

func generate() {

	fmt.Println("-----------------开启二次认证----------------------")
	//user := "wol"
	//secret, code := initAuth(user)
	//fmt.Println(secret, code)
	secret := "55TVQXWNAS24WCSAHC2LXLOLMYDXC2QA"
	fmt.Println("-----------------信息校验----------------------")

	//"55TVQXWNAS24WCSAHC2LXLOLMYDXC2QA"

	// secret最好持久化保存在
	// 验证,动态码(从谷歌验证器获取或者freeotp获取)
	flag, err := NewGoogleAuth().VerifyCode(secret, "721890")
	if flag {
		fmt.Println("√")
	} else {
		fmt.Println("X", err)
	}
}

// 初始化
func initAuth(user string) (secret, code string) {
	// 秘钥
	secret = NewGoogleAuth().GetSecret()
	fmt.Println("Secret:", secret)

	// 动态码(每隔30s会动态生成一个6位数的数字)
	code, err := NewGoogleAuth().GetCode(secret)
	fmt.Println("Code:", code, err)

	// 用户名
	qrCode := NewGoogleAuth().GetQrcode(user, code)
	fmt.Println("Qrcode", qrCode)

	// 打印二维码地址
	qrCodeUrl := NewGoogleAuth().GetQrcodeUrl(user, secret)
	fmt.Println("QrcodeUrl", qrCodeUrl)

	return
}
