package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"utils"

	"net"

	"github.com/golang/glog"
)

const (
	SERVER_TYPE_EDGE_UP = iota
	SERVER_TYPE_EDGE_DOWN
	SERVER_TYPE_ORIGIN

	STR_TYPE_EDGE_UP   = "edgeup"
	STR_TYPE_EDGE_DOWN = "edgedown"
	STR_TYPE_ORIGIN    = "origin"
)

type ServerManager struct {
	UpServers     map[string]*SrsServer
	DownServers   map[string]*SrsServer
	OriginServers map[string]*SrsServer
	SubNets       map[string]*utils.SubNet
	db            *DBSync
	upLock        sync.Mutex
	downLock      sync.Mutex
	originLock    sync.Mutex
}

func NewSrsServermanager(db *DBSync) (sm *ServerManager, err error) {
	sm = &ServerManager{
		db:            db,
		UpServers:     make(map[string]*SrsServer),
		DownServers:   make(map[string]*SrsServer),
		OriginServers: make(map[string]*SrsServer),
	}
	sm.SubNets, err = utils.LoadIpDatabase("isp.txt")

	return
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
		streams[h] = svr.Streams
	}
	mutex.Unlock()

	if b, err := json.Marshal(streams); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Warningln("Marshal", streams)
		return
	} else {
		w.Write(b)
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
		infos[h] = svr.Summary
	}
	mutex.Unlock()

	if b, err := json.Marshal(infos); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Warningln("Marshal", infos)
		return
	} else {
		w.Write(b)
	}
}

type ReqCreateServer struct {
	Host       string `json:"host"`
	Desc       string `json:"desc"`
	ServerType int    `json:"type"`
}

func (s *ServerManager) getSubnet(addr string) (subnet *utils.SubNet, err error) {
	var (
		net net.IPNet
		ok  bool
	)
	if net, err = utils.GetSubNet(addr); err != nil {
		return
	}
	if subnet, ok = s.SubNets[net.String()]; !ok {
		err = fmt.Errorf("unavali ip:%v not exsit ip.txt", addr)
	}
	return
}

// server/dege  PUT
func (s *ServerManager) serverHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req    ReqCreateServer
		result []byte
		err    error
	)
	if result, err = ioutil.ReadAll(r.Body); err != nil {
		return
	} else if err = json.Unmarshal(result, &req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO
	var subNet *utils.SubNet
	server := NewSrsServer(req.Host, req.Desc, req.ServerType, subNet)
	if err = s.AddSrsServer(server); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Warningln("AddsrsServer error", err, server)
		return
	}
	if result, err = json.Marshal(server); err != nil {
		glog.Warningln("error", err, server)
	}
	glog.Infoln("AddSrsServer done", server)
	w.Write(result)
}

func (s *ServerManager) AddSrsServer(svr *SrsServer) error {
	servers, mutex := s.getServersByType(svr.Type)

	if servers == nil {
		glog.Warningln("error server type", svr.Type, svr)
		return errors.New(fmt.Sprintln("err server type", svr.Type))
	}

	if _, ok := servers[svr.Host]; ok {
		glog.Warningln("error server host already exists", svr.Host, svr)
		return errors.New(fmt.Sprintln("err server type", svr.Host))
	}

	if err := s.db.InsertServer(svr); err != nil {
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
	switch serverType {
	case SERVER_TYPE_EDGE_UP:
		return s.UpServers, &s.upLock
	case SERVER_TYPE_EDGE_DOWN:
		return s.DownServers, &s.downLock
	case SERVER_TYPE_ORIGIN:
		return s.OriginServers, &s.originLock
	default:
		return nil, nil
	}
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
