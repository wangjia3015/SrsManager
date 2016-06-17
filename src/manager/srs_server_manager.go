package manager

import (
	"encoding/json"
	"net/http"
	"strings"

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

type SrsServerManager struct {
	EdgeUpServers   map[string]*SrsServer
	EdgeDownServers map[string]*SrsServer
	OriginServers   map[string]*SrsServer
	db              *DBSync
}

func NewSrsServermanager(db *DBSync) *SrsServerManager {
	return &SrsServerManager{
		db:            db,
		EdgeUpServers: make(map[string]*SrsServer),
		OriginServers: make(map[string]*SrsServer),
	}
}

func (s *SrsServerManager) LoadServers() error {
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

func (s *SrsServerManager) HttpHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	if strings.HasPrefix(url, URL_PATH_SUMMARIES) {
		s.summaryHandler(w, r)
	} else if strings.HasPrefix(url, URL_PATH_STREAMS) {
		s.streamHandler(w, r)
	}
}

// /stream/edge
func (s *SrsServerManager) streamHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *SrsServerManager) summaryHandler(w http.ResponseWriter, r *http.Request) {
	args := GetUrlParams(r.URL.Path, URL_PATH_SUMMARIES)
	if len(args) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		glog.Warningln("invalid arg num")
		return
	}

	var servers map[string]SrsServer
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

// stream 应该从 source 节点取
func (s *SrsServerManager) GetSrsServer(serverType int) map[string]SrsServer {
	var servers map[string]*SrsServer
	var summaries map[string]SrsServer

	switch serverType {
	case SERVER_TYPE_EDGE_UP:
		servers = s.EdgeUpServers
	case SERVER_TYPE_ORIGIN:
		servers = s.OriginServers
	case SERVER_TYPE_EDGE_DOWN:
		servers = s.EdgeDownServers
	}

	// copy
	if servers != nil {
		summaries = make(map[string]SrsServer)
		for k, v := range servers {
			summaries[k] = *v
		}
	}
	return summaries
}

func (s *SrsServerManager) AddSrsServer(svr *SrsServer) error {
	return nil
}
