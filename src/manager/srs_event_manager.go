package manager

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/golang/glog"
)

const (
	SRS_CB_ACTION_ON_CONNECT   = "on_connect"   // 连接
	SRS_CB_ACTION_ON_CLOSE     = "on_close"     // 连接关闭
	SRS_CB_ACTION_ON_PUBLISH   = "on_publish"   // 开始推流
	SRS_CB_ACTION_ON_UNPUBLISH = "on_unpublish" // 停止推流
	SRS_CB_ACTION_ON_PLAY      = "on_play"      // 开始播放
	SRS_CB_ACTION_ON_STOP      = "on_stop"      // 暂停播放
)

/*
{
	"action": "on_connect",
	"client_id": 1985,
	"ip": "192.168.1.10",
    "vhost": "video.test.com",
	"app": "live",
	"tcUrl": "rtmp://x/x?key=xxx",
	"pageUrl": "http://x/x.html"
}
*/
type ConnectInfo struct {
	Action     string `json:"action"`
	ClientID   int    `json:"client_id"`
	Ip         string `json:"ip"`
	VHost      string `json:"vhost"`
	AppName    string `json:"app"`
	StreamName string `json:"stream"`  // connect | close 不需要
	TcUrl      string `json:"tcUrl"`   // connect 专属
	PageUrl    string `json:"pageUrl"` // connect 专属
}

func (s *SrsEventManager) HttpHandler(w http.ResponseWriter, req *http.Request) {
	glog.Infoln("SrsEventManager")
	ret := 0
	var info ConnectInfo

	result, err := ioutil.ReadAll(req.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		glog.Warningln("read request err", err)
		ret = -1
	} else if err = json.Unmarshal(result, &info); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ret = -1
		glog.Warningln("json unmarshal", err)
	} else {
		glog.Infoln(string(result))
		glog.Infof("%+v\n", info)
		switch info.Action {
		case SRS_CB_ACTION_ON_PUBLISH:
			err = s.OnPublish(info)
		}
		if err != nil {
			ret = -1
			glog.Warningln("handler", info.Action, err)
		}
	}
	w.Write([]byte(strconv.Itoa(ret)))
}

type SrsEventManager struct {
	db *DBSync
}

// 建立链接时
func (s *SrsEventManager) OnConnect(info ConnectInfo) error {
	return nil
}

// 关闭连接时
func (s *SrsEventManager) OnClose(info ConnectInfo) error { return nil }

// 用来判断用户是否有权限播放
func (s *SrsEventManager) OnPlay(info ConnectInfo) error { return nil }

// 当客户端停止播放时。
// 备注：停止播放可能不会关闭连接，还能再继续播放
func (s *SrsEventManager) OnStop(info ConnectInfo) error { return nil }

func (s *SrsEventManager) OnUnpublish(info ConnectInfo) error { return nil }

// 主播推送时
func (s *SrsEventManager) OnPublish(info ConnectInfo) error {
	glog.Infoln("OnPublish", info)
	var room *Room
	var err error
	params := map[string]interface{}{"streamname": info.StreamName}
	if room, err = s.db.SelectRoom(params); err != nil {
		return err
	} else if room == nil {
		return errors.New("stream name not exists " + info.StreamName)
	} else if room.Status == ROOM_CLOSED {
		return errors.New("stream already closed " + info.StreamName)
	}

	room.PublishClientId = info.ClientID
	room.PublishHost = info.Ip // FIXME 此处IP是远端IP 并不是edge server的IP 需要远端edge server的 IP:PORT 来处理
	room.Status = ROOM_PUBLISH
	// update
	if err = s.db.UpdateRoom(room); err != nil {
		return err
	}
	return nil
}
