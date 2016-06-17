package manager

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrEdgeServerNotExsit    = errors.New("Edge server not exsit")
	ErrInvalidEdgeServerAddr = errors.New("inavalid Edge server Addr")
)

type EdgeManager struct {
	db      *DBSync
	Servers map[int64]*Edge
	sync.RWMutex
}

func NewEdgeManager() (em *EdgeManager) {
	em = new(EdgeManager)
	em.Servers = make(map[int64]*Edge, 0)

	return
}

func IpToInt64(ip net.IP) int64 {
	bits := strings.Split(ip.String(), ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])
	var sum int64
	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

func Int64ToIp(ipnr int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)
	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

func (em *EdgeManager) put(e *Edge) {
	em.Lock()
	defer em.Unlock()
	em.Servers[IpToInt64(net.ParseIP(e.Addr))] = e
}

func (em *EdgeManager) getAndPutEdgeServer(addr string) (e *Edge, err error) {
	var (
		ok bool
	)
	if net.ParseIP(addr) == nil {
		return nil, ErrInvalidEdgeServerAddr
	}
	em.RLock()
	e, ok = em.Servers[IpToInt64(net.ParseIP(addr))]
	em.RUnlock()
	if !ok {

	}
	em.put(e)

	return
}
