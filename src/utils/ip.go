package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type Isp struct {
	name    string
	subnets map[string]*SubNet
	nets     []string

}

func NewIsp(name string) (i *Isp) {
	return &Isp{name: name, subnets: make(map[string]*SubNet),nets:make([]string,0)}
}

type SubNet struct {
	Id        int
	Country   string
	Province  string
	City      string
	Ispname   string
	supperIsp string
	Latitude  float64
	Longitude float64
	Net       *net.IPNet
}

func (s *SubNet)ToString()(ss string){
	var(
		province,city string
	)
	if len(s.Province)!=0 {
		province="-"+s.Province
	}
	if len(s.City)!=0 {
		city="-"+s.City
	}
	return strconv.Itoa(s.Id)+","+s.Net.String()+","+s.supperIsp+","+s.Ispname+","+strconv.FormatFloat(s.Latitude, 'f', 6, 64)+
	","+strconv.FormatFloat(s.Longitude,'f',6,64)+",("+s.Country+province+city+")"

}

func parseIpDatabase(index int, line string) (s *SubNet, err error) {
	recordArr := strings.Split(line, ",")
	if len(recordArr) != 14 {
		return nil, fmt.Errorf("unavalid record %v,arrlen:%v", line,len(recordArr))
	}
	s = new(SubNet)
	if _, s.Net, err = net.ParseCIDR(recordArr[0]); err != nil {
		return nil, fmt.Errorf("unavalid record %v err %v", line, err)
	}
	s.Country = strings.Trim(recordArr[1],"\"")
	s.Province = strings.Trim(recordArr[2],"\"")
	s.Id = index
	s.City = strings.Trim(recordArr[3],"\"")
	s.Latitude, _ = strconv.ParseFloat(strings.Trim(recordArr[6],"\""), 64)
	s.Longitude, _ = strconv.ParseFloat(strings.Trim(recordArr[7],"\""), 64)

	return
}

func LoadIpDatabase(database string) (subnetArr []*SubNet, err error) {
	subnetArr = make([]*SubNet, 0)
	f, err := os.Open(database)
	if err != nil {
		return nil,fmt.Errorf("can not load database file %v", database)
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
			return nil,fmt.Errorf("load database file:%v err:%v", database, err)
		}
		var s *SubNet
		if s, err = parseIpDatabase(index, line); err != nil {
			fmt.Println(err)
			continue
		}
		subnetArr = append(subnetArr, s)
	}
	err=nil

	return
}

