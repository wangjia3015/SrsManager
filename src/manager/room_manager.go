package manager

import (
	"errors"
	"fmt"
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

	HTTP_HEADER_CDN_IP = "X-REAL-IP"
)

func (r *RoomManager) HttpHandler(w http.ResponseWriter, req *http.Request) {
	glog.Infoln("RoomManager", req.Method)
	var err error

	args := GetUrlParams(req.URL.Path, URL_PATH_ROOM)
	argsLen := len(args)

	remoteAddr := req.Header.Get(HTTP_HEADER_CDN_IP)
	switch req.Method {
	case HTTP_POST:
		var (
			request RoomCreateReq
			room    *Room
			result  []byte
		)
		err = utils.ReadAndUnmarshalObject(req.Body, &request)
		if err == nil {
			request.RealAddr = remoteAddr
			room, err = r.CreateRoom(request)
		}
		if err == nil {
			err = utils.WriteObjectResponse(w, room)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			glog.Warningln("POST err", req.URL.Path, string(result), err)
			return
		}
	case HTTP_DELETE:
		if argsLen != 1 {
			w.WriteHeader(http.StatusBadRequest)
			glog.Warningln("KickoffRoom invalid args count", args)
			return
		}
		if err = r.KickoffRoom(args[0]); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			glog.Warningln("KickoffRoom", err)
			return
		}
	case HTTP_GET:
		if argsLen != 1 {
			w.WriteHeader(http.StatusBadRequest)
			glog.Warningln("KickoffRoom invalid args count", args)
			return
		}
		rsp := r.GetRoom(args[0], remoteAddr)
		if err = utils.WriteObjectResponse(w, rsp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			glog.Warningln("GET err", req.URL.Path, rsp, err)
			return
		}
	}
}

type RoomManager struct {
	db            *DBSync
	serverManager *ServerManager
}

const (
	SALT = "JD_STD_2016"
)

type RoomCreateReq struct {
	Name     string
	Desc     string
	RealAddr string
}

const (
	ROOM_CREATE  = iota // 刚刚创建流 未推送
	ROOM_PUBLISH        // 正在推送中
	ROOM_CLOSED         // 推送结束
)

type Room struct {
	Id       int64  //
	UserName string //
	Desc     string //

	//UUID            string // 推送端唯一ID
	StreamName string // 随机生成的ID 作为唯一标识
	Expiration int64  // 过期时间
	Token      string // 不保存
	Status     int    // 判断状态
	Addrs      []string

	PublishClientId int    // 推送端的ID与PublishHost 一起作为KICKOFF回调的参数
	PublishHost     string // 边缘节点的IP

	CreateTime     int64
	LastUpdateTime int64
}

type ReqRoomResponse struct {
	StreamName string
	Servers    []string
}

func GetToken(stream string, expiration int64) string {
	str := fmt.Sprintf("%s_%d_%s", stream, expiration, SALT)
	str = utils.GetMD5String(str)
	return str
}

// 1. 创建一条记录
func (r *RoomManager) CreateRoom(req RoomCreateReq) (*Room, error) {
	room := &Room{
		UserName: req.Name,
		Desc:     req.Desc,
	}

	room.StreamName = utils.GenerateUuid()
	room.Expiration = time.Now().Add(time.Hour * 24).Unix()
	room.Token = GetToken(room.StreamName, room.Expiration)
	room.Status = ROOM_CREATE

	room.Addrs = r.serverManager.GetServers(req.RealAddr, SERVER_TYPE_EDGE_UP)

	// insert to db
	if err := r.db.InsertRoom(room); err != nil {
		return nil, err
	}
	glog.Infoln("CreateRoom", room)
	return room, nil
}

func (r *RoomManager) GetRoom(streamName, remoteAddr string) ReqRoomResponse {
	var rsp ReqRoomResponse
	rsp.StreamName = streamName
	rsp.Servers = r.serverManager.GetServers(remoteAddr, SERVER_TYPE_EDGE_DOWN)
	return rsp
}

func (r *RoomManager) tryKickOffClient(host string, clientID int) (err error) {
	var rsp utils.RspBase
	for i := 0; i < 3; i++ {
		if rsp, err = utils.KickOffClient(host, clientID); err != nil || rsp.Code != 0 {
			glog.Warningln("KickOffClient ", host, clientID, err, rsp.Code)
			continue
		}
		break
	}
	return err
}

func (r *RoomManager) KickoffRoom(streamName string) error {
	// 1. update from db
	// 2. delete from srs
	var room *Room
	var err error
	params := map[string]interface{}{"streamname": streamName}
	if room, err = r.db.SelectRoom(params); err != nil {
		return err
	} else if room == nil {
		return errors.New("stream name not exists " + streamName)
	}

	room.Status = ROOM_CLOSED

	// update
	if err = r.db.UpdateRoom(room); err != nil {
		glog.Warningln("UpdateRoom", err)
		return err
	}

	if err = r.tryKickOffClient(room.PublishHost, room.PublishClientId); err != nil {
		glog.Warningln("tryKickOffClient", err)
		return err
	}

	return nil
}
