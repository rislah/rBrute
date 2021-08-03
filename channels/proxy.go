package channels

import (
	"context"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type ProxyStatus int

const (
	AVAILABLE ProxyStatus = iota
	BUSY
	BANNED
)

type Proxy struct {
	addr              string
	totalSuccessCount int32
	mutex             sync.RWMutex
	status            ProxyStatus
}

func (p *Proxy) IncrementTotalSuccessCount() {
	atomic.AddInt32(&p.totalSuccessCount, 1)
}

func (p *Proxy) ChangeStatusToAvailable() {
	p.changeStatus(AVAILABLE)
}

func (p *Proxy) GetAddr() string {
	return p.addr
}

func (p *Proxy) ChangeStatusToBusy() {
	p.changeStatus(BUSY)
}

func (p *Proxy) ChangeStatusToBanned() {
	p.changeStatus(BANNED)
}

func (p *Proxy) changeStatus(status ProxyStatus) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.status = status
}

func (p *Proxy) IsStatus(status ProxyStatus) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	if p.status == status {
		return true
	}
	return false
}

type ProxyFO struct {
	channel         chan *Proxy
	filePath        string
	threadCount     int
	unbanAfterTries int
	proxies         []*Proxy
	mutex           sync.RWMutex
}

func NewProxyFO(ch chan *Proxy, filePath string, threadCount int, unbanAfterTries int) ProxyFO {
	return ProxyFO{channel: ch, filePath: filePath, threadCount: threadCount, unbanAfterTries: unbanAfterTries}
}

func (pfo *ProxyFO) Len() int {
	pfo.mutex.RLock()
	defer pfo.mutex.RUnlock()
	return len(pfo.proxies)
}

func (pfo *ProxyFO) Less(i, j int) bool {
	pfo.mutex.RLock()
	defer pfo.mutex.RUnlock()
	return pfo.proxies[i].totalSuccessCount > pfo.proxies[j].totalSuccessCount
}

func (pfo *ProxyFO) Swap(i, j int) {
	pfo.mutex.Lock()
	defer pfo.mutex.Unlock()
	pfo.proxies[i], pfo.proxies[j] = pfo.proxies[j], pfo.proxies[i]
}

func (pfo *ProxyFO) UpdateWorkingProxy(proxy *Proxy) {
	proxy.IncrementTotalSuccessCount()
	sort.Sort(pfo)
	proxy.ChangeStatusToAvailable()
}

func (pfo *ProxyFO) Produce(ctx context.Context) {
	lines, err := readLines(pfo.filePath)
	if err != nil {
		log.Fatal(err)
	}

	var pr []*Proxy
	for _, line := range lines {
		pr = append(pr, &Proxy{addr: line})
	}
	pfo.proxies = pr

	var counter int
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if counter == pfo.unbanAfterTries {
			pfo.unbanAll()
			counter = 0
		}

		nproxies := pfo.getNProxies(pfo.threadCount)
		if len(nproxies) == 0 {
			time.Sleep(1 * time.Second)
			counter++
			continue
		}
		counter = 0

		for _, nproxy := range nproxies {
			select {
			case pfo.channel <- nproxy:
			}
		}
	}
}

func (pfo *ProxyFO) getNProxies(n int) []*Proxy {
	pfo.mutex.RLock()
	defer pfo.mutex.RUnlock()

	var proxies []*Proxy
	for _, proxy := range pfo.proxies {
		if len(proxies) == n {
			return proxies
		}
		if proxy.IsStatus(AVAILABLE) {
			proxy.ChangeStatusToBusy()
			proxies = append(proxies, proxy)
		}
	}
	return proxies
}

func (pfo *ProxyFO) unbanAll() {
	pfo.mutex.RLock()
	defer pfo.mutex.RUnlock()
	for _, proxy := range pfo.proxies {
		if proxy.IsStatus(BANNED) {
			proxy.ChangeStatusToAvailable()
		}
	}
}
