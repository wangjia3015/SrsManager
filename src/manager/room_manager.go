package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"utils"

	"github.com/golang/glog"
)

const (
	HTTP_POST   = "POST"
	HTTP_GET    = "GET"
	HTTP_PUT    = "PUT"
	HTTP_DELETE = "DELETE"
)

func (r *RoomManager) HttpHandler(w http.ResponseWriter, req *http.Request) {
	glog.Infoln("RoomManager", req.Method)
	var err error
	var result []byte

	if result, err = ioutil.ReadAll(req.Body); err != nil {
		glog.Warningln("ReadBody err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch req.Method {
	case HTTP_POST:
		var req RoomCreateReq
		var room Room
		if err = json.Unmarshal(result, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else if room, err = r.CreateRoom(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else if result, err = json.Marshal(room); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write(result)
		}
		if err != nil {
			return
		}
	case HTTP_DELETE:
		//err = room_manager.KickoffRoom(value)
	}
}

type RoomManager struct {
	db *DBSync
}

const (
	SALT = "JD_STD_2016"
)

type RoomCreateReq struct {
	Name string
	Desc string
}

type Room struct {
	Id              int
	UserName        string
	Desc            string
	StreamName      string
	Token           string
	Expiration      int64
	PublishClientId int
	PublishHost     string
}

func GetToken(stream string, expiration int64) string {
	str := fmt.Sprintf("%s_%d_%s", stream, expiration, SALT)
	str = utils.GetMD5String(str)
	return str
}

// 1. 创建一条记录
func (r *RoomManager) CreateRoom(req RoomCreateReq) (Room, error) {
	room := Room{
		UserName: req.Name,
		Desc:     req.Desc,
	}

	// 计算一下， 保存到数据库中， 返回
	room.Expiration = time.Now().Add(time.Hour * 24).Unix()
	room.StreamName = utils.GenerateUuid()
	room.Token = GetToken(room.StreamName, room.Expiration)
	glog.Infoln("CreateRoom", room)
	// 更新一下数据库
	return room, nil
}

func (r *RoomManager) KickoffRoom(streamName string) error {
	// 1. update from db
	// 2. delete from srs
	return nil
}
