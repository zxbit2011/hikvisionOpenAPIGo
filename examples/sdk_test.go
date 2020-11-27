package examples

import (
	"github.com/zxbit2011/hikvisionOpenAPIGo"
	"testing"
)

func TestSDK(t *testing.T) {
	hk := hikvisionOpenAPIGo.HKConfig{
		Ip:      "172.17.207.240",
		Port:    443,
		AppKey:  "28057383",
		Secret:  "dZztQSSUAF4kLpURGQMa",
		IsHttps: true,
	}

	body := map[string]string{
		"pageNo":   "1",
		"pageSize": "100",
	}
	result, err := hk.HttpPost("/artemis/api/resource/v1/cameras", body, 15)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log("OK", string(result))

	/*body := map[string]string{
		"cameraIndexCode": "71c1e8bd1b0d406a94e7cdf88a251f9b",
		"protocol":        "rtmp",
	}
	result, err := hk.Post("/artemis/api/video/v2/cameras/previewURLs", body, 15)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log("OK", string(result))*/
}
