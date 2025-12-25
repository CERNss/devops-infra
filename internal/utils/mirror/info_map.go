package mirror

import "strings"

type Kind string

const (
	KindSystem         Kind = "system"
	KindDockerCE       Kind = "docker-ce"
	KindDockerRegistry Kind = "docker-registry"
)

type SystemCategory string

const (
	CategoryDefault SystemCategory = "default"
	CategoryEdu     SystemCategory = "edu"
	CategoryAbroad  SystemCategory = "abroad"
)

var categoryAlias = map[string]SystemCategory{
	"国内":      CategoryDefault,
	"默认":      CategoryDefault,
	"大陆":      CategoryDefault,
	"china":     CategoryDefault,
	"cn":        CategoryDefault,
	"default":   CategoryDefault,
	"mainland":  CategoryDefault,
	"教育":      CategoryEdu,
	"教育网":    CategoryEdu,
	"校园":      CategoryEdu,
	"edu":       CategoryEdu,
	"education": CategoryEdu,
	"海外":      CategoryAbroad,
	"境外":      CategoryAbroad,
	"abroad":    CategoryAbroad,
	"oversea":   CategoryAbroad,
	"overseas":  CategoryAbroad,
}

var aliasToKey = map[string]string{
	"阿里":             "aliyun",
	"阿里云":           "aliyun",
	"ali":              "aliyun",
	"aliyun":           "aliyun",
	"腾讯":             "tencent",
	"腾讯云":           "tencent",
	"tencent":          "tencent",
	"tencentyun":       "tencent",
	"华为":             "huawei",
	"华为云":           "huawei",
	"huawei":           "huawei",
	"huaweicloud":      "huawei",
	"天翼云":           "ctyun",
	"ctyun":            "ctyun",
	"netease":          "netease",
	"163":              "netease",
	"网易":             "netease",
	"volc":             "volc",
	"volces":           "volc",
	"火山":             "volc",
	"清华":             "tsinghua",
	"清华大学":         "tsinghua",
	"tuna":             "tsinghua",
	"tsinghua":         "tsinghua",
	"北大":             "pku",
	"北京大学":         "pku",
	"pku":              "pku",
	"浙大":             "zju",
	"浙江大学":         "zju",
	"zju":              "zju",
	"南大":             "nju",
	"南京大学":         "nju",
	"nju":              "nju",
	"交大":             "sjtu",
	"上海交通大学":     "sjtu",
	"sjtu":             "sjtu",
	"中科大":           "ustc",
	"中国科学技术大学": "ustc",
	"ustc":             "ustc",
	"中科院":           "iscas",
	"iscas":            "iscas",
	"azure":            "azure",
	"1ms":              "1ms",
	"dockerproxy":      "dockerproxy",
	"daocloud":         "daocloud",
	"1panel":           "1panel",
	"dockerhub":        "dockerhub",
	"docker":           "docker",
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

var systemCategoryDomains = map[SystemCategory]map[string]struct{}{
	CategoryDefault: newDomainSet(
		"mirrors.aliyun.com",
		"mirrors.tencent.com",
		"mirrors.huaweicloud.com",
		"mirrors.cmecloud.cn",
		"mirrors.ctyun.cn",
		"mirrors.163.com",
		"mirrors.volces.com",
		"mirrors.tuna.tsinghua.edu.cn",
		"mirrors.pku.edu.cn",
		"mirrors.zju.edu.cn",
		"mirrors.nju.edu.cn",
		"mirror.lzu.edu.cn",
		"mirror.sjtu.edu.cn",
		"mirrors.cqupt.edu.cn",
		"mirrors.ustc.edu.cn",
		"mirror.iscas.ac.cn",
	),
	CategoryEdu: newDomainSet(
		"mirrors.pku.edu.cn",
		"mirror.bjtu.edu.cn",
		"mirrors.bfsu.edu.cn",
		"mirrors.bupt.edu.cn",
		"mirrors.cqu.edu.cn",
		"mirrors.cqupt.edu.cn",
		"mirrors.neusoft.edu.cn",
		"mirrors.uestc.cn",
		"mirrors.scau.edu.cn",
		"mirrors.hust.edu.cn",
		"mirrors.jlu.edu.cn",
		"mirrors.jcut.edu.cn",
		"mirrors.jxust.edu.cn",
		"mirror.lzu.edu.cn",
		"mirrors.nju.edu.cn",
		"mirrors.njtech.edu.cn",
		"mirrors.njupt.edu.cn",
		"mirrors.sustech.edu.cn",
		"mirror.nyist.edu.cn",
		"mirrors.qlu.edu.cn",
		"mirrors.tuna.tsinghua.edu.cn",
		"mirrors.sdu.edu.cn",
		"mirrors.shanghaitech.edu.cn",
		"mirror.sjtu.edu.cn",
		"mirrors.sjtug.sjtu.edu.cn",
		"mirrors.wsyu.edu.cn",
		"mirrors.xjtu.edu.cn",
		"mirrors.nwafu.edu.cn",
		"mirrors.zju.edu.cn",
		"mirrors.ustc.edu.cn",
	),
	CategoryAbroad: newDomainSet(
		"mirrors.xtom.hk",
		"mirror.01link.hk",
		"download.nus.edu.sg/mirror",
		"mirror.sg.gs",
		"mirrors.xtom.sg",
		"free.nchc.org.tw",
		"mirror.ossplanet.net",
		"linux.cs.nctu.edu.tw",
		"ftp.tku.edu.tw",
		"mirror.twds.com.tw",
		"mirror.anigil.com",
		"ftp.udx.icscoe.jp/Linux",
		"ftp.jaist.ac.jp/pub/Linux",
		"linux2.yz.yamagata-u.ac.jp/pub/Linux",
		"mirrors.xtom.jp",
		"mirrors.gbnetwork.com",
		"mirror.kku.ac.th",
		"mirror.vorboss.net",
		"mirror.quickhost.uk",
		"mirror.dogado.de",
		"mirrors.xtom.de",
		"ftp.halifax.rwth-aachen.de",
		"ftp.agdsn.de",
		"mirror.in2p3.fr/pub/linux",
		"mirrors.ircam.fr/pub",
		"eclats.crans.org",
		"ftp.crihan.fr",
		"mirrors.xtom.nl",
		"mirror.datapacket.com",
		"eu.edge.kernel.org",
		"mirrors.xtom.ee",
		"mirror.netsite.dk",
		"mirrors.dotsrc.org",
		"mirror.accum.se",
		"ftp.lysator.liu.se",
		"mirror.yandex.ru",
		"mirror.linux-ia64.org",
		"mirror.truenetwork.ru",
		"ftp.belnet.be/mirror",
		"ftp.cc.uoc.gr/mirrors/linux",
		"ftp.fi.muni.cz/pub/linux",
		"ftp.sh.cvut.cz",
		"mirror.karneval.cz/pub/linux",
		"mirrors.nic.cz",
		"mirror.ethz.ch",
		"mirrors.kernel.org",
		"mirrors.mit.edu",
		"mirror.math.princeton.edu/pub",
		"ftp-chi.osuosl.org/pub",
		"mirror.fcix.net",
		"mirrors.xtom.com",
		"mirror.steadfast.net",
		"mirror.it.ubc.ca",
		"mirror.xenyth.net",
		"mirrors.switch.ca",
		"mirror.pop-sc.rnp.br/mirror",
		"mirror.uepg.br",
		"mirror.ufscar.br",
		"mirrors.eze.sysarmy.com",
		"gsl-syd.mm.fcix.net",
		"mirror.aarnet.edu.au/pub",
		"mirror.datamossa.io",
		"mirror.amaze.com.au",
		"mirrors.xtom.au",
		"mirror.overthewire.com.au",
		"mirror.fsmg.org.nz",
		"mirror.liquidtelecom.com",
		"mirror.dimensiondata.com",
	),
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

var dockerCECategoryDomains = map[SystemCategory]map[string]struct{}{
	CategoryDefault: newDomainSet(
		"mirrors.aliyun.com/docker-ce",
		"mirrors.tencent.com/docker-ce",
		"mirrors.huaweicloud.com/docker-ce",
		"mirrors.163.com/docker-ce",
		"mirrors.volces.com/docker",
		"mirror.azure.cn/docker-ce",
		"mirrors.tuna.tsinghua.edu.cn/docker-ce",
		"mirrors.pku.edu.cn/docker-ce",
		"mirrors.zju.edu.cn/docker-ce",
		"mirrors.nju.edu.cn/docker-ce",
		"mirror.sjtu.edu.cn/docker-ce",
		"mirrors.ustc.edu.cn/docker-ce",
		"mirror.iscas.ac.cn/docker-ce",
	),
	CategoryEdu: newDomainSet(
		"mirrors.tuna.tsinghua.edu.cn/docker-ce",
		"mirrors.pku.edu.cn/docker-ce",
		"mirrors.zju.edu.cn/docker-ce",
		"mirrors.nju.edu.cn/docker-ce",
		"mirror.sjtu.edu.cn/docker-ce",
		"mirrors.ustc.edu.cn/docker-ce",
		"mirror.iscas.ac.cn/docker-ce",
	),
	CategoryAbroad: newDomainSet(
		"download.docker.com",
	),
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

var dockerRegistryCategoryDomains = map[SystemCategory]map[string]struct{}{
	CategoryDefault: newDomainSet(
		"docker.1ms.run",
		"dockerproxy.net",
		"docker.m.daocloud.io",
		"docker.1panel.live",
		"registry.cn-hangzhou.aliyuncs.com",
		"registry.cn-shanghai.aliyuncs.com",
		"registry.cn-qingdao.aliyuncs.com",
		"registry.cn-beijing.aliyuncs.com",
		"registry.cn-zhangjiakou.aliyuncs.com",
		"registry.cn-huhehaote.aliyuncs.com",
		"registry.cn-wulanchabu.aliyuncs.com",
		"registry.cn-shenzhen.aliyuncs.com",
		"registry.cn-heyuan.aliyuncs.com",
		"registry.cn-guangzhou.aliyuncs.com",
		"registry.cn-chengdu.aliyuncs.com",
		"registry.cn-hongkong.aliyuncs.com",
		"mirror.ccs.tencentyun.com",
	),
	CategoryEdu: newDomainSet(
		"docker.1ms.run",
		"dockerproxy.net",
		"docker.m.daocloud.io",
		"docker.1panel.live",
		"registry.cn-hangzhou.aliyuncs.com",
		"registry.cn-shanghai.aliyuncs.com",
		"registry.cn-qingdao.aliyuncs.com",
		"registry.cn-beijing.aliyuncs.com",
		"registry.cn-zhangjiakou.aliyuncs.com",
		"registry.cn-huhehaote.aliyuncs.com",
		"registry.cn-wulanchabu.aliyuncs.com",
		"registry.cn-shenzhen.aliyuncs.com",
		"registry.cn-heyuan.aliyuncs.com",
		"registry.cn-guangzhou.aliyuncs.com",
		"registry.cn-chengdu.aliyuncs.com",
		"registry.cn-hongkong.aliyuncs.com",
		"mirror.ccs.tencentyun.com",
	),
	CategoryAbroad: newDomainSet(
		"registry.ap-northeast-1.aliyuncs.com",
		"registry.ap-southeast-1.aliyuncs.com",
		"registry.ap-southeast-3.aliyuncs.com",
		"registry.ap-southeast-5.aliyuncs.com",
		"registry.eu-central-1.aliyuncs.com",
		"registry.eu-west-1.aliyuncs.com",
		"registry.us-west-1.aliyuncs.com",
		"registry.us-east-1.aliyuncs.com",
		"registry.me-east-1.aliyuncs.com",
		"gcr.io",
		"asia.gcr.io",
		"eu.gcr.io",
		"registry.hub.docker.com",
	),
}

func ResolveSystem(input string) (string, bool) {
	category, value := splitCategoryInput(input)
	resolved, ok := resolveWith(systemDomains, value)
	if !ok {
		return "", false
	}
	if category == "" {
		return resolved, true
	}
	normalizedCategory := normalizeCategory(category)
	if normalizedCategory == "" {
		return "", false
	}
	if !categoryContains(normalizedCategory, resolved, systemCategoryDomains) {
		return "", false
	}
	return resolved, true
}

func ResolveDockerCE(input string) (string, bool) {
	category, value := splitCategoryInput(input)
	resolved, ok := resolveWith(dockerCEDomains, value)
	if !ok {
		return "", false
	}
	if category == "" {
		return resolved, true
	}
	normalizedCategory := normalizeCategory(category)
	if normalizedCategory == "" {
		return "", false
	}
	if !categoryContains(normalizedCategory, resolved, dockerCECategoryDomains) {
		return "", false
	}
	return resolved, true
}

func ResolveDockerRegistry(input string) (string, bool) {
	category, value := splitCategoryInput(input)
	resolved, ok := resolveWith(dockerRegistryDomains, value)
	if !ok {
		return "", false
	}
	if category == "" {
		return resolved, true
	}
	normalizedCategory := normalizeCategory(category)
	if normalizedCategory == "" {
		return "", false
	}
	if !categoryContains(normalizedCategory, resolved, dockerRegistryCategoryDomains) {
		return "", false
	}
	return resolved, true
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

func splitCategoryInput(input string) (string, string) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", ""
	}
	separators := []string{"-", "－", "—", "–"}
	for _, separator := range separators {
		if strings.Contains(value, separator) {
			parts := strings.SplitN(value, separator, 2)
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}
	return "", value
}

func normalizeCategory(input string) SystemCategory {
	value := normalize(input)
	if value == "" {
		return ""
	}
	if category, ok := categoryAlias[value]; ok {
		return category
	}
	return ""
}

func categoryContains(
	category SystemCategory,
	domain string,
	categories map[SystemCategory]map[string]struct{},
) bool {
	if category == "" {
		return true
	}
	set, ok := categories[category]
	if !ok {
		return false
	}
	_, exists := set[domain]
	return exists
}

func newDomainSet(domains ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		set[domain] = struct{}{}
	}
	return set
}
