package manager

import (
	"fmt"
	"time"
	"utils"

	"github.com/golang/glog"
)

const (
	UPDATE_STATUS_INTERVAL = 10 * time.Second
)

type StreamInfo struct {
	Host       string
	Streams    []utils.Stream
	UpdateTime int64
}

type SummaryInfo struct {
	Host       string
	Data       utils.SummaryData
	UpdateTime int64
}

type SrsServer struct {
	ID     int64
	Host   string
	Type   int
	Status int // 暂时没用

	Desc string

	Net     *SubNet
	Streams *StreamInfo
	Summary *SummaryInfo
}

func (s *SrsServer) getLoad() float64 {
	return s.Summary.Data.Sys.Load1m * float64(s.Summary.Data.Sys.NetSend)
}

type SortSrsServers []*SrsServer

func (sp SortSrsServers) Len() int {
	return len(sp)
}

func (sp SortSrsServers) Swap(i, j int) {
	sp[i], sp[j] = sp[j], sp[i]
}

func (sp SortSrsServers) Less(i, j int) bool {
	return sp[i].getLoad() < sp[j].getLoad()
}

func NewSrsServer(host, desc string, serverType int) *SrsServer {
	return &SrsServer{
		Host: host,
		Type: serverType,
	}
}

func (s *SrsServer) UpdateStatusLoop() {
	for {
		s.UpdateServerStreams()
		s.UpdateServerSummaries()
		time.Sleep(UPDATE_STATUS_INTERVAL)
	}
}

func (s *SrsServer) UpdateServerStreams() {
	if rsp, err := utils.GetStreams(s.Host); err != nil {
		glog.Warningln("UpdateServer GetStreams", s.Host, err)
	} else if rsp.Code != 0 {
		msg := fmt.Sprintln("GetStream server return err", s.Host, rsp.Code)
		glog.Warningln(msg)
	} else {
		si := &StreamInfo{Host: s.Host, UpdateTime: time.Now().Unix()}
		si.Streams = rsp.Streams
		s.Streams = si
		//glog.Infoln("UpdateServerStreams", s.Streams)
	}
}

func (s *SrsServer) UpdateServerSummaries() {
	if rsp, err := utils.GetSummaries(s.Host); err != nil {
		glog.Warningln("UpdateServer GetSummaries", s.Host, err)
	} else if rsp.Code != 0 {
		msg := fmt.Sprintln("GetSummaries server return err", s.Host, rsp.Code)
		glog.Warningln(msg)
	} else {
		summary := &SummaryInfo{Host: s.Host, UpdateTime: time.Now().Unix()}
		summary.Data = rsp.Data
		s.Summary = summary
		//glog.Infoln("UpdateServerSummaries", s.Summary)
	}
}
