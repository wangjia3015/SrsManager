package manager

import (
	"errors"
	"fmt"
	"strings"
	"sync"
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
	ID         int64
	Host       string
	PublicHost string
	Type       int
	Status     int // 暂时没用
	Desc       string
	Net        *SubNet

	streamsLock sync.RWMutex
	summaryLock sync.RWMutex
	streams     *StreamInfo
	summary     *SummaryInfo
}

func (s *SrsServer) GetPublicAddr() (string, error) {
	strs := strings.Split(s.PublicHost, ":")
	strsLen := len(strs)
	if strsLen < 1 || strsLen > 2 {
		return "", errors.New(fmt.Sprintf("invalid PublicHost", s.PublicHost))
	}
	return strs[0], nil
}

func (s *SrsServer) GetStreams() *StreamInfo {
	s.streamsLock.RLock()
	defer s.streamsLock.RUnlock()
	return s.streams
}

func (s *SrsServer) GetSummary() *SummaryInfo {
	s.summaryLock.Lock()
	defer s.summaryLock.RUnlock()
	return s.summary
}

func (s *SrsServer) getLoad() float64 {
	return s.GetSummary().Data.Sys.Load1m * float64(s.GetSummary().Data.Sys.NetSend)
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

func NewSrsServer(host, desc, publicHost string, serverType int) *SrsServer {
	return &SrsServer{
		Host:       host,
		Type:       serverType,
		PublicHost: publicHost,
		streams:    &StreamInfo{},
		summary:    &SummaryInfo{},
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
		s.streamsLock.Lock()
		s.streams = si
		s.streamsLock.Unlock()
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
		s.summaryLock.Lock()
		s.summary = summary
		s.summaryLock.Unlock()
		//glog.Infoln("UpdateServerSummaries", s.Summary)
	}
}

func (s *SrsServer) IsAvaliable() bool {
	s.summaryLock.RLock()
	sendBytes := s.summary.Data.Sys.NetSendi
	load5m := s.summary.Data.Sys.Load5m
	load1m := s.summary.Data.Sys.Load1m
	s.summaryLock.RUnlock()
	if load1m > 64 || load5m > 64 || sendBytes > 100*1024*1024 {
		return false
	}

	return true
}
