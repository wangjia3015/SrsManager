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

	"github.com/golang/glog"
	"net"
)

const (
	SERVER_TYPE_EDGE_UP = iota
	SERVER_TYPE_EDGE_DOWN
	SERVER_TYPE_ORIGIN

	STR_TYPE_EDGE_UP   = "edgeup"
	STR_TYPE_EDGE_DOWN = "edgedown"
	STR_TYPE_ORIGIN    = "origin"
)

type SrsManager struct {
	EdgeUpServers   map[string]*SrsServer
	EdgeDownServers map[string]*SrsServer
	OriginServers   map[string]*SrsServer
	SubNets         map[string]*utils.SubNet
	db              *DBSync
	mutex_up        sync.Mutex
	mutex_down      sync.Mutex
	mutex_origin    sync.Mutex
}

func NewSrsServermanager(db *DBSync) (sm *SrsManager, err error) {
	sm = &SrsManager{
		db:              db,
		EdgeUpServers:   make(map[string]*SrsServer),
		EdgeDownServers: make(map[string]*SrsServer),
		OriginServers:   make(map[string]*SrsServer),
	}
	sm.SubNets, err = utils.LoadIpDatabase("isp.txt")

	return
}

func (s *SrsManager) LoadServers() error {
	servers, err := s.db.LoadSrsServers()
	if err != nil {
		glog.Warningln("LoadSrsServers", err)
		return err
	}

	for _, svr := range servers {
		if svr.ServerType == SERVER_TYPE_EDGE_UP {
			s.EdgeUpServers[svr.Host] = svr
			go svr.UpdateStatusLoop()
		} else if svr.ServerType == SERVER_TYPE_ORIGIN {
			s.OriginServers[svr.Host] = svr
			go svr.UpdateStatusLoop()
		} else if svr.ServerType == SERVER_TYPE_EDGE_DOWN {
			s.OriginServers[svr.Host] = svr
			go svr.UpdateStatusLoop()
		} else {
			glog.Warningln("Server type undefeine", svr)
		}
	}
	return nil
}

func (s *SrsManager) HttpHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *SrsManager) streamHandler(w http.ResponseWriter, r *http.Request) {
	args := GetUrlParams(r.URL.Path, URL_PATH_STREAMS)
	var svrtype int
	if len(args) == 0 || args[0] == STR_TYPE_ORIGIN {
		svrtype = SERVER_TYPE_ORIGIN
	} else if args[0] == STR_TYPE_EDGE_UP {
		svrtype = SERVER_TYPE_EDGE_UP
	} else if args[0] == STR_TYPE_EDGE_DOWN {
		svrtype = SERVER_TYPE_EDGE_DOWN
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	streams := make(map[string]*StreamInfo)
	servers := s.GetSrsServer(svrtype)
	for h, svr := range servers {
		streams[h] = svr.Streams
	}

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
func (s *SrsManager) summaryHandler(w http.ResponseWriter, r *http.Request) {
	args := GetUrlParams(r.URL.Path, URL_PATH_SUMMARIES)
	if len(args) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		glog.Warningln("invalid arg num")
		return
	}

	var servers map[string]*SrsServer
	infos := make(map[string]*SummaryInfo)
	if args[0] == STR_TYPE_EDGE_UP {
		servers = s.GetSrsServer(SERVER_TYPE_EDGE_UP)
	} else if args[0] == STR_TYPE_ORIGIN {
		servers = s.GetSrsServer(SERVER_TYPE_ORIGIN)
	} else if args[0] == STR_TYPE_EDGE_DOWN {
		servers = s.GetSrsServer(SERVER_TYPE_EDGE_DOWN)
	}

	for h, svr := range servers {
		infos[h] = svr.Summary
	}

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

func (s *SrsManager) getSubnet(addr string) (subnet *utils.SubNet, err error) {
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
func (s *SrsManager) serverHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("serverHandler")
	var (
		req    ReqCreateServer
		result []byte
		err    error
	)

	if result, err = ioutil.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		glog.Warningln("read request err", err)
		return
	} else if err = json.Unmarshal(result, &req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Warningln("json unmarshal", err)
		return
	}

	server := NewSrsServer(req.Host, req.Desc, req.ServerType, loc)
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

// stream 应该从 source 节点取
func (s *SrsManager) GetSrsServer(serverType int) map[string]*SrsServer {
	var summaries map[string]*SrsServer

	servers, mutex := s.getServersByType(serverType)
	if servers == nil {
		return summaries
	}

	mutex.Lock()
	// copy
	if servers != nil {
		summaries = make(map[string]*SrsServer)
		for k, v := range servers {
			summaries[k] = v
		}
	}
	mutex.Unlock()
	return summaries
}

func (s *SrsManager) getServersByType(serverType int) (map[string]*SrsServer,
	*sync.Mutex) {
	switch serverType {
	case SERVER_TYPE_EDGE_UP:
		return s.EdgeUpServers, &s.mutex_up
	case SERVER_TYPE_EDGE_DOWN:
		return s.EdgeDownServers, &s.mutex_down
	case SERVER_TYPE_ORIGIN:
		return s.OriginServers, &s.mutex_origin
	default:
		return nil, nil
	}
}

func (s *SrsManager) getTypeByName(name string) int {
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

func (s *SrsManager) getServersByName(name string) (map[string]*SrsServer, *sync.Mutex) {
	return s.getServersByType(s.getTypeByName(name))
}

func (s *SrsManager) AddSrsServer(svr *SrsServer) error {
	servers, mutex := s.getServersByType(svr.ServerType)

	if servers == nil {
		glog.Warningln("error server type", svr.ServerType, svr)
		return errors.New(fmt.Sprintln("err server type", svr.ServerType))
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
