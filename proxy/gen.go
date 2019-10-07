package proxy

import (
	"context"
	"sort"
	"sync"
	"time"
)

type Generator struct {
	ctx             context.Context
	mutex           sync.RWMutex
	proxies         []*Proxy
	unbanAfterTries int
	threadCount     int
}

func NewGenerator(ctx context.Context, proxies []*Proxy, threadCount, unbanAfterTries int) *Generator {
	return &Generator{
		ctx:             ctx,
		proxies:         proxies,
		threadCount:     threadCount,
		unbanAfterTries: unbanAfterTries,
	}
}

func (g *Generator) Len() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return len(g.proxies)
}

func (g *Generator) Less(i, j int) bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.proxies[i].totalSuccessCount > g.proxies[j].totalSuccessCount
}

func (g *Generator) Swap(i, j int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.proxies[i], g.proxies[j] = g.proxies[j], g.proxies[i]
}

func (g *Generator) Start() <-chan *Proxy {
	stream := make(chan *Proxy, g.threadCount)
	go func() {
		defer close(stream)
		var counter int
		for {
			select {
			case <-g.ctx.Done():
				return
			default:
			}

			if counter == g.unbanAfterTries {
				g.unbanAll()
				counter = 0
			}

			proxies := g.getNProxies(g.threadCount)
			if len(proxies) == 0 {
				time.Sleep(1 * time.Second)
				counter++
				continue
			}

			counter = 0

			for _, proxy := range proxies {
				select {
				case <-g.ctx.Done():
					return
				case stream <- proxy:
				}
			}
		}
	}()
	return stream
}

func (g *Generator) getNProxies(n int) []*Proxy {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	proxies := []*Proxy{}
	for _, proxy := range g.proxies {
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

func (g *Generator) unbanAll() {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	for _, proxy := range g.proxies {
		if proxy.IsStatus(BANNED) {
			proxy.ChangeStatusToAvailable()
		}
	}
}

func (g *Generator) UpdateWorkingProxy(proxy *Proxy) {
	proxy.IncrementTotalSuccessCount()
	sort.Sort(g)
	proxy.ChangeStatusToAvailable()
}
