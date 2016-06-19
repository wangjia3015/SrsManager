package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"manager"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CT     = 0
	CNC    = 1
	CMCC   = 2
	Unknow = 3
)

type IpDatabase struct {
	SubNets   map[string]*SubNet
	Provinces map[string]*Province
}

func NewIpDatabase() (i *IpDatabase, err error) {
	i = &IpDatabase{SubNets: make(map[string]*SubNet), Provinces: make(map[string]*Province)}
	if err = i.LoadIpDatabase("./isp.txt"); err != nil {
		return
	}
	i.initProvince()
	return
}

/*
  dispatch algorithm
*/
func (i *IpDatabase) DisPatch(addr string, disType, count int) (servers []*manager.SrsServer) {
	net, err := i.GetSubNet(addr)
	if err != nil {
		net = &SubNet{IspType: CT, Province: "beijing"}
	}
	p, ok := i.Provinces[net.Province]
	if !ok {
		p = i.Provinces["beijing"]
	}
	needIspType := net.IspType
	if needIspType != CT && needIspType != CMCC && needIspType != CNC {
		needIspType = CT
	}
	return p.dispatch(i, count, net.IspType, disType)
}

func (i *IpDatabase) AddServer(s *manager.SrsServer) {
	net, err := i.GetSubNet(s.Host)
	if err != nil {
		net = &SubNet{IspType: CT, Province: "beijing"}
	}
	p, ok := i.Provinces[net.Province]
	if !ok {
		p = i.Provinces["beijing"]
	}
	s.Net = p.subnet
	p.AddServer(s)
}

type Province struct {
	OriginName string
	Distances  []*DistanceProvince
	subnet     *SubNet
	UpEdge     map[int][]*manager.SrsServer
	uplock     [Unknow]sync.RWMutex
	DownEdge   map[int][]*manager.SrsServer
	downlock   [Unknow]sync.RWMutex
	Orign      map[int][]*manager.SrsServer
	orginlock  [Unknow]sync.RWMutex
}

func NewProvince(name string, subnet *SubNet) (p *Province) {
	p = new(Province)
	p.OriginName = name
	p.Distances = make([]*DistanceProvince, 0)
	p.subnet = subnet
	p.UpEdge = make(map[int][]*manager.SrsServer)
	p.DownEdge = make(map[int][]*manager.SrsServer)
	p.Orign = make(map[int][]*manager.SrsServer)
	for i := CT; i <= Unknow; i++ {
		p.UpEdge[i] = make([]*manager.SrsServer, 0)
		p.DownEdge[i] = make([]*manager.SrsServer, 0)
		p.Orign = make([]*manager.SrsServer, 0)
	}

	return
}

func (p *Province) AddServer(s *manager.SrsServer) {
	servers, lock := p.getDispServers(s.Net.IspType, s.Type)
	lock.Lock()
	servers = append(servers, s)
	lock.Unlock()
	lock.RLock()
	sort.Sort(manager.SortSrsServers(servers))
	lock.RUnlock()
}

func (p *Province) sortByLoad() {
	for i := CT; i <= CMCC; i++ {
		p.uplock[i].RLock()
		sort.Sort(manager.SortSrsServers(p.UpEdge[i]))
		p.uplock[i].RUnlock()
		p.downlock[i].RLock()
		sort.Sort(manager.SortSrsServers(p.DownEdge[i]))
		p.downlock[i].RUnlock()
		p.orginlock[i].RLock()
		sort.Sort(manager.SortSrsServers(p.Orign[i]))
		p.orginlock[i].RUnlock()
	}
}

func (p *Province) dispatch(i *IpDatabase, count, ispType, disType int) (servers []*manager.SrsServer) {
	servers = make([]*manager.SrsServer, 0)
	execult := make([]*manager.SrsServer, 0)
	for _, d := range p.Distances {
		destname := d.DestName
		dp, _ := i.Provinces[destname]
		dispServers, lock := dp.getDispServers(ispType, disType)
		lock.RLock()
		for _, e := range dispServers {
			isExsit := false
			for _, es := range execult {
				if es.Host == e.Host {
					isExsit = true
					break
				}
			}
			if !isExsit {
				servers = append(servers, e)
				execult = append(execult, e)
			}
			if len(servers) == count {
				lock.RUnlock()
				return
			}
		}
		lock.RUnlock()
	}

	return
}

