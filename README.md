# hikvisionOpenAPIGo
> 海康威视OpenAPI安全认证库 - Golang版本实现
# 官网

接口调用认证：[文档说明](https://open.hikvision.com/docs/7d0beeded66543999bff7bc2f91414d4)

其他语言版本：[下载链接](https://open.hikvision.com/download/5c67f1e2f05948198c909700?type=10)
# 快速使用
````
> go get github.com/zxbit2011/hikvisionOpenAPIGo
````
# 示例代码
````
func TestSDK(t *testing.T) {
	hk := hikvisionOpenAPIGo.HKConfig{
		Ip:      "127.0.0.1",
		Port:    443,
		AppKey:  "28057000",
		Secret:  "dZztQSS0000kLpURG000",
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
}
```` 
# 输出结果
````
TestSDK: sdk_test.go:26: OK {
                            	"code": "0",
                            	"msg": "success",
                            	"data": {
                            		"total": 1,
                            		"pageNo": 1,
                            		"pageSize": 100,
                            		"list": [{
                            			"altitude": 0.0,
                            			"cameraIndexCode": "01c1e8bd1b0d406a94e7cdf88a251f9b",
                            			"cameraName": "cameraTest",
                            			"cameraType": 0,
                            			"cameraTypeName": "枪机",
                            			"capabilitySet": "event_vss,io,vss,record,ptz,remote_vss,maintenance,status",
                            			"capabilitySetName": "视频事件能力,IO能力,视频能力,录像能力,云台能力,视频设备远程获取能力,设备维护能力,状态能力",
                            			"intelligentSet": null,
                            			"intelligentSetName": null,
                            			"channelNo": "1",
                            			"channelType": "analog",
                            			"channelTypeName": "模拟通道",
                            			"createTime": "2020-11-17T18:13:08.935+08:00",
                            			"encodeDevIndexCode": "0d983edda2694411ac15fa64bf29a8ca",
                            			"encodeDevResourceType": null,
                            			"encodeDevResourceTypeName": null,
                            			"gbIndexCode": null,
                            			"installLocation": "",
                            			"keyBoardCode": null,
                            			"latitude": "29.684556",
                            			"longitude": "106.703696",
                            			"pixel": null,
                            			"ptz": null,
                            			"ptzName": null,
                            			"ptzController": null,
                            			"ptzControllerName": null,
                            			"recordLocation": null,
                            			"recordLocationName": null,
                            			"regionIndexCode": "0ceebbf2-b7fd-4e5f-8c02-0d1725643444",
                            			"status": null,
                            			"statusName": null,
                            			"transType": 1,
                            			"transTypeName": "TCP",
                            			"treatyType": null,
                            			"treatyTypeName": null,
                            			"viewshed": null,
                            			"updateTime": "2020-11-24T16:16:41.368+08:00"
                            		}]
                            	}
                            }
````