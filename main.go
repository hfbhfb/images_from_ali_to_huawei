package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"

	"github.com/AliyunContainerService/image-syncer/cmd"
)

var (
	jsonTemplate = `{
	"auth":{
		"registry.REPLACEEDTEmpLate1.aliyuncs.com":{
			"username":"registryuserAli",
			"password":"registrypasswdAli"
		},
		"swr.REPLACEEDTEmpLate2.myhuaweicloud.com":{
			"username":"registryuserHuaWei",
			"password":"registrypasswdHuaWei"
		}
	},
	"images":{
		REPLACEEDTEmpLate3
	}
}`

	allReginAli = `
## 阿里云RegionID参数说明 区域列表: https://help.aliyun.com/document_detail/198107.html
华东1（杭州）	cn-hangzhou
华东2（上海）	cn-shanghai
华北1（青岛）	cn-qingdao
华北2（北京）	cn-beijing
华南1（深圳）	cn-shenzhen
新加坡	ap-southeast-1
马来西亚（吉隆坡）	ap-southeast-3
印度尼西亚（雅加达）	ap-southeast-5
日本（东京）	ap-northeast-1
英国（伦敦）	eu-west-1
`

	allReginHw = `
## 华为云区域 地区和终端节点 区域列表: https://developer.huaweicloud.com/endpoint?SWR
非洲-约翰内斯堡	af-south-1	swr-api.af-south-1.myhuaweicloud.com	HTTPS
华北-北京四	cn-north-4	swr-api.cn-north-4.myhuaweicloud.com	HTTPS
华北-北京一	cn-north-1	swr-api.cn-north-1.myhuaweicloud.com	HTTPS
华东-上海二	cn-east-2	swr-api.cn-east-2.myhuaweicloud.com	HTTPS
华东-上海一	cn-east-3	swr-api.cn-east-3.myhuaweicloud.com	HTTPS
华南-广州	cn-south-1	swr-api.cn-south-1.myhuaweicloud.com	HTTPS
拉美-圣地亚哥	la-south-2	swr-api.la-south-2.myhuaweicloud.com	HTTPS
欧洲-巴黎	eu-west-0	swr-api.eu-west-0.myhuaweicloud.com	HTTPS
西南-贵阳一	cn-southwest-2	swr-api.cn-southwest-2.myhuaweicloud.com	HTTPS
亚太-曼谷	ap-southeast-2	swr-api.ap-southeast-2.myhuaweicloud.com	HTTPS
亚太-新加坡	ap-southeast-3	swr-api.ap-southeast-3.myhuaweicloud.com	HTTPS
中国-香港	ap-southeast-1	swr-api.ap-southeast-1.myhuaweicloud.com	HTTPS
`

	useAge = `用法:
export AK="xxxx" #阿里云ak
export SK="xxxx" #阿里云sk
export RegionAli="cn-hangzhou" # 阿里云区域 https://help.aliyun.com/document_detail/198107.html
export UserAli="xxx" #阿里云镜像用户
export PasswdAli="xxx" #阿里云镜像密码
export RegionHw="cn-south-1" # 华为云区域  https://developer.huaweicloud.com/endpoint?SWR
export UserHw="xxx" #华为云镜像用户
export PasswdHw="xxx" #华为云镜像密码
export RunFlag="1"  # 开关,当此值为1时才真正的执行镜像同步
export OnlyRun="false" #开关,如果打开则跳过config.json的生成过程
./images_from_ali_to_huawei --config=./config.json

`

	_ = `
原理(在官方镜像工具套了一层): 
1. 列举所有阿里的镜像,输出config.json文件
2. 需要预先手动通过华为云管理后台[web]创建同名组织(在阿里云是命名空间)
3. 使用 "阿里云官方" 镜像工具(https://github.com/AliyunContainerService/image-syncer) 进行镜像同步`

// ReginHanZhou = "cn-hangzhou"
)

type Items struct {
	ReginId       string `json:"regionId"`
	RepoType      string `json:"repoType"`
	RepoNamespace string `json:"repoNamespace"`
	RepoName      string `json:"repoName"`
}
type RespData struct {
	Repos []Items
	Total int
}
type RespStrucL struct {
	Data RespData
}

