package manager

import (
	"net/http"
	"strings"
	"utils"

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
	manager.HttpHandler(w, req)
}

var manager *SrsManager

func InitRestHandler(config *utils.Config) error {
	var err error
	dbDriver := "mysql"
	dbSource := config.GetString("dbSource")
	glog.Infoln("dbSource", dbSource)
	db := NewDBSync(dbDriver, dbSource)
	if manager, err = NewSrsManager(config, db); err != nil {
		glog.Errorln("NewSrsManager err", err)
	}

	return err
}

type SrsManager struct {
	config           *utils.Config
	db               *DBSync
	eventManager     *EventManager
	roomManager      *RoomManager
	srsServerManager *ServerManager
}

func NewSrsManager(config *utils.Config, dbSync *DBSync) (*SrsManager, error) {
	event := &EventManager{db: dbSync}
	server, err := NewSrsServermanager(dbSync)
	if err != nil {
		return nil, fmt.Errorf("Load ip.txt failed:%v", err)
	}

	if err = server.LoadServers(); err != nil {
		return nil, err
	}

	room := &RoomManager{db: dbSync, serverManager: server}
	return &SrsManager{
		config:           config,
		db:               dbSync,
		eventManager:     event,
		roomManager:      room,
		srsServerManager: server,
	}, nil
}

func (s *SrsManager) HttpHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	glog.Infoln("HttpHandler url", url)
	if strings.HasPrefix(url, URL_PATH_EVENT) {
		s.eventManager.HttpHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_ROOM) {
		s.roomManager.HttpHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_SUMMARIES) ||
		strings.HasPrefix(url, URL_PATH_STREAMS) {
		s.srsServerManager.HttpHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_SERVER) {
		s.srsServerManager.HttpHandler(w, r)
	}
}
