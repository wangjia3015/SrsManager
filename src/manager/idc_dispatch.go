package manager

import (
	"sort"
	"sync"
)

const (
	LangFang = 0
	MajuQiao = 1
	HuangCun = 2
	IdcCount = 3
)

type Intranet struct {
	Id        int
	SrcName   string
	UpEdge    [IdcCount][]*SrsServer
	uplock    [IdcCount]sync.RWMutex
	DownEdge  [IdcCount][]*SrsServer
	downlock  [IdcCount]sync.RWMutex
	Orign     [IdcCount][]*SrsServer
	orginlock [IdcCount]sync.RWMutex
}

func NewIntranet(name string, subnet *SubNet) (p *Intranet) {
	p = new(Intranet)
	p.SrcName = name
	p.Id = subnet.Id
	for i := 0; i < IdcCount; i++ {
		p.UpEdge[i] = make([]*SrsServer, 0)
		p.DownEdge[i] = make([]*SrsServer, 0)
		p.Orign[i] = make([]*SrsServer, 0)
	}

	return
}

func (p *Intranet) AddServer(s *SrsServer) {
	servers, lock := p.getDispServers(s.Net.IspType, s.Type)
	lock.Lock()
	servers = append(servers, s)
	lock.Unlock()
	lock.RLock()
	sort.Sort(SortSrsServers(servers))
	lock.RUnlock()
}

func (p *Intranet) sortByLoad() {
	for i := 0; i < IdcCount; i++ {
		p.uplock[i].RLock()
		sort.Sort(SortSrsServers(p.UpEdge[i]))
		p.uplock[i].RUnlock()
		p.downlock[i].RLock()
		sort.Sort(SortSrsServers(p.DownEdge[i]))
		p.downlock[i].RUnlock()
		p.orginlock[i].RLock()
		sort.Sort(SortSrsServers(p.Orign[i]))
		p.orginlock[i].RUnlock()
	}
}

func (p *Intranet) dispatch(i *IpDatabase, count, ispType, disType int) (servers []*SrsServer) {
	servers = make([]*SrsServer, 0)

	return
}

func (p *Intranet) getDispServers(idcType, dispType int) (servers []*SrsServer,
	lock *sync.RWMutex) {
	if dispType == SERVER_TYPE_EDGE_UP {
		servers = p.UpEdge[idcType]
		lock = &p.uplock[idcType]
	} else if dispType == SERVER_TYPE_EDGE_DOWN {
		servers = p.DownEdge[idcType]
		lock = &p.downlock[idcType]
	} else if dispType == SERVER_TYPE_ORIGIN {
		servers = p.Orign[idcType]
		lock = &p.orginlock[idcType]
	}

	return
}
