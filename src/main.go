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

func SayHello(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello"))
}

func main() {
	http.HandleFunc("/srsevent", srsevent.SrsEventsHandler)
	http.HandleFunc("/room/", room.RoomHandler)
	http.ListenAndServe(":8085", nil)
}
