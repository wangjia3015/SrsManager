package manager

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang/glog"
)

type SrsServerManager struct {
	EdgeServers   map[string]*SrsServer
	SourceServers map[string]*SrsServer
	db            *DBSync
}

func NewSrsServermanager(db *DBSync) *SrsServerManager {
	return &SrsServerManager{
		db:            db,
		EdgeServers:   make(map[string]*SrsServer),
		SourceServers: make(map[string]*SrsServer),
	}
}

func (s *SrsServerManager) LoadServers() error {
	servers, err := s.db.LoadSrsServers()
	if err != nil {
		glog.Warningln("LoadSrsServers", err)
		return err
	}

	for _, svr := range servers {
		if svr.ServerType == SERVER_TYPE_EDGE {
			s.EdgeServers[svr.Host] = svr
			go svr.UpdateStatusLoop()
		} else if svr.ServerType == SERVER_TYPE_SOURCE {
			s.SourceServers[svr.Host] = svr
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
	if len(args) == 0 || args[0] == "source" {
		svrtype = SERVER_TYPE_SOURCE
	} else if args[0] == "edge" {
		svrtype = SERVER_TYPE_EDGE
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
	if args[0] == "edge" {
		servers = s.GetSrsServer(SERVER_TYPE_EDGE)
	} else if args[0] == "source" {
		servers = s.GetSrsServer(SERVER_TYPE_SOURCE)
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
	case SERVER_TYPE_EDGE:
		servers = s.EdgeServers
	case SERVER_TYPE_SOURCE:
		servers = s.SourceServers
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
