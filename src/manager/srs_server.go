package manager

import (
	"errors"
	"fmt"
	"srs_client"
	"time"

	"github.com/golang/glog"
	"utils"
)

const (
	SERVER_TYPE_EDGE   = 0
	SERVER_TYPE_SOURCE = 1
	SERVER_TYPE_ALL    = 2

	UPDATE_STATUS_INTERVAL = 10 * time.Second
)

type SrsServer struct {
	ID                int64
	Host              string
	ServerType        int
	Status            int
	Desc              string
	Loc               utils.Loc
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

func (s *SrsServer) UpdateServerStreams() {
	if rsp, err := srs_client.GetStreams(s.Host); err != nil {
		glog.Warningln("UpdateServer GetStreams", s.Host, err)
		return err
	} else if rsp.Code != 0 {
		msg := fmt.Sprintln("GetStream server return err", s.Host, rsp.Code)
		glog.Warningln(msg)
	} else {
		s.Streams = rsp.Streams
		s.StreamUpdateTime = time.Now().Unix()
	}
}

func (s *SrsServer) UpdateServerSummaries() {
	if rsp, err := srs_client.GetSummaries(s.Host); err != nil {
		glog.Warningln("UpdateServer GetSummaries", s.Host, err)
	} else if rsp.Code != 0 {
		msg := fmt.Sprintln("GetSummaries server return err", s.Host, rsp.Code)
		glog.Warningln(msg)
	} else {
		s.Summary = &rsp.Data
		s.SummaryUpdateTime = time.Now().Unix()
	}
}
