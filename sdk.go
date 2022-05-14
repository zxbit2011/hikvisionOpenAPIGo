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
	"github.com/gofrs/uuid"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
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

// 返回结果
type Result struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 返回值data
type Data struct {
	Total    int                      `json:"total"`
	PageSize int                      `json:"pageSize"`
	PageNo   int                      `json:"pageNo"`
	List     []map[string]interface{} `json:"list"`
}

// @title		HTTP Post请求
// @url			HTTP接口Url		string				 HTTP接口Url，不带协议和端口，如/artemis/api/resource/v1/org/advance/orgList
// @body		请求参数			map[string]string
// @return		请求结果			参数类型
func (hk HKConfig) HttpPost(url string, body map[string]string, timeout int) (result Result, err error) {
	var header = make(map[string]string)
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return result, err
	}
	err = hk.initRequest(header, url, string(bodyJson), true)
	if err != nil {
		return Result{}, err
	}
	var sb []string
	if hk.IsHttps {
		sb = append(sb, "https://")
	} else {
		sb = append(sb, "http://")
	}
	sb = append(sb, fmt.Sprintf("%s:%d", hk.Ip, hk.Port))
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
	req, err := http.NewRequest("POST", strings.Join(sb, ""), bytes.NewReader(bodyJson))
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
		var resBody []byte
		resBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		err = json.Unmarshal(resBody, &result)
	} else if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		reqUrl := resp.Header.Get("Location")
		err = fmt.Errorf("HttpPost Response StatusCode：%d，Location：%s", resp.StatusCode, reqUrl)
	} else {
		err = fmt.Errorf("HttpPost Response StatusCode：%d", resp.StatusCode)
	}
	return
}

// initRequest 初始化请求头
func (hk HKConfig) initRequest(header map[string]string, url, body string, isPost bool) error {
	header["Accept"] = "application/json"
	header["Content-Type"] = "application/json"
	if isPost {
		var err error
		header["content-md5"], err = computeContentMd5(body)
		if err != nil {
			return err
		}
	}
	header["x-ca-timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	uid, err := uuid.NewV4()
	if err != nil {
		return err
	}
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
		return err
	}
	header["x-ca-signature"] = signedStr
	return nil
}

// computeContentMd5 计算content-md5
func computeContentMd5(body string) (string, error) {
	h := md5.New()
	_, err := h.Write([]byte(body))
	if err != nil {
		return "", err
	}
	md5Str := hex.EncodeToString(h.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(md5Str)), nil
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