func prepareEnv() {

	fmt.Println(`步骤一:列举所有阿里的镜像,输出config.json文件`)
	AK := os.Getenv("AK")
	SK := os.Getenv("SK")
	RegionAli := os.Getenv("RegionAli")
	RegionHw := os.Getenv("RegionHw")

	UserAli := os.Getenv("UserAli")
	PasswdAli := os.Getenv("PasswdAli")
	UserHw := os.Getenv("UserHw")
	PasswdHw := os.Getenv("PasswdHw")
	if UserAli == "" || PasswdAli == "" || UserHw == "" || PasswdHw == "" {
		fmt.Println(useAge)
		fmt.Println("请先在环境变量中设置 docker帐号(阿里,华为)信息")
		os.Exit(1)
	}

	if AK == "" || SK == "" {
		fmt.Println(useAge)
		fmt.Println("请先在环境变量中设置AK,SK")
		os.Exit(1)
	}
	if RegionAli == "" {
		fmt.Println(useAge)
		fmt.Println(allReginAli)
		fmt.Println("请先在环境变量中设置阿里云区域 RegionAli")
		os.Exit(1)
	}
	if RegionHw == "" {
		fmt.Println(useAge)
		fmt.Println(allReginHw)
		fmt.Println("请先在环境变量中设置华为云区域 RegionHw")
		os.Exit(1)
	}

	fmt.Println("阿里区域:", RegionAli)
	fmt.Println("华为区域:", RegionHw)
	client, err := cr.NewClientWithAccessKey(RegionAli, AK, SK)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var allItems []Items
	i := 1
	for {
		a := cr.CreateGetRepoListRequest()
		a.PageSize = requests.Integer("99") // 接口中不能超过99
		a.Page = requests.Integer(fmt.Sprintf("%v", i))
		resp, err := client.GetRepoList(a)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		var res RespStrucL
		bs := []byte(resp.GetHttpContentString())
		// fmt.Println(string(bs))
		err = json.Unmarshal(bs, &res)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// fmt.Println(res)
		for _, v := range res.Data.Repos {
			var t Items
			t.ReginId = v.ReginId
			t.RepoType = v.RepoType
			t.RepoNamespace = v.RepoNamespace
			t.RepoName = v.RepoName
			allItems = append(allItems, t)
		}

		if len(res.Data.Repos) != 99 {
			break
		}
		i++
	}

	allstrlv := ""
	for _, v := range allItems {
		lv := fmt.Sprintf("%v%v%v%v%v%v%v%v%v%v%v%v%v", "\"registry.", RegionAli, ".aliyuncs.com/", v.RepoNamespace, "/", v.RepoName, "\":\"swr.", RegionHw, ".myhuaweicloud.com/", v.RepoNamespace, "/", v.RepoName, "\"")
		if allstrlv != "" {
			allstrlv += ",\n"
		}
		allstrlv += lv
		// fmt.Println(lv)
		// fmt.Println("registry.", RegionAli, ".aliyuncs.com/", v.RepoNamespace, "/", v.RepoName)
		// fmt.Println("swr.", RegionHw, ".myhuaweicloud.com/", v.RepoNamespace, "/", v.RepoName)
	}

	// "registry.cn-hangzhou.aliyuncs.com/hfbhfb1/repostry_a":"swr.cn-south-1.myhuaweicloud.com/hfbhw1/aaa"
	jsonTemplate = strings.ReplaceAll(jsonTemplate, "REPLACEEDTEmpLate1", RegionAli)
	jsonTemplate = strings.ReplaceAll(jsonTemplate, "REPLACEEDTEmpLate2", RegionHw)

	jsonTemplate = strings.ReplaceAll(jsonTemplate, "registryuserAli", UserAli)
	jsonTemplate = strings.ReplaceAll(jsonTemplate, "registrypasswdAli", PasswdAli)
	jsonTemplate = strings.ReplaceAll(jsonTemplate, "registryuserHuaWei", UserHw)
	jsonTemplate = strings.ReplaceAll(jsonTemplate, "registrypasswdHuaWei", PasswdHw)

	jsonTemplate = strings.ReplaceAll(jsonTemplate, "REPLACEEDTEmpLate3", allstrlv)

	// fmt.Println(len(allItems))
	// fmt.Println(len(allItems))
	// fmt.Println(len(allItems))
	// fmt.Println(allItems)
	// fmt.Println(allstrlv)
	fmt.Println("生成config.json文件 如下:")
	fmt.Println(jsonTemplate)
	ioutil.WriteFile("config.json", []byte(jsonTemplate), 0666)

	// fmt.Println(resp.GetHttpContentString())
}

func main() {
	for _, v := range os.Args {
		if v == "-h" || v == "--help" || v == "help" {
			fmt.Println(useAge)
			return
		}
	}

	OnlyRun := os.Getenv("OnlyRun")
	if !(OnlyRun == "1" || OnlyRun == "true") {
		prepareEnv()
	}

	RunFlag := os.Getenv("RunFlag")
	if RunFlag == "1" || RunFlag == "true" {
		fmt.Println(`步骤二:执行镜像同步
阿里云容器服务工具: https://github.com/AliyunContainerService/image-syncer`)
		cmd.Execute()
	}
}
