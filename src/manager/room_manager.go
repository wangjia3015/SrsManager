package manager

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"time"
	"utils"
)

const (
	HTTP_POST   = "POST"
	HTTP_GET    = "GET"
	HTTP_PUT    = "PUT"
	HTTP_DELETE = "DELETE"
)

var room_manager RoomManager

func RoomHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	url := req.URL.Path

	// split
	value := "create"

	switch req.Method {
	case HTTP_POST:
		err = room_manager.CreateRoom()
	case HTTP_DELETE:
		err = room_manager.KickoffRoom(value)
	}
}

type RoomManager struct {
}

const (
	SALT = "JD_STD_2016"
)

type Room struct {
	UserName string
	Desc     string

	Token      string // ??
	StreamName string
	Expiration int64 // unixtime
}

func (r *Room) GetToken() string {
	str := fmt.Sprintf("%s_%d_%s", r.StreamName, r.Expiration, SALT)
	str = utils.GetMd5String(str)
	return str
}

// 1. 创建一条记录
func (r *RoomManager) CreateRoom(req *http.Request) (Room, error) {
	// 计算一下， 保存到数据库中， 返回
	room.Expiration = time.Now().Add(time.Hour * 24).Unix()
	room.StreamName = utils.GenerateUuid()
	room.Token = room.GetToken()
	glog.Infoln("CreateRoom", room)
	// 更新一下数据库
	return room, nil
}

func (r *RoomManager) KickoffRoom(streamName string) error {
	// 1. update from db
	// 2. delete from srs
	return nil
}
