package manager

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"strings"
	"sync"
	"utils"
)

const (
	SERVER_TYPE_EDGE_UP = iota
	SERVER_TYPE_EDGE_DOWN
	SERVER_TYPE_ORIGIN
	SERVER_TYPE_COUNT

	STR_TYPE_EDGE_UP   = "up"
	STR_TYPE_EDGE_DOWN = "down"
	STR_TYPE_ORIGIN    = "origin"

	DefaultDisPatchCount = 2
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

func (s *ServerManager) initServers() (err error) {
	for i := 0; i < SERVER_TYPE_COUNT; i++ {
		for _, svr := range s.servers[i] {
			var addr string
			if addr, err = svr.GetPublicAddr(); err != nil {
				return err
			}
			if svr.Net, err = s.ipDatabase.GetSubNet(addr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ServerManager) GetServers(addr string, disType int) (result []string) {
	servers := s.ipDatabase.DisPatch(addr, disType, DefaultDisPatchCount)
	result = make([]string, 0)
	for _, svr := range servers {
		result = append(result, svr.Addr)
	}

	return
}

func (s *ServerManager) LoadServers() error {
	servers, err := s.db.LoadSrsServers()
	if err != nil {
		return fmt.Errorf("Load Srsservers error:%v", err)
	}

	for _, svr := range servers {
		if ss, mutex := s.getServersByType(svr.Type); ss != nil {
			mutex.Lock()
			ss[svr.Addr] = svr
			mutex.Unlock()
			go svr.UpdateStatusLoop()
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

type StreamTotalInfo struct {
	Name    string
	AppName string

	ClientTotal    int
	SendBytesTotal int64
	RecvBytesTotal int64
	KbpsTotal      utils.KbpsInfo
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

	var result interface{}
	result = streams
	if len(args) > 2 {
		var total StreamTotalInfo
		total.AppName = args[1]
		total.Name = args[2]
		for _, v := range streams {
			for _, i := range v.Streams {
				if i.Name == total.Name && i.AppName == total.AppName {
					total.ClientTotal += i.ClientNum
					total.SendBytesTotal += i.SendBytes
					total.RecvBytesTotal += i.RecvBytes
					total.KbpsTotal.Recv30s += i.Kbps.Recv30s
					total.KbpsTotal.Send30s += i.Kbps.Send30s
				}
			}
		}
		result = &total
	}

	if err := utils.WriteObjectResponse(w, result); err != nil {
		glog.Warningln("writeRespons err", result)
	}
}

// /summary/edge
// /summary/edge/ip/port
func (s *ServerManager) summaryHandler(w http.ResponseWriter, r *http.Request) {
	args := GetUrlParams(r.URL.Path, URL_PATH_SUMMARIES)
	if len(args) < 1 {
		w.WriteHeader(http.StatusBadRequest)
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
		glog.Warningf("summaryHandler-writeRespons infos[%v] err[%v]\n", infos, err)
	}
}

type ReqCreateServer struct {
	Addr       string `json:"addr"`
	Desc       string `json:"desc"`
	ServerType int    `json:"type"`
}

// server/dege  PUT
func (s *ServerManager) serverHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req    ReqCreateServer
		err    error
		server *SrsServer
	)
	code := http.StatusBadRequest
	if err = utils.ReadAndUnmarshalObject(r.Body, &req); err != nil {
		goto errDeal
	}

	server = NewSrsServer(req.Addr, req.Desc, req.ServerType)
	if err = s.AddServer(server); err != nil {
		code = http.StatusInternalServerError
		goto errDeal
	}

	if err = utils.WriteObjectResponse(w, server); err != nil {
		goto errDeal
	}

	return
errDeal:
	w.WriteHeader(code)
	glog.Warningf("Add server error-req[%v] err[%v]\n", req, err)

}

func (s *ServerManager) AddServer(svr *SrsServer) (err error) {
	servers, mutex := s.getServersByType(svr.Type)
	if servers == nil {
		return fmt.Errorf("AddServer-err server type[%v]", svr.Type)
	}

	if _, ok := servers[svr.Addr]; ok {
		return fmt.Errorf("AddServer-error server[%v] host already exists", svr.Addr)
	}

	if err = s.ipDatabase.AddServer(svr); err != nil {
		return fmt.Errorf("AddServer-IpDataBase Add server:%v err:%v", svr.Addr, err)
	}

	if err = s.db.InsertServer(svr); err != nil {
		return fmt.Errorf("AddServer-dbInsert server:%v err:%v", svr.Addr, err)
	}

	mutex.Lock()
	servers[svr.Addr] = svr
	mutex.Unlock()
	go svr.UpdateStatusLoop()

	return
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
