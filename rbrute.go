package rbrute

import (
	"context"
	"sync"

	"github.com/rislah/rBrute/combolist"
	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
	"github.com/rislah/rBrute/proxy"
)

type RBrute struct {
	ctx            context.Context
	config         *config.Config
	credsStream    <-chan *combolist.Credentials
	proxyStream    <-chan *proxy.Proxy
	proxyGenerator *proxy.Generator
}

func NewRBrute(ctx context.Context, config *config.Config) *RBrute {
	return &RBrute{
		config: config,
		ctx:    ctx,
	}
}

func (rb *RBrute) Start(proxyPath, credsPath string) {
	var wg sync.WaitGroup
	wg.Add(rb.config.Settings.BotCount)

	if rb.config.Settings.UseProxy {
		rb.proxyGenerator = proxy.Start(rb.ctx, proxyPath, rb.config.Settings.BotCount, rb.config.Settings.UnbanProxiesAfter)
		rb.proxyStream = rb.proxyGenerator.Start()
	}
	rb.credsStream = combolist.Start(rb.ctx, rb.config.Settings.BotCount, credsPath)

	logger := logger.NewLogger("/home/rsl/go/src/github.com/rislah/rBrute/results")
	logger.Init(rb.config.Settings.ConfigName)

	for i := 0; i < rb.config.Settings.BotCount; i++ {
		worker := NewWorker(rb.ctx, rb.credsStream, rb.proxyStream, &logger, rb.config.Stages, rb.config.Settings.UseProxy, rb.proxyGenerator)
		go worker.start(&wg)
	}

	wg.Wait()
}
