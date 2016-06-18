package main

import (
	"flag"
	"fmt"
	"manager"
	"net/http"
	"utils"
)

var configPath = flag.String("c", "", "config file path")

/*
	ROOM:
	1. 创建room，
	2. 关闭发布者
	4. 查询所有的ROOM
	3. 关闭room
	// 1. room 管理 创建room, 并且可以关闭room 或者关闭发布者


	监控功能
	1. 统计一共有多少个stream， 每个流有多少个client
	2. 定时拉取每个srs server的系统信息
*/

func GetConfig(path string) (*utils.Config, error) {
	var err error
	config := utils.NewConfig()
	if err = config.LoadFromFile(path); err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	flag.Parse()

	var (
		config *utils.Config
		err    error
		port   int
	)
	if config, err = GetConfig(*configPath); err != nil {
		fmt.Println("GetConfig", err)
		return
	} else if err := manager.InitRestHandler(config); err != nil {
		fmt.Println("err", err)
		return
	}
	if port = config.GetInt("port"); port < 0 {
		fmt.Println("Invalid port param")
		return
	}
	fmt.Println("Init success")
	http.HandleFunc("/", manager.RestHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
