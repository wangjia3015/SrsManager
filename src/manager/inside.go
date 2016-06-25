package manager

import (
	"sort"
	"sync"
	"sync/atomic"
)

const (
	LangFang = 0
	MajuQiao = 1
	IdcCount = 2
)

type InsideLive struct {
	Orign     [IdcCount][]*SrsServer
	uplock    [IdcCount]sync.RWMutex
	DownEdge  [IdcCount][]*SrsServer
	downlock  [IdcCount]sync.RWMutex
	currIndex int32
}

func NewInsideLive() (p *InsideLive) {
	p = new(InsideLive)
	for i := 0; i < IdcCount; i++ {
		p.Orign[i] = make([]*SrsServer, 0)
		p.DownEdge[i] = make([]*SrsServer, 0)
	}

	return
}

func (p *InsideLive) AddServer(s *SrsServer) (err error) {
	servers, lock := p.getDispServers(s.Idc, s.Type)
	lock.Lock()
	servers = append(servers, s)
	lock.Unlock()
	lock.RLock()
	sort.Sort(SortSrsServers(servers))
	lock.RUnlock()

	return
}

func (p *InsideLive) sortByLoad() {
	for i := 0; i < IdcCount; i++ {
		p.uplock[i].RLock()
		sort.Sort(SortSrsServers(p.Orign[i]))
		p.uplock[i].RUnlock()
		p.downlock[i].RLock()
		sort.Sort(SortSrsServers(p.DownEdge[i]))
		p.downlock[i].RUnlock()
	}
}

func (p *InsideLive) dispatch(count, disType int) (servers []*SrsServer) {
	servers = make([]*SrsServer, 0)
	for i := 0; i < 5; i++ {
		needCount := count / IdcCount
		idcservers, lock := p.getIdcServers(disType)
		lock.RLock()
		start := 0
		for _, s := range idcservers {
			var notExsit bool
			for _, exsitS := range servers {
				if exsitS.Addr == s.Addr {
					notExsit = true
				}
			}
			if !notExsit && s.IsAvaliable() {
				servers = append(servers, s)
				start++
			}
			if len(servers) == count || start >= needCount {
				break
			}
		}
		lock.RUnlock()
		if len(servers) == count {
			break
		}
	}

	return
}

func (p *InsideLive) getIdcServers(dispType int) (servers []*SrsServer,
	lock *sync.RWMutex) {
	var index int32
	if index = atomic.LoadInt32(&p.currIndex); index >= IdcCount {
		atomic.StoreInt32(&p.currIndex, 0)
	}
	if dispType == SERVER_TYPE_ORIGIN {
		servers = p.Orign[int(index)]
		lock = &p.uplock[int(index)]
	} else if dispType == SERVER_TYPE_EDGE_DOWN {
		servers = p.DownEdge[int(index)]
		lock = &p.downlock[int(index)]
	}
	atomic.AddInt32(&p.currIndex, 1)
	if index = atomic.LoadInt32(&p.currIndex); index >= IdcCount {
		atomic.StoreInt32(&p.currIndex, 0)
	}

	return
}

func (p *InsideLive) getDispServers(idc, dispType int) (servers []*SrsServer,
	lock *sync.RWMutex) {
	if dispType == SERVER_TYPE_EDGE_DOWN {
		servers = p.DownEdge[idc]
		lock = &p.downlock[idc]
	} else if dispType == SERVER_TYPE_ORIGIN {
		servers = p.Orign[idc]
		lock = &p.uplock[idc]
	}

	return
}
