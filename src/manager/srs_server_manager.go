package manager

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"utils"

	"github.com/golang/glog"
)

const (
	SERVER_TYPE_EDGE_UP = iota
	SERVER_TYPE_EDGE_DOWN
	SERVER_TYPE_ORIGIN
	SERVER_TYPE_COUNT

	STR_TYPE_EDGE_UP   = "up"
	STR_TYPE_EDGE_DOWN = "down"
	STR_TYPE_ORIGIN    = "origin"
)

type ServerManager struct {
	db         *DBSync
	ipDatabase *IpDatabase
	servers    []map[string]*SrsServer
	locks      []sync.Mutex
}

func NewSrsServermanager(db *DBSync) (sm *ServerManager, err error) {

	servers := make([]map[string]*SrsServer, SERVER_TYPE_COUNT)
	for i := 0; i < SERVER_TYPE_COUNT; i++ {
		servers[i] = make(map[string]*SrsServer)
	}

	sm = &ServerManager{
		db:      db,
		servers: servers,
		locks:   make([]sync.Mutex, SERVER_TYPE_COUNT),
	}
	sm.ipDatabase, err = NewIpDatabase()
	err = sm.initServers()

	return
}

func (s *ServerManager) initServers() error {
	var err error
	for i := 0; i < SERVER_TYPE_COUNT; i++ {
		for _, svr := range s.servers[i] {
			if svr.Net, err = s.ipDatabase.GetSubNet(svr.PublicAddr); err != nil {
				return err
			}
		}

	}
	return nil
}

func (s *ServerManager) LoadServers() error {
	servers, err := s.db.LoadSrsServers()
	if err != nil {
		glog.Warningln("LoadSrsServers", err)
		return err
	}

	for _, svr := range servers {
		if ss, mutex := s.getServersByType(svr.Type); ss != nil {
			mutex.Lock()
			ss[svr.Host] = svr
			mutex.Unlock()
			go svr.UpdateStatusLoop()
		} else {
			glog.Warningln("Server type undefeine", svr)
		}
	}
	return nil
}

func (s *ServerManager) HttpHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	if strings.HasPrefix(url, URL_PATH_SUMMARIES) {
		s.summaryHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_STREAMS) {
		s.streamHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_SERVER) {
		s.serverHandler(w, r)
	}
}

// /stream/edge
func (s *ServerManager) streamHandler(w http.ResponseWriter, r *http.Request) {
	args := GetUrlParams(r.URL.Path, URL_PATH_STREAMS)

	var paramName string
	if len(args) < 1 {
		paramName = STR_TYPE_ORIGIN
	} else {
		paramName = args[0]
	}

	servers, mutex := s.getServersByName(paramName)
	if servers == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	streams := make(map[string]*StreamInfo)

	mutex.Lock()
	for h, svr := range servers {
		streams[h] = svr.GetStreams()
	}
	mutex.Unlock()

	if err := utils.WriteObjectResponse(w, streams); err != nil {
		glog.Warningln("writeRespons err", streams)
	}
}

// /summary/edge
// /summary/edge/ip/port
func (s *ServerManager) summaryHandler(w http.ResponseWriter, r *http.Request) {
	args := GetUrlParams(r.URL.Path, URL_PATH_SUMMARIES)
	if len(args) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		glog.Warningln("invalid arg num")
		return
	}

	infos := make(map[string]*SummaryInfo)
	servers, mutex := s.getServersByName(args[0])

	mutex.Lock()
	for h, svr := range servers {
		infos[h] = svr.GetSummary()
	}
	mutex.Unlock()

	if err := utils.WriteObjectResponse(w, infos); err != nil {
		glog.Warningln("writeRespons err", infos)
	}
}

type ReqCreateServer struct {
	Host       string `json:"host"`
	Desc       string `json:"desc"`
	PublicAddr string `json:"address"`
	ServerType int    `json:"type"`
}

// server/dege  PUT
func (s *ServerManager) serverHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req ReqCreateServer
		err error
	)

	if err = utils.ReadAndUnmarshalObject(r.Body, &req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Warningln("ReadAndUnmarshalObject", err)
		return
	}

	server := NewSrsServer(req.Host, req.Desc, req.PublicAddr, req.ServerType)
	if err = s.AddServer(server); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Warningln("AddsrsServer error", err, server)
		return
	}
	if err := utils.WriteObjectResponse(w, server); err != nil {
		glog.Warningln("writeRespons err", server)
	}
	glog.Infoln("AddSrsServer done", server)
}

func (s *ServerManager) AddServer(svr *SrsServer) error {
	servers, mutex := s.getServersByType(svr.Type)

	if servers == nil {
		glog.Warningln("error server type", svr.Type, svr)
		return errors.New(fmt.Sprintln("err server type", svr.Type))
	}

	if _, ok := servers[svr.Host]; ok {
		glog.Warningln("error server host already exists", svr.Host, svr)
		return errors.New(fmt.Sprintln("err server type", svr.Host))
	}

	var err error
	if err = s.ipDatabase.AddServer(svr); err != nil {
		return err
	}

	if err = s.db.InsertServer(svr); err != nil {
		return err
	}

	mutex.Lock()
	servers[svr.Host] = svr
	mutex.Unlock()
	glog.Infoln("add server", svr.Host, svr)
	go svr.UpdateStatusLoop()
	return nil
}

func (s *ServerManager) getServersByType(serverType int) (map[string]*SrsServer,
	*sync.Mutex) {
	if serverType > -1 && serverType < SERVER_TYPE_COUNT {
		return s.servers[serverType], &s.locks[serverType]
	}
	return nil, nil
}

func (s *ServerManager) getTypeByName(name string) int {
	switch name {
	case STR_TYPE_EDGE_UP:
		return SERVER_TYPE_EDGE_UP
	case STR_TYPE_EDGE_DOWN:
		return SERVER_TYPE_EDGE_DOWN
	case STR_TYPE_ORIGIN:
		return SERVER_TYPE_ORIGIN
	default:
		return -1
	}
}

func (s *ServerManager) getServersByName(name string) (map[string]*SrsServer, *sync.Mutex) {
	return s.getServersByType(s.getTypeByName(name))
}
