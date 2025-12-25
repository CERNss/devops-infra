package mirrormap

import "strings"

type Kind string

const (
	KindSystem         Kind = "system"
	KindDockerCE       Kind = "docker-ce"
	KindDockerRegistry Kind = "docker-registry"
)

var aliasToKey = map[string]string{
	"阿里":           "aliyun",
	"阿里云":          "aliyun",
	"ali":          "aliyun",
	"aliyun":       "aliyun",
	"腾讯":           "tencent",
	"腾讯云":          "tencent",
	"tencent":      "tencent",
	"tencentyun":   "tencent",
	"华为":           "huawei",
	"华为云":          "huawei",
	"huawei":       "huawei",
	"huaweicloud":  "huawei",
	"天翼云":          "ctyun",
	"ctyun":        "ctyun",
	"netease":      "netease",
	"163":          "netease",
	"网易":           "netease",
	"volc":         "volc",
	"volces":       "volc",
	"火山":           "volc",
	"清华":           "tsinghua",
	"清华大学":         "tsinghua",
	"tuna":         "tsinghua",
	"tsinghua":     "tsinghua",
	"北大":           "pku",
	"北京大学":         "pku",
	"pku":          "pku",
	"浙大":           "zju",
	"浙江大学":         "zju",
	"zju":          "zju",
	"南大":           "nju",
	"南京大学":         "nju",
	"nju":          "nju",
	"交大":           "sjtu",
	"上海交通大学":       "sjtu",
	"sjtu":         "sjtu",
	"中科大":          "ustc",
	"中国科学技术大学":     "ustc",
	"ustc":         "ustc",
	"中科院":          "iscas",
	"iscas":        "iscas",
	"azure":        "azure",
	"1ms":          "1ms",
	"dockerproxy":  "dockerproxy",
	"daocloud":     "daocloud",
	"1panel":       "1panel",
	"dockerhub":    "dockerhub",
	"docker":       "docker",
}

var systemDomains = map[string]string{
	"aliyun":   "mirrors.aliyun.com",
	"tencent":  "mirrors.tencent.com",
	"huawei":   "mirrors.huaweicloud.com",
	"ctyun":    "mirrors.ctyun.cn",
	"netease":  "mirrors.163.com",
	"volc":     "mirrors.volces.com",
	"tsinghua": "mirrors.tuna.tsinghua.edu.cn",
	"pku":      "mirrors.pku.edu.cn",
	"zju":      "mirrors.zju.edu.cn",
	"nju":      "mirrors.nju.edu.cn",
	"sjtu":     "mirror.sjtu.edu.cn",
	"ustc":     "mirrors.ustc.edu.cn",
	"iscas":    "mirror.iscas.ac.cn",
}

var dockerCEDomains = map[string]string{
	"aliyun":   "mirrors.aliyun.com/docker-ce",
	"tencent":  "mirrors.tencent.com/docker-ce",
	"huawei":   "mirrors.huaweicloud.com/docker-ce",
	"netease":  "mirrors.163.com/docker-ce",
	"volc":     "mirrors.volces.com/docker",
	"azure":    "mirror.azure.cn/docker-ce",
	"tsinghua": "mirrors.tuna.tsinghua.edu.cn/docker-ce",
	"pku":      "mirrors.pku.edu.cn/docker-ce",
	"zju":      "mirrors.zju.edu.cn/docker-ce",
	"nju":      "mirrors.nju.edu.cn/docker-ce",
	"sjtu":     "mirror.sjtu.edu.cn/docker-ce",
	"ustc":     "mirrors.ustc.edu.cn/docker-ce",
	"iscas":    "mirror.iscas.ac.cn/docker-ce",
	"docker":   "download.docker.com",
}

var dockerRegistryDomains = map[string]string{
	"1ms":         "docker.1ms.run",
	"dockerproxy": "dockerproxy.net",
	"daocloud":    "docker.m.daocloud.io",
	"1panel":      "docker.1panel.live",
	"aliyun":      "registry.cn-hangzhou.aliyuncs.com",
	"tencent":     "mirror.ccs.tencentyun.com",
	"dockerhub":   "registry.hub.docker.com",
}

func ResolveSystem(input string) (string, bool) {
	return resolveWith(systemDomains, input)
}

func ResolveDockerCE(input string) (string, bool) {
	return resolveWith(dockerCEDomains, input)
}

func ResolveDockerRegistry(input string) (string, bool) {
	return resolveWith(dockerRegistryDomains, input)
}

func resolveWith(domains map[string]string, input string) (string, bool) {
	value := normalize(input)
	if value == "" {
		return "", false
	}
	if looksLikeAddress(value) {
		return stripProtocol(value), true
	}

	key := resolveKey(value)
	domain, ok := domains[key]
	return domain, ok
}

func resolveKey(input string) string {
	if key, ok := aliasToKey[input]; ok {
		return key
	}
	return input
}

func normalize(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func looksLikeAddress(input string) bool {
	return strings.Contains(input, ".") || strings.Contains(input, "/")
}

func stripProtocol(input string) string {
	value := strings.TrimSpace(input)
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	return strings.TrimSuffix(value, "/")
}
