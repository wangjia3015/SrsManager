package manager

import (
	"net/http"
	"strings"

	"fmt"
	"github.com/golang/glog"
)

const (
	URL_PATH_EVENT     = "/event"
	URL_PATH_ROOM      = "/room"
	URL_PATH_SUMMARIES = "/summary"
	URL_PATH_STREAMS   = "/stream"
	URL_PATH_SERVER    = "/server"
)

func RestHandler(w http.ResponseWriter, req *http.Request) {
	srs_manager.HttpHandler(w, req)
}

var srs_manager *SrsManager

func InitRestHandler() error {
	dbDriver := "mysql"
	dbSource := "test:test@tcp(192.168.88.129:3306)/srs_manager"
	db := NewDBSync(dbDriver, dbSource, "srs_manager")
	var err error
	if srs_manager, err = NewSrsManager(db); err != nil {
		glog.Errorln("NewSrsManager err", err)
	}

	return err
}

type SrsManager struct {
	db                 *DBSync
	event_manager      *SrsEventManager
	room_manager       *RoomManager
	srs_server_manager *SrsManager
}

func NewSrsManager(dbSync *DBSync) (*SrsManager, error) {
	event := &SrsEventManager{db: dbSync}
	room := &RoomManager{db: dbSync}
	server, err := NewSrsServermanager(dbSync)
	if err != nil {
		return nil, fmt.Errorf("Load ip.txt failed:%v", err)
	}

	if err = server.LoadServers(); err != nil {
		return nil, err
	}
	return &SrsManager{
		db:                 dbSync,
		event_manager:      event,
		room_manager:       room,
		srs_server_manager: server,
	}, nil
}

func (s *SrsManager) HttpHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	glog.Infoln("HttpHandler url", url)
	if strings.HasPrefix(url, URL_PATH_EVENT) {
		s.event_manager.HttpHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_ROOM) {
		s.room_manager.HttpHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_SUMMARIES) ||
		strings.HasPrefix(url, URL_PATH_STREAMS) {
		s.srs_server_manager.HttpHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_SERVER) {
		s.srs_server_manager.HttpHandler(w, r)
	}
}
