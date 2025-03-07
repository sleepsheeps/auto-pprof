package main

import (
	"auto-pprof/server"
	"os"
)

func main() {
	// 如果传入的参数路径不为空则读取配置文件, 否则默认当前路径下的config.yaml
	var cfgPath string
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	} else {
		cfgPath = "config.yaml"
	}

	server := server.NewAutoPprofServer(cfgPath)
	server.Start()
}
