package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

var srs_event_manager SrsEventManager

func SrsEventsHandler(w http.ResponseWriter, req *http.Request) {
	var ret []byte
	var info ConnectInfo

	result, err := ioutil.ReadAll(req.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		ret = []byte("-2")
		// log error
	} else if err = json.Unmarshal(result, &info); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ret = []byte("-1")
		// log error
	} else {
		// LOG
		fmt.Printf("%+v\n", info)
		switch info.Action {
		case SRS_CB_ACTION_ON_PUBLISH:
			err = srs_event_manager.OnPublish(info)
		}
		if err != nil {
			// TODO log
			ret = []byte("-3")
		} else {
			ret = []byte("0")
		}
	}
	w.Write(ret)
}

type SrsEventManager struct {
	// checkroom room.CheckRoomlegal // 判断room name是否有效
}

// 建立链接时
func (s *SrsEventManager) OnConnect(info ConnectInfo) error { return nil }

// 关闭连接时
func (s *SrsEventManager) OnClose(info ConnectInfo) error { return nil }

// 主播推送时
func (s *SrsEventManager) OnPublish(info ConnectInfo) error {
	return nil
}

func (s *SrsEventManager) OnUnpublish(info ConnectInfo) error { return nil }

// 用来判断用户是否有权限播放
func (s *SrsEventManager) OnPlay(info ConnectInfo) error { return nil }

// 当客户端停止播放时。
// 备注：停止播放可能不会关闭连接，还能再继续播放
func (s *SrsEventManager) OnStop(info ConnectInfo) error { return nil }
