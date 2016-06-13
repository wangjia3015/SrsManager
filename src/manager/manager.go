package manager

import (
	"net/http"
	"strings"

	"github.com/golang/glog"
)

const (
	URL_PATH_EVENT = "/event"
	URL_PATH_ROOM  = "/room"
)

func RestHandler(w http.ResponseWriter, req *http.Request) {
	srs_manager.HttpHandler(w, req)
}

var srs_manager *SrsManager

func InitRestHandler() error {
	//dbDriver := "mysql"
	//dbSource := "123"
	//db, _ := NewDBSync(dbDriver, dbSource)
	//if err != nil {
	//	return err
	//}
	srs_manager = NewSrsManager(nil)
	return nil
}

type SrsManager struct {
	db            *DBSync
	event_manager *SrsEventManager
	room_manager  *RoomManager
}

func NewSrsManager(dbSync *DBSync) *SrsManager {
	event := &SrsEventManager{db: dbSync}
	room := &RoomManager{db: dbSync}
	return &SrsManager{
		db:            dbSync,
		event_manager: event,
		room_manager:  room,
	}
}

func (s *SrsManager) HttpHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	glog.Infoln("HttpHandler url", url)
	if strings.HasPrefix(url, URL_PATH_EVENT) {
		s.event_manager.HttpHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_ROOM) {
		s.room_manager.HttpHandler(w, r)
	}
}
