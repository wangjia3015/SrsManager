package manager

import (
	"errors"
	"fmt"
	"srs_client"
	"time"

	"github.com/golang/glog"
)

const (
	SERVER_TYPE_EDGE   = 0
	SERVER_TYPE_SOURCE = 1
	SERVER_TYPE_ALL    = 2

	UPDATE_STATUS_INTERVAL = 10 * time.Second
)

type SrsServer struct {
	ID         int64
	Host       string
	ServerType int
	Status     int

	Streams           []srs_client.Stream
	Summary           *srs_client.SummaryData
	StreamUpdateTime  int64
	SummaryUpdateTime int64
}

func (s *SrsServer) UpdateStatusLoop() {
	for {
		s.UpdateServerStreams()
		s.UpdateServerSummaries()
		time.Sleep(UPDATE_STATUS_INTERVAL)
	}
}

func (s *SrsServer) UpdateServerStreams() error {
	if rsp, err := srs_client.GetStreams(s.Host); err != nil {
		glog.Warningln("UpdateServer GetStreams", s.Host, err)
		return err
	} else if rsp.Code != 0 {
		msg := fmt.Sprintln("GetStream server return err", s.Host, rsp.Code)
		glog.Warningln(msg)
		return errors.New(msg)
	} else {
		s.Streams = rsp.Streams
		s.StreamUpdateTime = time.Now().Unix()
		//glog.Infoln("UpdateServerStreams", s.Streams)
		return nil
	}
}

func (s *SrsServer) UpdateServerSummaries() error {
	if rsp, err := srs_client.GetSummaries(s.Host); err != nil {
		glog.Warningln("UpdateServer GetSummaries", s.Host, err)
		return err
	} else if rsp.Code != 0 {
		msg := fmt.Sprintln("GetSummaries server return err", s.Host, rsp.Code)
		glog.Warningln(msg)
		return errors.New(msg)
	} else {
		s.Summary = &rsp.Data
		s.SummaryUpdateTime = time.Now().Unix()
		//glog.Infoln("UpdateServerSummaries", s.Summary)
		return nil
	}
}
