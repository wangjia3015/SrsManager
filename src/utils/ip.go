package utils

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)



type SubNet struct {
	Id        int
	Ispname   string
	supperIsp string
	Latitude  float64
	Longitude float64
	Net       *net.IPNet
	Desc      string
}

func parseIpDatabase(line string) (s *SubNet, err error) {
	recordArr := strings.Split(line, ",")
	if len(recordArr) != 7 {
		return nil, fmt.Errorf("unavalid record %v,arrlen:%v", line, len(recordArr))
	}
	s = new(SubNet)
	if _, s.Net, err = net.ParseCIDR(recordArr[1]); err != nil {
		return nil, fmt.Errorf("unavalid record %v err %v", line, err)
	}
	s.supperIsp = recordArr[2]
	s.Ispname = recordArr[3]
	s.Latitude, _ = strconv.ParseFloat(recordArr[4], 64)
	s.Longitude, _ = strconv.ParseFloat(recordArr[5], 64)
	s.Desc = recordArr[6]

	return
}

func LoadIpDatabase(database string) (subnets []*SubNet,err error) {
	subnets=make([]*SubNet,0)
	f, err := os.Open(database)
	if err != nil {
		return fmt.Errorf("can not load database file %v", database)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	index := 0
	for {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}
		index++
		if err != nil {
			return fmt.Errorf("load database file:%v err:%v", database, err)
		}
		var s *SubNet
		if s, err = parseIpDatabase(line); err != nil {
			fmt.Println(err)
			continue
		}
		subnets=append(subnets,s)
	}
	err = nil

	return
}

func main() {
	LoadIpDatabase("isp.txt")
}