func (p *Province) getDispServers(needIspType, dispType int) (servers []*manager.SrsServer,
	lock *sync.RWMutex) {
	if dispType == manager.SERVER_TYPE_EDGE_UP {
		servers, _ = p.UpEdge[needIspType]
		lock = &p.uplock[needIspType]
	} else if dispType == manager.SERVER_TYPE_EDGE_DOWN {
		servers, _ = p.DownEdge[needIspType]
		lock = &p.downlock[needIspType]
	} else if dispType == manager.SERVER_TYPE_ORIGIN {
		servers, _ = p.Orign[needIspType]
		lock = &p.orginlock[needIspType]
	}

	return
}

type DistanceProvince struct {
	DestName string
	Distance float64
}

type SortDistanceProvince []*DistanceProvince

func (sp SortDistanceProvince) Len() int {
	return len(sp)
}

func (sp SortDistanceProvince) Swap(i, j int) {
	sp[i], sp[j] = sp[j], sp[i]
}

func (sp SortDistanceProvince) Less(i, j int) bool {
	return sp[i].Distance < sp[j].Distance
}

func EarthDistance(lat1, lng1, lat2, lng2 float64) float64 {
	var radius float64 = 6371000 // 6378137
	rad := math.Pi / 180.0

	lat1 = lat1 * rad
	lng1 = lng1 * rad
	lat2 = lat2 * rad
	lng2 = lng2 * rad

	theta := lng2 - lng1
	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))
	return dist * radius
}

type SubNet struct {
	Id        int
	Ispname   string
	SupperIsp string
	IspType   int
	Province  string
	Latitude  float64
	Longitude float64
	Net       *net.IPNet
	Desc      string
	IsCapital bool
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
	s.SupperIsp = recordArr[2]
	s.Ispname = recordArr[3]
	switch s.SupperIsp {
	case "cnc":
		s.IspType = CNC
	case "cmcc":
		s.IspType = CMCC
	case "ct":
		s.IspType = CT
	default:
		s.IspType = CT
	}
	arr := strings.Split(s.Ispname, "_")
	if len(arr) == 2 {
		s.Province = arr[0]
	}
	if recordArr[0] == "E" {
		s.IsCapital = true
	}
	s.Latitude, _ = strconv.ParseFloat(recordArr[4], 64)
	s.Longitude, _ = strconv.ParseFloat(recordArr[5], 64)
	s.Desc = strings.Replace(recordArr[6], "\n", "", 100)

	return
}

func (i *IpDatabase) GetSubNet(addr string) (subnet *SubNet, err error) {
	var ok bool
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, fmt.Errorf("unavali ip:%v", addr)
	}
	mask := ip.DefaultMask()
	network := ip.Mask(mask)
	net := net.IPNet{IP: network, Mask: mask}
	if subnet, ok = i.SubNets[net.String()]; !ok {
		return nil, fmt.Errorf("addr :%v subnet:%v not exsits ipdatabase:%v", addr, net.String())
	}

	return
}

func (i *IpDatabase) LoadIpDatabase(database string) (err error) {
	i.SubNets = make(map[string]*SubNet)
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
			continue
		}
		i.SubNets[s.Net.String()] = s
	}
	err = nil

	return
}

func (i *IpDatabase) initProvince() {
	for _, s := range i.SubNets {
		if s.IsCapital {
			i.Provinces[s.Province] = NewProvince(s.Province, s)
		}
	}
	for _, p := range i.Provinces {
		srcNet := p.subnet
		for name, pp := range i.Provinces {
			destNet := pp.subnet
			d := &DistanceProvince{DestName: name, Distance: EarthDistance(srcNet.Latitude,
				srcNet.Longitude, destNet.Latitude, destNet.Longitude)}
			p.Distances = append(p.Distances, d)
			if math.IsNaN(d.Distance) {
				d.Distance = 0
			}
		}
	}
	go i.sort()
}

func (i *IpDatabase) sort() {
	for _, p := range i.Provinces {
		sort.Sort((SortDistanceProvince)(p.Distances))
		time.Sleep(time.Minute)
	}
}

func main() {
	i, err := NewIpDatabase()
	if err != nil {
		fmt.Println(err)
		return
	}
	data, err := json.Marshal(i.Provinces)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(data))
}
