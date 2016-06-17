package manager

import (
	"fmt"
	"srs_client"
	"time"

	"utils"

	"github.com/golang/glog"
)

const (
	SERVER_TYPE_EDGE   = 0
	SERVER_TYPE_SOURCE = 1

	UPDATE_STATUS_INTERVAL = 10 * time.Second
)

type StreamInfo struct {
	Host       string
	Streams    []srs_client.Stream
	UpdateTime int64
}

type SummaryInfo struct {
	Host       string
	Data       srs_client.SummaryData
	UpdateTime int64
}

type SrsServer struct {
	ID         int64
	Host       string
	ServerType int
	Status     int
	Desc       string

	Loc     utils.Loc
	Streams *StreamInfo
	Summary *SummaryInfo
}

func (s *SrsServer) UpdateStatusLoop() {
	for {
		s.UpdateServerStreams()
		s.UpdateServerSummaries()
		time.Sleep(UPDATE_STATUS_INTERVAL)
	}
}

func (s *SrsServer) UpdateServerStreams() {
	if rsp, err := srs_client.GetStreams(s.Host); err != nil {
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
	if rsp, err := srs_client.GetSummaries(s.Host); err != nil {
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
