package config

import (
	"flag"
)

var (
	AdminURL  = flag.String("adminUrl", "", "admin 的地址")
	Profile   = flag.String("profile", "", "环境")
	ProxyFile = flag.String("proxyFile", "", "测试环境的数据")
)

func InitConfig() {
	flag.Parse()
}
