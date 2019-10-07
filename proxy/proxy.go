package proxy

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

type Status int

const (
	AVAILABLE Status = iota
	BUSY
	BANNED
)

type Proxy struct {
	addr              string
	totalSuccessCount int32
	mutex             sync.RWMutex
	status            Status
}

func Start(ctx context.Context, proxyPath string, threadCount, unbanAfterTries int) *Generator {
	proxies := newProxies(proxyPath)
	generator := NewGenerator(ctx, proxies, threadCount, unbanAfterTries)
	return generator
}

func newProxies(path string) []*Proxy {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	proxies := []*Proxy{}
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		cleaned := strings.TrimSpace(string(line))
		proxies = append(proxies, NewProxy(cleaned))
	}
	return proxies
}

func NewProxy(addr string) *Proxy {
	return &Proxy{addr: addr}
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

func (p *Proxy) changeStatus(status Status) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.status = status
}

func (p *Proxy) IsStatus(status Status) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	if p.status == status {
		return true
	}
	return false
}
