package proxy_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/rislah/rBrute/proxy"
)

func createProxies(n int) []*Proxy {
	proxies := []*Proxy{}
	for i := 0; i < n; i++ {
		proxies = append(proxies, NewProxy(fmt.Sprintf("192.168.1.%d", i)))
	}
	return proxies
}

func setup(ctx context.Context, proxiesCount, threadCount, unbanAfter int) <-chan *Proxy {
	proxies := createProxies(proxiesCount)
	generator := NewGenerator(ctx, proxies, threadCount, unbanAfter)
	return generator.Start()
}

func TestGeneratorGetProxy(t *testing.T) {
	stream := setup(context.Background(), 2, 1, 1)
	select {
	case p := <-stream:
		if p == nil {
			t.Error("got nil from stream")
		}
		return
	case <-time.After(2 * time.Second):
	}
}

func TestGeneratorCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	stream := setup(ctx, 5, 1, 1)
	cancel()
	select {
	case _, ok := <-stream:
		if ok {
			t.Error("stream was left open")
		}
	case <-time.After(2 * time.Second):
	}
}
