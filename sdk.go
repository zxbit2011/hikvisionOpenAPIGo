package hikvisionOpenAPIGo

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

// HKConfig 海康OpenAPI配置参数
type HKConfig struct {
	Ip      string //平台ip
	Port    int    //平台端口
	AppKey  string //平台APPKey
	Secret  string //平台APPSecret
	IsHttps bool   //是否使用HTTPS协议
}

// @title		HTTP Post请求
// @url			HTTP接口Url		string				 HTTP接口Url，不带协议和端口，如/artemis/api/resource/v1/org/advance/orgList
// @body		请求参数			map[string]string
// @return		请求结果			参数类型
func (hk HKConfig) HttpPost(url string, body map[string]string, timeout int) (result []byte, err error) {
	var header = make(map[string]string)
	bodyJson := MustJsonString(body)
	hk.initRequest(header, url, bodyJson, true)
	var sb []string
	if hk.IsHttps {
		sb = append(sb, "https://")
	} else {
		sb = append(sb, "http://")
	}
	sb = append(sb, hk.Ip)
	sb = append(sb, ":")
	sb = append(sb, fmt.Sprintf("%d", hk.Port))
	sb = append(sb, url)

	client := &http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	client.Timeout = time.Duration(timeout) * time.Second
	if hk.IsHttps {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}
	req, err := http.NewRequest("POST", strings.Join(sb, ""), bytes.NewReader([]byte(bodyJson)))
	if err != nil {
		return
	}

	req.Header.Set("Accept", header["Accept"])
	req.Header.Set("Content-Type", header["Content-Type"])
	for k, v := range header {
		if strings.Contains(k, "x-ca-") {
			req.Header.Set(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		result, err = ioutil.ReadAll(resp.Body)
	} else if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		reqUrl := resp.Header.Get("Location")
		panic(fmt.Errorf("HttpPost Response StatusCode：%d，Location：%s", resp.StatusCode, reqUrl))
	} else {
		err = fmt.Errorf("HttpPost Response StatusCode：%d", resp.StatusCode)
	}
	return
}

// initRequest 初始化请求头
func (hk HKConfig) initRequest(header map[string]string, url, body string, isPost bool) {
	header["Accept"] = "application/json"
	header["Content-Type"] = "application/json"
	if isPost {
		header["content-md5"] = computeContentMd5(body)
	}
	header["x-ca-timestamp"] = MustString(time.Now().UnixNano() / 1e6)
	uid, _ := uuid.NewV4()
	header["x-ca-nonce"] = uid.String()
	header["x-ca-key"] = hk.AppKey

	var strToSign string
	if isPost {
		strToSign = buildSignString(header, url, "POST")
	} else {
		strToSign = buildSignString(header, url, "GET")
	}
	signedStr, err := computeForHMACSHA256(strToSign, hk.Secret)
	if err != nil {
		println(err.Error())
		return
	}
	header["x-ca-signature"] = signedStr
}

// computeContentMd5 计算content-md5
func computeContentMd5(body string) string {
	return base64.StdEncoding.EncodeToString([]byte( Md5(body)))
}

// computeForHMACSHA256 计算HMACSHA265
func computeForHMACSHA256(str, secret string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write([]byte(str))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// buildSignString 计算签名字符串
func buildSignString(header map[string]string, url, method string) string {
	var sb []string
	sb = append(sb, strings.ToUpper(method))
	sb = append(sb, "\n")

	if header != nil {
		if _, ok := header["Accept"]; ok {
			sb = append(sb, header["Accept"])
			sb = append(sb, "\n")
		}
		if _, ok := header["Content-MD5"]; ok {
			sb = append(sb, header["Content-MD5"])
			sb = append(sb, "\n")
		}
		if _, ok := header["Content-Type"]; ok {
			sb = append(sb, header["Content-Type"])
			sb = append(sb, "\n")
		}
		if _, ok := header["Date"]; ok {
			sb = append(sb, header["Date"])
			sb = append(sb, "\n")
		}
	}
	sb = append(sb, buildSignHeader(header))
	sb = append(sb, url)
	return strings.Join(sb, "")
}

// buildSignHeader 计算签名头
func buildSignHeader(header map[string]string) string {
	var sortedDicHeader map[string]string
	sortedDicHeader = header

	var sslice []string
	for key, _ := range sortedDicHeader {
		sslice = append(sslice, key)
	}
	sort.Strings(sslice)

	var sbSignHeader []string
	var sb []string
	//在将key输出
	for _, k := range sslice {
		if strings.Contains(strings.ReplaceAll(k, " ", ""), "x-ca-") {
			sb = append(sb, k+":")
			if sortedDicHeader[k] != "" {
				sb = append(sb, sortedDicHeader[k])
			}
			sb = append(sb, "\n")
			if len(sbSignHeader) > 0 {
				sbSignHeader = append(sbSignHeader, ",")
			}
			sbSignHeader = append(sbSignHeader, k)
		}
	}

	header["x-ca-signature-headers"] = strings.Join(sbSignHeader, "")
	return strings.Join(sb, "")
}

func MustJson(i interface{}) []byte {
	if d, err := json.Marshal(i); err == nil {
		return d
	} else {
		panic(err)
	}
}

func MustJsonString(i interface{}) string {
	return string(MustJson(i))
}

func MustString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
