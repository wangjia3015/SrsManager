package manager

import (
	"sort"
	"sync"
)

const (
	LangFang = 0
	MajuQiao = 1
	IdcCount = 2
)

type InsideLive struct {
	Id       int
	SrcName  string
	Orign    [IdcCount]*SrsServer
	uplock   [IdcCount]sync.RWMutex
	DownEdge [IdcCount]*SrsServer
	downlock [IdcCount]sync.RWMutex
}

func NewInsideLive(name string, subnet *SubNet) (p *InsideLive) {
	p = new(InsideLive)
	p.SrcName = name
	p.Id = subnet.Id
	for i := 0; i < IdcCount; i++ {
		p.Orign = make([]*SrsServer, 0)
		p.DownEdge = make([]*SrsServer, 0)
	}

	return
}

func (p *InsideLive) AddServer(s *SrsServer) {
	servers, lock := p.getDispServers(s.Type)
	lock.Lock()
	servers = append(servers, s)
	lock.Unlock()
	lock.RLock()
	sort.Sort(SortSrsServers(servers))
	lock.RUnlock()
}

func (p *InsideLive) sortByLoad() {
	for i:=0;i<IdcCount;i++ {
		p.uplock[i].RLock()
		sort.Sort(SortSrsServers(p.Orign[i]))
		p.uplock[i].RUnlock()
		p.downlock[i].RLock()
		sort.Sort(SortSrsServers(p.DownEdge[i]))
		p.downlock[i].RUnlock()
	}
}

func (p *InsideLive) dispatch(i *IpDatabase, count, idc, disType int) (servers []*SrsServer) {
	servers = make([]*SrsServer, 0)

	return
}

func (p *InsideLive) getDispServers(idc,dispType int) (servers []*SrsServer,
	lock *sync.RWMutex) {
	if dispType == SERVER_TYPE_ORIGIN {
		servers = p.Orign[idc]
		lock = &p.uplock
	} else if dispType == SERVER_TYPE_EDGE_DOWN {
		servers = p.DownEdge
		lock = &p.downlock
	}

	return
}
