package main

import (
	"context"
	rbrute "github.com/rislah/rBrute"
	"github.com/rislah/rBrute/channels"
	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.NewConfig("/home/risto/code/go/src/github.com/rislah/rBrute/config.yaml")
	start(ctx, cfg)
	cancel()
}

func start(ctx context.Context, cfg *config.Config) {
	credentialsChannel := make(chan *channels.Credentials, cfg.Settings.BotCount)
	cfo := channels.NewCredentialsFO(credentialsChannel, cfg.Settings.CredentialsPath)
	go cfo.Produce(ctx)

	proxyChannel := make(chan *channels.Proxy, cfg.Settings.BotCount)
	pfo := channels.NewProxyFO(proxyChannel, cfg.Settings.ProxyPath, cfg.Settings.BotCount, cfg.Settings.UnbanProxiesAfter)
	if cfg.Settings.UseProxy {
		go pfo.Produce(ctx)
	}

	lg := logger.NewLogger(cfg.Settings.ResultsPath)
	lg.Init(cfg.Settings.ConfigName)

	var wg sync.WaitGroup
	wg.Add(cfg.Settings.BotCount)
	for i := 0; i < cfg.Settings.BotCount; i++ {
		worker := rbrute.NewWorker(ctx, credentialsChannel, proxyChannel, &lg, cfg.Stages, cfg.Settings.UseProxy, &pfo, cfg.Settings.ProxyMaxRetries)
		go worker.Start(&wg)
	}
	wg.Wait()
}
