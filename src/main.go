package main

import (
	"net/http"
	"room"
	"srsevent"
)

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

func main() {
	http.HandleFunc("/event", srsevent.SrsEventsHandler)
	http.HandlerFunc("/room/", nil) // PUT 处理room 创建
	http.HandlerFunc("/room/", nil) // DELETE 处理踢人
	http.ListenAndServe(":8085", nil)
}
