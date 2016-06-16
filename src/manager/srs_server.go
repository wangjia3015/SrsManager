package manager

import "srs_client"

type SrsServer struct {
	ID         int64
	Host       string
	ServerType int
	Status     int

	Streams []srs_client.Stream
}

//func (s *SrsServer)
